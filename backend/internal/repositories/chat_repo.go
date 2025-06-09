package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"social-network/internal/models"
)

// ChatRepository handles chat-related database operations
type ChatRepository struct {
	DB *sql.DB
}

// NewChatRepository creates a new instance of ChatRepository
func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{DB: db}
}

func (repo *ChatRepository) GetMessages(user1, user2 int) ([]models.ChatMessage, error) {
	rows, err := repo.DB.Query(`
		SELECT id, sender_id, receiver_id, content, sent_at 
		FROM messages 
		WHERE (sender_id = ? AND receiver_id = ?) 
		OR (sender_id = ? AND receiver_id = ?) 
		ORDER BY sent_at ASC`, user1, user2, user2, user1)
	if err != nil {
		log.Println("❌ Error fetching messages:", err)
		return nil, err
	}
	defer rows.Close()

	var messages []models.ChatMessage
	for rows.Next() {
		var msg models.ChatMessage
		if err := rows.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.Content, &msg.SentAt); err != nil {
			log.Println("❌ Error scanning chat history row:", err)
			return nil, err
		}
		messages = append(messages, msg)
	}

	log.Printf("📜 Retrieved %d messages between users %d and %d", len(messages), user1, user2)
	return messages, nil
}

func (repo *ChatRepository) SaveMessage(senderID, receiverID int, content string) error {
	query := `INSERT INTO messages (sender_id, receiver_id, content, sent_at) VALUES (?, ?, ?, ?)`

	result, err := repo.DB.Exec(query, senderID, receiverID, content, time.Now())
	if err != nil {
		log.Println("❌ Error saving message:", err)
		return fmt.Errorf("failed to save message: %w", err) // Return wrapped error for better debugging
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("❌ no message inserted")
	}

	log.Printf("📩 Message saved: Sender %d -> Receiver %d: %s", senderID, receiverID, content)
	return nil
}
