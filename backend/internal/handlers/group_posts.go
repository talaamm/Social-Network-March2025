package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"social-network/internal/config"
	"social-network/internal/middlewars"
	"social-network/internal/models"
	"social-network/internal/repositories"
	"social-network/internal/websocket"
)

func CreateGroupPostHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var post models.GroupPost
	// if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
	// 	log.Println("error decoding body", err)
	// 	http.Error(w, "Invalid input", http.StatusBadRequest)
	// 	return
	// }

	// ✅ Ensure request is multipart/form-data (since we're using FormData in Vue)
	err := r.ParseMultipartForm(10 << 20) // 10MB max file size
	if err != nil {
		log.Println("error in parsing multi part form", err)
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	post.Content = r.FormValue("content")
	post.GroupID, err = strconv.Atoi(r.FormValue("group_id"))
	if err != nil {
		log.Println("error parsing group id to int")
	}
	post.MemberID = user.ID
	post.Nickname = user.Nickname
	if post.Content == "" || post.GroupID == 0 {
		log.Println("post cannot be empty", post.GroupID, post.Content)
		http.Error(w, "Group ID and content are required", http.StatusBadRequest)
		return
	}

	var imagePath string = ""
	imageFile, header, err := r.FormFile("image")
	if err == nil { // If an image was uploaded
		defer imageFile.Close()
		// ✅ Validate MIME Type
		mimeType := header.Header.Get("Content-Type")
		validMimeTypes := []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
		// isValidMimeType := false
		// for _, validType := range validMimeTypes {
		// 	if mimeType == validType {
		// 		isValidMimeType = true
		// 		break
		// 	}
		// }
		if !slices.Contains(validMimeTypes, mimeType) {
			log.Println("invalid file type:", mimeType)
			http.Error(w, "Invalid image type. Only JPEG, PNG, GIF, and WEBP allowed.", http.StatusBadRequest)
			return
		}
		folderpath := "group_uploads" // fmt.Sprintf("group_%d_uploads", post.GroupID) // Better folder name
		// ✅ Ensure `uploads` folder exists
		if _, err := os.Stat(folderpath); os.IsNotExist(err) {
			// os.Mkdir(folderpath, 0755)
			err = os.MkdirAll(folderpath, 0o755)
			if err != nil {
				log.Println("Error creating directory:", err)
				http.Error(w, "Error creating upload directory", http.StatusInternalServerError)
				return
			}
		}

		// ✅ Generate unique file name
		fileName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), user.ID, filepath.Ext(header.Filename))
		// imagePath = fmt.Sprintf("/"+folderpath+"/%s", fileName)
		imagePath = fmt.Sprintf("%s/%s", folderpath, fileName)

		// ✅ Save Image
		out, err := os.Create(imagePath)
		if err != nil {
			log.Println("error saving image", err)
			http.Error(w, "Error saving image", http.StatusInternalServerError)
			return
		}
		defer out.Close()
		io.Copy(out, imageFile)
		fmt.Println("📂 Saving file to:", imagePath)
		fmt.Println("📸 MIME Type:", mimeType)

	}
	// ✅ Create Post Object
	post.Image = &imagePath
	log.Println("received load:", post)
	db := config.GetDB()
	repo := repositories.NewGroupPostRepository(db)

	newPost, err := repo.CreateGroupPost(&post)
	if err != nil {
		log.Println("Error creating group post:", err)
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}
	log.Println("newPost:", newPost)

	// w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")

	// json.NewEncoder(w).Encode(map[string]string{"message": "Group post created successfully"})
	if err := json.NewEncoder(w).Encode(newPost); err != nil {
		log.Println("Error encoding JSON response:", err)
		http.Error(w, "Failed to return post", http.StatusInternalServerError)
		return
	}
	websocket.BroadcastGroupPostUpdate(newPost.GroupID, newPost.MemberID, newPost.ID, newPost.Nickname, newPost.Content, newPost.CreatedAt)
}

func GetGroupPostsHandler(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.Atoi(r.URL.Query().Get("group_id"))
	if err != nil || groupID == 0 {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	db := config.GetDB()

	query := `
        SELECT id, group_id, member_id, content, image, created_at , username
        FROM group_posts WHERE group_id = ? ORDER BY created_at DESC`

	rows, err := db.Query(query, groupID)
	if err != nil {
		log.Println("Error fetching group posts:", err)
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []models.GroupPost
	for rows.Next() {
		var post models.GroupPost
		if err := rows.Scan(&post.ID, &post.GroupID, &post.MemberID, &post.Content, &post.Image, &post.CreatedAt, &post.Nickname); err != nil {
			log.Println("Error scanning post:", err)
			continue
		}
		// Get like & dislike counts
		post.LikeCount, err = GetGroupLikeCount(post.ID, true)
		if err != nil {
			return
		}
		post.DisLikeCount, err = GetGroupLikeCount(post.ID, false)
		if err != nil {
			return
		}
		posts = append(posts, post)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(posts)
}

func LikeGroupPostHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := config.GetDB()

	var like models.Like
	if err := json.NewDecoder(r.Body).Decode(&like); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var existingLike sql.NullBool
	err := db.QueryRow("SELECT is_like FROM group_likes WHERE post_id = ? AND member_id = ?", like.Postid, user.ID).Scan(&existingLike)

	if err != nil && err != sql.ErrNoRows {
		log.Println("❌ Error retrieving like:", err)
		http.Error(w, "Error checking like status", http.StatusInternalServerError)
		return
	}

	if !existingLike.Valid {
		_, err = db.Exec("INSERT INTO group_likes (post_id, member_id, is_like) VALUES (?, ?, ?)", like.Postid, user.ID, like.IsLike)
	} else {
		// _, err = db.Exec("UPDATE group_likes SET is_like = ? WHERE post_id = ? AND member_id = ?", like.IsLike, like.Postid, user.ID)
		if existingLike.Valid && existingLike.Bool == like.IsLike {
			log.Println("🟡 Removing existing like from database...")
			_, err = db.Exec("DELETE FROM group_likes WHERE post_id = ? AND member_id = ?", like.Postid, user.ID)
		} else {
			// log.Printf("🔍 Existing Like Check: err=%v, existingLike.Valid=%v, existingLike.Bool=%v", err, existingLike.Valid, existingLike.Bool)
			log.Println("🟠 Updating like in database...")
			_, err = db.Exec("UPDATE group_likes SET is_like = ? WHERE post_id = ? AND member_id = ?", like.IsLike, like.Postid, user.ID)
		}
	}

	if err != nil {
		log.Println("❌ Error updating/removing like:", err)
		http.Error(w, "Error updating like/dislike", http.StatusInternalServerError)
		return
	}
	log.Println("post liked:", like.IsLike, "in group for post:", like.Postid, "by memberid:", user.ID)

	likeCount, _ := GetGroupLikeCount(like.Postid, true)
	dislikeCount, _ := GetGroupLikeCount(like.Postid, false)

	response := map[string]any{
		"message":     "Action successful",
		"post_id":     like.Postid,
		"likes":       likeCount,
		"dislikes":    dislikeCount,
		"user_action": like.IsLike,
	}
	json.NewEncoder(w).Encode(response)
}

func GetGroupLikeCount(postID int, islike bool) (int, error) {
	db := config.GetDB()
	var likeCount int
	err := db.QueryRow("SELECT COUNT(*) FROM group_likes WHERE post_id = ? AND is_like = ?", postID, islike).Scan(&likeCount)
	if err != nil {
		log.Println("error counting likes", err)
		return 0, err
	}
	return likeCount, nil
}
