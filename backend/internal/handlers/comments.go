package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"social-network/internal/config"
	"social-network/internal/middlewars"
	"social-network/internal/models"
	"social-network/internal/repositories"
	"social-network/internal/websocket"
)

func GetCommentsForPostHandler(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(r.URL.Query().Get("post_id"))
	if err != nil || postID == 0 {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	db := config.GetDB()
	repo := repositories.NewCommentRepository(db)

	comments, err := repo.GetCommentsForPost(postID)
	if err != nil {
		log.Println("Error retrieving comments:", err)
		http.Error(w, "Failed to retrieve comments", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(comments)
}

func CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var comment models.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	comment.UserID = user.ID
	comment.Nickname = user.Nickname

	db := config.GetDB()
	commentRepo := repositories.NewCommentRepository(db)

	err := commentRepo.AddComment(&comment)
	if err != nil {
		log.Println("Error adding comment:", err)
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	newComment, err := commentRepo.GetLastInsertedComment(comment.UserID, comment.PostID)
	if err != nil {
		log.Println("Error retrieving new comment:", err)
		http.Error(w, "Failed to retrieve comment", http.StatusInternalServerError)
		return
	}

	var postCreatorID int
	err = db.QueryRow("SELECT user_id FROM posts WHERE id = ?", comment.PostID).Scan(&postCreatorID)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "cannot retreive post creator id", http.StatusInternalServerError)
		return
	}
	notificationMsg := user.Nickname + " commented on your post."

	websocket.SendNotification(postCreatorID, "comment", notificationMsg)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newComment)
}
