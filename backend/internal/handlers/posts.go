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
	"strconv"
	"time"

	"social-network/internal/config"
	"social-network/internal/middlewars"
	"social-network/internal/models"
	"social-network/internal/repositories"
	"social-network/internal/websocket"
)

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	fmt.Println("Userid from posthan:", user.ID)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var post models.Post
	db := config.GetDB()
	repo := repositories.NewPostRepository(db)

	err := r.ParseMultipartForm(10 << 20) // 10MB max file size
	if err != nil {
		log.Println("error in parsing multi part form", err)
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")
	privacy := r.FormValue("privacy")
	var imagePath string = ""
	imageFile, header, err := r.FormFile("image")
	if err == nil { // If an image was uploaded
		defer imageFile.Close()
		mimeType := header.Header.Get("Content-Type")
		validMimeTypes := []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
		isValidMimeType := false
		for _, validType := range validMimeTypes {
			if mimeType == validType {
				isValidMimeType = true
				break
			}
		}
		if !isValidMimeType {
			http.Error(w, "Invalid image type. Only JPEG, PNG, GIF, and WEBP allowed.", http.StatusBadRequest)
			return
		}

		if _, err := os.Stat("uploads"); os.IsNotExist(err) {
			os.Mkdir("uploads", 0o755)
		}

		fileName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), user.ID, filepath.Ext(header.Filename))
		imagePath = fmt.Sprintf("uploads/%s", fileName)

		out, err := os.Create(imagePath)
		if err != nil {
			http.Error(w, "Error saving image", http.StatusInternalServerError)
			return
		}
		defer out.Close()
		io.Copy(out, imageFile)
	}

	post = models.Post{
		UserID:   user.ID,
		Nickname: user.Nickname,
		Content:  content,
		Privacy:  privacy,
		Image:    &imagePath, // Nullable
	}

	newPost, err := repo.CreatePost(&post)
	if err != nil {
		log.Println("Error creating post:", err)
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(newPost); err != nil {
		log.Println("Error encoding JSON response:", err)
		http.Error(w, "Failed to return post", http.StatusInternalServerError)
		return
	}
	fmt.Println("NEW POST", post)
}

func GetAllPostsHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	db := config.GetDB()
	repo := repositories.NewPostRepository(db)

	posts, err := repo.GetFeedPosts(user.ID)
	if err != nil {
		log.Println("❌ Error retrieving user feed posts:", err)
		http.Error(w, "Failed to retrieve posts", http.StatusInternalServerError)
		return
	}

	// w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(posts)
}

func LikePost(w http.ResponseWriter, r *http.Request) {
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
	err := db.QueryRow("SELECT is_like FROM likes WHERE post_id = ? AND user_id = ?", like.Postid, user.ID).Scan(&existingLike)

	if err != nil && err != sql.ErrNoRows {
		log.Println("❌ Error retrieving like:", err)
		http.Error(w, "Error checking like status", http.StatusInternalServerError)
		return
	}

	// **Fetch Post Creator ID**
	var postCreator int
	err = db.QueryRow("SELECT user_id FROM posts WHERE id = ?", like.Postid).Scan(&postCreator)
	if err != nil {
		log.Println("❌ Error retrieving creator ID:", err)
		http.Error(w, "Cannot retrieve post creator ID", http.StatusInternalServerError)
		return
	}

	notificationMsg := "liked your post."

	// **✅ If like does NOT exist, INSERT it**
	// if err == sql.ErrNoRows {
	if !existingLike.Valid {
		log.Println("🟢 Inserting new like into database...")
		_, err = db.Exec("INSERT INTO likes (post_id, user_id, is_like) VALUES (?, ?, ?)", like.Postid, user.ID, like.IsLike)
		if err != nil {
			log.Println("❌ Error inserting like:", err)
			http.Error(w, "Error adding like", http.StatusInternalServerError)
			return
		}

		// **Send Notification**
		if like.IsLike {
			websocket.SendNotification(postCreator, "like", user.Nickname+" "+notificationMsg)
		} else {
			websocket.SendNotification(postCreator, "dislike", user.Nickname+" disliked your post.")
		}
	} else {
		// **If like exists, update or remove it**
		if existingLike.Valid && existingLike.Bool == like.IsLike {
			log.Println("🟡 Removing existing like from database...")
			_, err = db.Exec("DELETE FROM likes WHERE post_id = ? AND user_id = ?", like.Postid, user.ID)
		} else {
			log.Printf("🔍 Existing Like Check: err=%v, existingLike.Valid=%v, existingLike.Bool=%v", err, existingLike.Valid, existingLike.Bool)
			log.Println("🟠 Updating like in database...")
			_, err = db.Exec("UPDATE likes SET is_like = ? WHERE post_id = ? AND user_id = ?", like.IsLike, like.Postid, user.ID)
		}

		if err != nil {
			log.Println("❌ Error updating/removing like:", err)
			http.Error(w, "Error updating like/dislike", http.StatusInternalServerError)
			return
		}
	}
	likeCount, _ := repositories.GetLikeCount(like.Postid)
	dislikeCount, _ := repositories.GetDislikeCount(like.Postid)

	websocket.BroadcastPostUpdate(like.Postid, likeCount, dislikeCount)

	response := map[string]any{
		"message":     "Action successful",
		"post_id":     like.Postid,
		"likes":       likeCount,
		"dislikes":    dislikeCount,
		"user_action": like.IsLike,
	}
	json.NewEncoder(w).Encode(response)
}

func GetUserPostsHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	viewingUserID, err := strconv.Atoi(r.URL.Query().Get("user_id"))
	if err != nil || viewingUserID == 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	db := config.GetDB()
	repo := repositories.NewPostRepository(db)
	posts, err := repo.GetUserPosts(viewingUserID, user.ID)
	if err != nil {
		log.Println("Error retrieving posts:", err)
		http.Error(w, "Failed to retrieve posts", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(posts)
}
