package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"social-network/internal/config"
	"social-network/internal/repositories"
	ws "social-network/internal/websocket"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // ✅ Allows all frontend origins (e.g., localhost:5173)
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[int]*Client), // ✅ Changed from `map[*Client]bool`
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.userID] = client // ✅ Store by userID
		case client := <-h.unregister:
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	db := config.GetDB()                       // Get the database connection
	repo := repositories.NewChatRepository(db) // Initialize chat repository

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		// ✅ Convert raw message to structured JSON
		var msg struct {
			ReceiverID int    `json:"receiver_id"`
			Content    string `json:"content"`
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println("❌ Invalid message format:", err)
			continue
		}
		log.Printf("📩 Received message: Sender %d -> Receiver %d: %s", c.userID, msg.ReceiverID, msg.Content)

		// ✅ Save message to database
		if err := repo.SaveMessage(c.userID, msg.ReceiverID, msg.Content); err != nil {
			log.Println("❌ Failed to store message:", err)
			continue
		}

		formattedMessage, _ := json.Marshal(map[string]any{
			"sender_id":   c.userID, // Attach sender ID
			"receiver_id": msg.ReceiverID,
			"content":     msg.Content,
			"sent_at":     time.Now().Format(time.RFC3339),
		})
		log.Printf("🚀 Sending message to User %d", msg.ReceiverID)

		// Send message to only the recipient
		c.hub.sendToClient(msg.ReceiverID, formattedMessage)
		repo := repositories.NewUserRepository(db)
		userData, err := repo.GetUserDataById(c.userID)
		if err != nil {
			log.Println("Error retrieving userData:", err)
			// http.Error(w, "Failed to retrieve userData", http.StatusInternalServerError)
			return
		} // ✅ **Send Notification ONLY if the recipient is NOT actively chatting**

		if !c.hub.isUserActive(msg.ReceiverID) {
			notificationMessage := fmt.Sprintf("New message from %s", userData.Nickname) // Assuming `c.username` stores the sender's name
			ws.SendNotification(msg.ReceiverID, "message", notificationMessage)
		}
	}
}

func (h *Hub) isUserActive(userID int) bool {
	for _, client := range h.clients {
		if client.userID == userID { // ✅ Correct comparison
			return true
		}
	}
	return false
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			log.Printf("📩 Writing message to User %d: %s", c.userID, string(message)) // ✅ Debugging log
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

type Hub struct {
	clients    map[int]*Client // ✅ Stores clients by userID instead of pointer
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}
type Client struct {
	userID int
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
}

func (h *Hub) sendToClient(receiverID int, message []byte) {
	if client, exists := h.clients[receiverID]; exists {
		select {
		case client.send <- message:
			log.Printf("📩 Successfully sent message to User %d", receiverID)
		default:
			log.Printf("❌ Failed to send message to User %d (client closed)", receiverID)
			close(client.send)
			delete(h.clients, receiverID)
		}
	} else {
		log.Printf("⚠️ No active connection found for User %d", receiverID)
	}
}

// serveWs handles websocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true } // ✅ Allow frontend

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID == 0 {
		log.Println("❌ Invalid user ID in WebSocket connection")
		conn.Close()
		return
	}

	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), userID: userID}

	log.Printf("✅ User %d connected to WebSocket", userID)

	hub.register <- client
	go client.writePump()
	go client.readPump()
}
