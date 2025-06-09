package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"social-network/internal/config"
	"social-network/internal/middlewars"
	"social-network/internal/models"
	"social-network/internal/repositories"
)

func GetRecentChats(w http.ResponseWriter, r *http.Request) {
	userID := middlewars.GetUserIDFromSession(w, r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := config.GetDB()
	query := `
		SELECT u.id, u.nickname, u.first_name, u.last_name, MAX(m.sent_at) as last_message_time
		FROM messages m
		JOIN users u ON (m.sender_id = u.id OR m.receiver_id = u.id) AND u.id != ?
		WHERE m.sender_id = ? OR m.receiver_id = ?
		GROUP BY u.id
		ORDER BY last_message_time DESC`

	rows, err := db.Query(query, userID, userID, userID)
	if err != nil {
		log.Println("❌ Error retrieving recent chats:", err)
		http.Error(w, "Failed to retrieve recent chats", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []any
	for rows.Next() {
		var user models.User
		var tm string
		err := rows.Scan(&user.ID, &user.Nickname, &user.FirstName, &user.LastName, &tm)
		if err != nil {
			log.Println("❌ Error scanning recent chat user:", err)
			continue
		}
		u := map[string]any{
			"id":                user.ID,
			"nickname":          user.Nickname,
			"first_name":        user.FirstName,
			"last_name":         user.LastName,
			"last_message_time": tm,
		}
		users = append(users, u)
	}

	json.NewEncoder(w).Encode(users)
}

func GetChatHistoryHandler(w http.ResponseWriter, r *http.Request) {
	userID := middlewars.GetUserIDFromSession(w, r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	receiverIDStr := r.URL.Query().Get("receiver_id")
	log.Printf("📌 Extracted receiverID from request: %s", receiverIDStr)

	receiverID, err := strconv.Atoi(receiverIDStr)
	if err != nil || receiverID == 0 {
		log.Println("❌ Invalid user ID:", receiverIDStr)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	db := config.GetDB()
	repo := repositories.NewChatRepository(db) // ✅ FIXED function name

	messages, err := repo.GetMessages(userID, receiverID) // ✅ FIXED function name
	if err != nil {
		log.Println("❌ Error retrieving chat history:", err)
		http.Error(w, "Failed to retrieve chat history", http.StatusInternalServerError)
		return
	}

	log.Printf("📜 Chat history retrieved between %d and %d", userID, receiverID)
	json.NewEncoder(w).Encode(messages)
}

func GetAvailableChatUsers(w http.ResponseWriter, r *http.Request) {
	userID := middlewars.GetUserIDFromSession(w, r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := config.GetDB()
	query := `
		SELECT DISTINCT u.id, u.nickname, u.first_name, u.last_name
		FROM users u
		LEFT JOIN followers f1 ON f1.follower_id = ? AND f1.following_id = u.id AND f1.status = 'accepted'
		LEFT JOIN followers f2 ON f2.following_id = ? AND f2.follower_id = u.id AND f2.status = 'accepted'
		WHERE u.id != ? 
		  AND (u.is_private = 0 OR f1.follower_id IS NOT NULL OR f2.following_id IS NOT NULL)
		  AND u.id NOT IN (
			-- Exclude users already in recent chats
			SELECT DISTINCT CASE 
				WHEN m.sender_id = ? THEN m.receiver_id 
				WHEN m.receiver_id = ? THEN m.sender_id 
			END
			FROM messages m
			WHERE m.sender_id = ? OR m.receiver_id = ?
		)
		ORDER BY u.nickname ASC`

	rows, err := db.Query(query, userID, userID, userID, userID, userID, userID, userID)
	if err != nil {
		log.Println("❌ Error retrieving available chat users:", err)
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Nickname, &user.FirstName, &user.LastName)
		if err != nil {
			log.Println("❌ Error scanning available chat user:", err)
			continue
		}
		users = append(users, user)
	}

	json.NewEncoder(w).Encode(users)
}
