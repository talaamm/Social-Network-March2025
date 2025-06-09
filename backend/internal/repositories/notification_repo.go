package repositories

import (
	"database/sql"
	"log"

	"social-network/internal/models"
)

// NotificationRepository handles database operations for notifications.
type NotificationRepository struct {
	DB *sql.DB
}

// NewNotificationRepository creates a new instance of NotificationRepository.
func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{DB: db}
}

func (repo *NotificationRepository) GetNotifications(userID int) ([]models.Notification, error) {
	rows, err := repo.DB.Query(`
		SELECT id, user_id, type, message, is_read, created_at 
		FROM notifications 
		WHERE user_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		log.Printf("❌ Error retrieving notifications for User %d: %v", userID, err)
		return nil, err
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var notif models.Notification
		var createdAt sql.NullString

		err := rows.Scan(&notif.ID, &notif.UserID, &notif.Type, &notif.Message, &notif.IsRead, &createdAt)
		if err != nil {
			log.Println("❌ Error scanning notification row:", err)
			continue
		}

		// ✅ Print what Go actually reads from the database
		log.Printf("📌 DEBUG: Notification ID=%d, UserID=%d, Type=%s, Message=%s, IsRead=%v, CreatedAt=(Valid: %v, Value: %s)",
			notif.ID, notif.UserID, notif.Type, notif.Message, notif.IsRead, createdAt.Valid, createdAt.String)

		if createdAt.Valid {
			notif.CreatedAt = createdAt.String
		} else {
			notif.CreatedAt = "" // ✅ Use an empty string instead of omitting the field
		}

		notifications = append(notifications, notif)
	}

	log.Printf("📩 Retrieved %d notifications for User %d", len(notifications), userID)
	return notifications, nil
}

func (repo *NotificationRepository) MarkNotificationsAsRead(userID int, notifID int) error {
	_, err := repo.DB.Exec(`
		UPDATE notifications SET is_read = 1 WHERE id = ? AND user_id = ?`, notifID, userID)
	if err != nil {
		log.Printf("❌ Error marking notifications as read for User %d: %v", userID, err)
		return err
	}
	log.Printf("✅ Marked notifications as read for User %d", userID)
	return nil
}
