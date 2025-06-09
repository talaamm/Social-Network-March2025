package websocket

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"social-network/internal/config"
	"social-network/internal/middlewars"
	"social-network/internal/repositories"

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

type GroupHub struct {
	clients    map[int]map[*Client]bool // ✅ Tracks connections by groupID
	broadcast  chan GroupMessage        // ✅ Channel for group messages
	register   chan *Client
	unregister chan *Client
}

type GroupMessage struct {
	GroupID        int    `json:"group_id"`
	SenderID       int    `json:"sender_id"`
	Content        string `json:"content"`
	SentAt         string `json:"sent_at"`
	SenderNickname string `json:"sender_nickname"`
}

// ✅ Create New Group Hub
func NewGroupHub() *GroupHub {
	return &GroupHub{
		clients:    make(map[int]map[*Client]bool),
		broadcast:  make(chan GroupMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// ✅ GroupHub: Manages clients and broadcasts messages
func (h *GroupHub) Run() {
	for {
		select {
		case client := <-h.register:
			if _, ok := h.clients[client.groupID]; !ok {
				h.clients[client.groupID] = make(map[*Client]bool)
			}
			h.clients[client.groupID][client] = true

		case client := <-h.unregister:
			if _, ok := h.clients[client.groupID]; ok {
				delete(h.clients[client.groupID], client)
				close(client.send)
				if len(h.clients[client.groupID]) == 0 {
					delete(h.clients, client.groupID) // Remove empty groups
				}
			}

		case message := <-h.broadcast:
			if clients, ok := h.clients[message.GroupID]; ok {
				formattedMessage, _ := json.Marshal(message)
				for client := range clients {
					select {
					case client.send <- formattedMessage:
					default:
						close(client.send)
						delete(h.clients[message.GroupID], client)
					}
				}
			}
		}
	}
}

var groupName string

type Client struct {
	userID  int
	groupID int
	hub     *GroupHub
	conn    *websocket.Conn
	send    chan []byte
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	db := config.GetDB()
	repo := repositories.NewGroupChatRepository(db)

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		message = bytes.TrimSpace(message)

		var msg GroupMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}
		msg.SenderID = c.userID
		msg.SentAt = time.Now().Format(time.RFC3339)

		// ✅ Save message to database
		if err := repo.SaveGroupChatMessage(msg.GroupID, msg.SenderID, msg.Content); err != nil {
			log.Println("❌ Failed to store message:", err)
			continue
		}

		query := `SELECT group_name FROM groups WHERE id = ?`
		db.QueryRow(query, msg.GroupID).Scan(&groupName)
		notifMsg := msg.SenderNickname + " sent a message to: " + groupName
		SendNotification(c.userID, "message", notifMsg)

		// ✅ Broadcast message
		c.hub.broadcast <- msg
	}
}

func ServeGroupChatWs(hub *GroupHub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("❌ WebSocket Upgrade Failed:", err)
		return
	}

	// ✅ Extract user_id and group_id from query parameters
	userIDStr := r.URL.Query().Get("user_id")
	groupIDStr := r.URL.Query().Get("group_id")

	userID, err := strconv.Atoi(userIDStr)
	groupID, err2 := strconv.Atoi(groupIDStr)

	if err != nil || err2 != nil || userID == 0 || groupID == 0 {
		log.Println("❌ Invalid WebSocket parameters")
		conn.Close()
		return
	}

	// ✅ Create client and register it to the hub
	client := &Client{userID: userID, groupID: groupID, hub: hub, conn: conn, send: make(chan []byte, 256)}
	hub.register <- client

	log.Printf("✅ User %d joined Group %d via WebSocket", userID, groupID)

	// ✅ Start read and write pumps
	go client.readPump()
	go client.writePump()
}

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

// ✅ Write Messages to WebSocket
func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for message := range c.send {
		c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}
