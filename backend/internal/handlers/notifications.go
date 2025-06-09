package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"social-network/internal/config"
	"social-network/internal/middlewars"
	"social-network/internal/repositories"
	ws "social-network/internal/websocket"

	"github.com/gorilla/websocket"
)

func GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	userID := middlewars.GetUserIDFromSession(w, r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := config.GetDB()
	repo := repositories.NewNotificationRepository(db)

	notifications, err := repo.GetNotifications(userID)
	if err != nil {
		log.Println("❌ Error fetching notifications:", err)
		http.Error(w, "Failed to retrieve notifications", http.StatusInternalServerError)
		return
	}

	// ✅ Return notifications as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(notifications)
	if err != nil {
		log.Println("❌ JSON Encoding Error:", err)
		http.Error(w, "Error encoding notifications", http.StatusInternalServerError)
	}
}

func WebSocketNotificationHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // ✅ Allow requests from any frontend domain
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	// conn, err := websocket.Upgrade(w, r, nil, 1024, 1024) // ✅ Fix: Added missing buffer sizes
	if err != nil {
		log.Println("❌ Failed to upgrade WebSocket:", err)
		http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}

	// Store connection in WebSocket manager
	ws.NotificationManager.RegisterClient(userID, conn)
	log.Printf("✅ WebSocket connected for User %d", userID)
}

func MarkNotificationsAsReadHandler(w http.ResponseWriter, r *http.Request) {
	userID := middlewars.GetUserIDFromSession(w, r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var notif struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&notif); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	db := config.GetDB()
	repo := repositories.NewNotificationRepository(db)

	err := repo.MarkNotificationsAsRead(notif.ID, userID)
	if err != nil {
		http.Error(w, "Failed to mark notifications as read", http.StatusInternalServerError)
		return
	}

	// ✅ Return success message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "All notifications marked as read"})
}

func ClearNotifications(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := config.GetDB()
	_, err := db.Exec("DELETE FROM notifications WHERE user_id = ?", user.ID)
	if err != nil {
		http.Error(w, "Failed to clear notifications", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
