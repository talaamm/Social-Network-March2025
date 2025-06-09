package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"social-network/internal/config"
	"social-network/internal/middlewars"
)

func GetGroupChatHistoryHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupIDStr := r.URL.Query().Get("group_id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil || groupID == 0 {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	db := config.GetDB()

	// ✅ Check if user is a member of the group
	var isMember bool
	err = db.QueryRow(`SELECT EXISTS(SELECT 1 FROM group_members WHERE user_id = ? AND group_id = ?)`, user.ID, groupID).Scan(&isMember)
	if err != nil || !isMember {
		http.Error(w, "Unauthorized: You are not in this group", http.StatusForbidden)
		return
	}

	// ✅ Fetch chat history
	rows, err := db.Query(`
		SELECT gm.id, gm.sender_id, u.nickname, gm.content, gm.sent_at
		FROM group_messages gm
		JOIN users u ON gm.sender_id = u.id
		WHERE gm.group_id = ?
		ORDER BY gm.sent_at ASC`, groupID)
	if err != nil {
		http.Error(w, "Failed to retrieve chat history", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []map[string]any
	for rows.Next() {
		var msg struct {
			ID       int    `json:"id"`
			SenderID int    `json:"sender_id"`
			Nickname string `json:"nickname"`
			Content  string `json:"content"`
			SentAt   string `json:"sent_at"`
		}
		if err := rows.Scan(&msg.ID, &msg.SenderID, &msg.Nickname, &msg.Content, &msg.SentAt); err != nil {
			continue
		}
		messages = append(messages, map[string]any{
			"id":        msg.ID,
			"sender_id": msg.SenderID,
			"nickname":  msg.Nickname,
			"content":   msg.Content,
			"sent_at":   msg.SentAt,
		})
	}

	json.NewEncoder(w).Encode(messages)
}
