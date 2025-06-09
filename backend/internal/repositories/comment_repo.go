package repositories

import (
	"database/sql"

	"social-network/internal/models"
)

type CommentRepository struct {
	DB *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{DB: db}
}

func (repo *CommentRepository) GetCommentsForPost(postID int) ([]models.Comment, error) {
	var comments []models.Comment

	rows, err := repo.DB.Query(`
        SELECT *
        FROM comments
        WHERE post_id = ?
        ORDER BY id ASC`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Nickname, &comment.Content, &comment.Image)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

func (repo *CommentRepository) AddComment(comment *models.Comment) error {
	_, err := repo.DB.Exec(`
        INSERT INTO comments (post_id, user_id, content, username) 
        VALUES (?, ?, ?, ?)`,
		comment.PostID, comment.UserID, comment.Content, comment.Nickname,
	)
	return err
}

func (repo *CommentRepository) GetLastInsertedComment(userID, postID int) (*models.Comment, error) {
	var comment models.Comment
	err := repo.DB.QueryRow(`
		SELECT id, post_id, user_id, content, image, username 
		FROM comments 
		WHERE user_id = ? AND post_id = ? 
		ORDER BY id DESC LIMIT 1`, userID, postID).
		Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.Image, &comment.Nickname)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}
