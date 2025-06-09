package repositories

import (
	"database/sql"
	"log"

	"social-network/internal/config"
)

type GroupChatRepository struct {
	DB *sql.DB
}

func NewGroupChatRepository(db *sql.DB) *GroupChatRepository {
	return &GroupChatRepository{DB: db}
}

func (repo *GroupChatRepository) SaveGroupChatMessage(groupID, senderID int, content string) error {
	db := config.GetDB()
	_, err := db.Exec(`
			INSERT INTO group_messages (group_id, sender_id, content)
			VALUES (?, ?, ?)`, groupID, senderID, content)
	if err != nil {
		log.Println("❌ Error saving message:", err)
	}
	return err
}
