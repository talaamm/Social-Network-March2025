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
	"social-network/internal/websocket"
)

func GetGroupPostCommentsHandler(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(r.URL.Query().Get("post_id"))
	if err != nil || postID == 0 {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	db := config.GetDB()

	query := `
        SELECT id, member_id, g_post_id, content, image , username
        FROM group_comments WHERE g_post_id = ? ORDER BY id ASC`

	rows, err := db.Query(query, postID)
	if err != nil {
		log.Println("Error fetching comments:", err)
		http.Error(w, "Error fetching comments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []models.GroupComment
	for rows.Next() {
		var comment models.GroupComment
		if err := rows.Scan(&comment.ID, &comment.MemberID, &comment.GPostID, &comment.Content, &comment.Image, &comment.Nickname); err != nil {
			log.Println("Error scanning comment:", err)
			continue
		}
		comments = append(comments, comment)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(comments)
}

func AddGroupPostCommentHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var comment models.GroupComment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		log.Println("error decoding body for group comment", err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	var post_creator int
	db := config.GetDB()
	err := db.QueryRow(`
	SELECT group_id, member_id
	FROM group_posts 
	WHERE id = ? `, comment.GPostID).
		Scan(&comment.GroupID, &post_creator)
	if err != nil {
		log.Println("Error getting groupid for comment:", err)
		http.Error(w, "Error getting groupid for comment", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO group_comments (member_id, g_post_id, content, image , username, group_id) VALUES (?, ?, ?, ?, ?, ?)",
		comment.MemberID, comment.GPostID, comment.Content, comment.Image, user.Nickname, comment.GroupID)
	if err != nil {
		log.Println("Error inserting comment:", err)
		http.Error(w, "Error adding comment", http.StatusInternalServerError)
		return
	}

	newComment, err := LastInsertedGroupComment(comment.MemberID, comment.GPostID)
	if err != nil {
		log.Println("Error retrieving new comment:", err)
		http.Error(w, "Failed to retrieve comment", http.StatusInternalServerError)
		return
	}
	var GroupName string
	db.QueryRow(`
	SELECT name
	FROM groups 
	WHERE id = ? `, comment.GroupID).
		Scan(&GroupName)

	websocket.SendNotification(post_creator, "comment", user.Nickname+" commented on your post in Group: "+GroupName)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newComment)
}

func LastInsertedGroupComment(memberID, postID int) (*models.GroupComment, error) {
	var comment models.GroupComment

	db := config.GetDB()
	err := db.QueryRow(`
		SELECT id, member_id, g_post_id, content, image , username
		FROM group_comments 
		WHERE member_id = ? AND g_post_id = ? 
		ORDER BY id DESC LIMIT 1`, memberID, postID).
		Scan(&comment.ID, &comment.MemberID, &comment.GPostID, &comment.Content, &comment.Image, &comment.Nickname)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No comment found
		}
		return nil, err
	}
	return &comment, nil
}
