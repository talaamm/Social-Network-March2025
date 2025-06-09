package repositories

import (
	"database/sql"
	"log"
	"time"

	"social-network/internal/config"
	"social-network/internal/models"
)

type PostRepository struct {
	DB *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{DB: db}
}

func (r *PostRepository) CreatePost(post *models.Post) (*models.Post, error) {
	query := `INSERT INTO posts (username, user_id, content , created_at, privacy , image) VALUES (?, ?, ?, ?, ?, ?) 
	RETURNING id, username, user_id, content , created_at, privacy , image`
	var newPost models.Post
	err := r.DB.QueryRow(query, post.Nickname, post.UserID, post.Content, time.Now(), post.Privacy, post.Image).
		Scan(&newPost.ID, &newPost.Nickname, &newPost.UserID, &newPost.Content, &newPost.CreatedAt, &newPost.Privacy, &newPost.Image)
	if err != nil {
		return nil, err
	}
	return &newPost, nil
}

func GetLikeCount(postID int) (int, error) {
	db := config.GetDB()
	var likeCount int
	err := db.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = ? AND is_like = true", postID).Scan(&likeCount)
	if err != nil {
		log.Println("error counting likes", err)
		return 0, err
	}
	return likeCount, nil
}

func GetDislikeCount(postID int) (int, error) {
	db := config.GetDB()
	var dislikeCount int
	err := db.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = ? AND is_like = false", postID).Scan(&dislikeCount)
	if err != nil {
		log.Println("error counting dislikes", err)
		return 0, err
	}
	return dislikeCount, nil
}

func (repo *PostRepository) GetFeedPosts(userID int) ([]models.Post, error) {
	var posts []models.Post
	/*✅ This will now:
	  ✔ Show public posts from non-private users
	  ✔ Show public posts from private users only if followed
	  ✔ Show follower-only posts to accepted followers
	  ✔ Show selected posts only to the users selected by the creator
	  ✔ Show the creator's own posts in their feed*/
	query := `
		SELECT p.id, p.user_id, p.content, p.image, p.username, p.privacy, p.created_at 
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN followers f ON f.follower_id = ? AND f.following_id = p.user_id AND f.status = 'accepted'
		LEFT JOIN posts_visibility pv ON pv.post_creator = p.user_id AND pv.user_id = ?
		WHERE 
			(p.privacy = 'public' AND u.is_private = 0)
			OR (p.privacy = 'public' AND u.is_private = 1 AND f.follower_id IS NOT NULL)
			OR (p.privacy = 'followers' AND f.follower_id IS NOT NULL)
			OR (p.privacy = 'selected' AND pv.user_id IS NOT NULL)
			OR (p.user_id = ?)
		ORDER BY p.created_at DESC`

	rows, err := repo.DB.Query(query, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.ID, &post.UserID, &post.Content, &post.Image, &post.Nickname, &post.Privacy, &post.CreatedAt)
		if err != nil {
			return nil, err
		}

		likeCount, err := GetLikeCount(post.ID)
		if err != nil {
			return nil, err
		}
		dis, err := GetDislikeCount(post.ID)
		if err != nil {
			return nil, err
		}
		post.LikeCount = likeCount
		post.DisLikeCount = dis

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (repo *PostRepository) GetUserPosts(userID int, viewerID int) ([]models.Post, error) {
	// rows, err := repo.DB.Query(`
	// 	SELECT *
	// 	FROM posts 
	// 	WHERE user_id = ? 
	// 	AND (privacy = 'public' OR 
	// 	    (privacy = 'followers' AND user_id IN 
	// 	        (SELECT following_id FROM followers WHERE follower_id = ? AND status = 'accepted')) 
	// 	    OR user_id = ?
	// 		OR ) 
	// 	ORDER BY created_at DESC`, userID, viewerID, viewerID)
	rows, err := repo.DB.Query(`
    SELECT * 
    FROM posts p
    WHERE 
        p.user_id = ? 
        AND ( 
		(p.privacy = 'public') 
        OR (p.privacy = 'followers' AND p.user_id IN 
            (SELECT following_id FROM followers WHERE follower_id = ? AND status = 'accepted'))
        OR (p.privacy = 'selected' AND EXISTS 
            (SELECT 1 FROM posts_visibility pv WHERE post_creator = ? AND pv.user_id = ?)) 
		OR (user_id = ? )
		)
    ORDER BY p.created_at DESC`, userID, viewerID, userID, viewerID , viewerID)
	if err != nil {
		log.Println("Error fetching posts:", err)
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.ID, &post.UserID, &post.Content, &post.Image, &post.Nickname, &post.Privacy, &post.CreatedAt)
		if err != nil {
			log.Println("Error scanning post:", err)
			return nil, err
		}
		likeCount, err := GetLikeCount(post.ID)
		if err != nil {
			return nil, err
		}
		dis, err := GetDislikeCount(post.ID)
		if err != nil {
			return nil, err
		}
		post.LikeCount = likeCount
		post.DisLikeCount = dis

		posts = append(posts, post)
	}

	return posts, nil
}
