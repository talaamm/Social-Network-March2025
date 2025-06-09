package websocket

import (
	"log"
	"sync"

	"social-network/internal/config"
	"social-network/internal/models"

	"github.com/gorilla/websocket"
)

var NotificationManager = &WebSocketNotificationManager{
	Clients: make(map[int]*WebSocketConn),
}

type WebSocketNotificationManager struct {
	Clients map[int]*WebSocketConn
	Mutex   sync.Mutex
}
type WebSocketConn struct {
	Conn  *websocket.Conn
	Mutex sync.Mutex
}

func BroadcastPostUpdate(postID, likes, dislikes int) {
	NotificationManager.Mutex.Lock()
	defer NotificationManager.Mutex.Unlock()

	notification := map[string]any{
		"type":     "post_update",
		"post_id":  postID,
		"likes":    likes,
		"dislikes": dislikes,
	}

	for _, client := range NotificationManager.Clients {
		client.Mutex.Lock()
		err := client.Conn.WriteJSON(notification)
		client.Mutex.Unlock()

		if err != nil {
			log.Printf("❌ Failed to send post update: %v", err)
			client.Conn.Close()
		}
	}
}

func BroadcastGroupPostUpdate(groupID, memberID, postID int, authorName, content, createdAt string) {
	NotificationManager.Mutex.Lock()
	defer NotificationManager.Mutex.Unlock()

	notification := map[string]any{
		"type":       "group_post_update",
		"group_id":   groupID,
		"post_id":    postID,
		"nickname":   authorName,
		"content":    content,
		"created_at": createdAt,
		"member_id":  memberID,
	}
	// models.GroupPost

	for _, client := range NotificationManager.Clients {
		client.Mutex.Lock()
		err := client.Conn.WriteJSON(notification)
		client.Mutex.Unlock()

		if err != nil {
			log.Printf("❌ Failed to send group post update: %v", err)
			client.Conn.Close()
		}
	}
}
func BroadcastGroupEvents(groupID int) {
	NotificationManager.Mutex.Lock()
	defer NotificationManager.Mutex.Unlock()

	notification := map[string]any{
		"type":     "new_group_event",
		"group_id": groupID,
		// "event_id":    postID,
		// "nickname":   authorName,
		// "title":    content,
		// "created_at": time.Now().Unix(),
	}
	// models.GroupPost

	for _, client := range NotificationManager.Clients {
		client.Mutex.Lock()
		err := client.Conn.WriteJSON(notification)
		client.Mutex.Unlock()

		if err != nil {
			log.Printf("❌ Failed to send group post update: %v", err)
			client.Conn.Close()
		}
	}
}

func SendNotification(userID int, notifType, message string) {
	log.Printf("📢 Sending real-time notification to User %d: %s", userID, message)

	notification := models.Notification{
		UserID:  userID,
		Type:    notifType,
		Message: message,
		IsRead:  false,
	}

	// Get WebSocket client
	NotificationManager.Mutex.Lock()
	client, exists := NotificationManager.Clients[userID]
	NotificationManager.Mutex.Unlock()

	if exists && client != nil {
		client.Mutex.Lock()
		err := client.Conn.WriteJSON(notification)
		client.Mutex.Unlock()

		if err != nil {
			log.Printf("❌ Error sending WebSocket notification: %v", err)
			NotificationManager.RemoveClient(userID) // Remove broken connection
		} else {
			log.Printf("📨 Sent real-time notification to User %d", userID)
		}
	} else {
		log.Printf("📌 User %d is offline. Notification stored.", userID)
	}
	storeNotification(notification) // Store if user is offline
}

func storeNotification(notification models.Notification) {
	db := config.GetDB()
	_, err := db.Exec(`
		INSERT INTO notifications (user_id, type, message, is_read) 
		VALUES (?, ?, ?, 0)`,
		notification.UserID, notification.Type, notification.Message)
	if err != nil {
		log.Printf("❌ Failed to store notification: %v", err)
	}
}

// ✅ Remove a Disconnected WebSocket Client
func (wm *WebSocketNotificationManager) RemoveClient(userID int) {
	wm.Mutex.Lock()
	defer wm.Mutex.Unlock()

	if client, exists := wm.Clients[userID]; exists {
		client.Conn.Close()
		delete(wm.Clients, userID)
		log.Printf("⚠️ User %d disconnected from notifications.", userID)
	}
}

func (wm *WebSocketNotificationManager) RegisterClient(userID int, conn *websocket.Conn) {
	wm.Mutex.Lock()
	defer wm.Mutex.Unlock()

	// ✅ Close previous connection if user reconnects
	if oldClient, exists := wm.Clients[userID]; exists {
		log.Printf("⚠️ Closing old connection for User %d.", userID)
		oldClient.Conn.Close()
		delete(wm.Clients, userID)
	}

	wm.Clients[userID] = &WebSocketConn{Conn: conn}
	log.Printf("✅ User %d connected for real-time notifications.", userID)
}
