package repositories

import (
	"database/sql"
	"log"

	"social-network/internal/models"
)

// GroupEventRepository handles group event-related operations
type GroupEventRepository struct {
	DB *sql.DB
}

// NewGroupEventRepository creates a new instance of GroupEventRepository
func NewGroupEventRepository(db *sql.DB) *GroupEventRepository {
	return &GroupEventRepository{DB: db}
}

// CreateGroupEvent adds a new event to a group
func (repo *GroupEventRepository) CreateGroupEvent(event *models.GroupEvent) error {
	_, err := repo.DB.Exec(`
        INSERT INTO group_events (group_id, creator_id, title, description, event_date)
        VALUES (?, ?, ?, ?, ?)`,
		event.GroupID, event.CreatorID, event.Title, event.Description, event.EventDate)
	return err
}

// ✅ Get all events for a group
func (repo *GroupEventRepository) GetGroupEvents(groupID int) ([]models.GroupEvent, error) {
	rows, err := repo.DB.Query(`
		SELECT id, group_id, creator_id, title, description, event_date, created_at 
		FROM group_events 
		WHERE group_id = ? 
		ORDER BY event_date DESC`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.GroupEvent
	for rows.Next() {
		var event models.GroupEvent
		err := rows.Scan(&event.ID, &event.GroupID, &event.CreatorID, &event.Title, &event.Description, &event.EventDate, &event.CreatedAt)
		if err != nil {
			return nil, err
		}
		err = repo.DB.QueryRow("SELECT COUNT(*) FROM group_event_attendees WHERE event_id = ? AND status = 'going'", event.ID).Scan(&event.Going)
		err = repo.DB.QueryRow("SELECT COUNT(*) FROM group_event_attendees WHERE event_id = ? AND status = 'not going'", event.ID).Scan(&event.NotGoing)
		if err != nil {
			log.Println("error counting going / not going", err)
		}
		repo.DB.QueryRow("SELECT nickname FROM users WHERE id = ?", event.CreatorID).Scan(&event.Creator)
		events = append(events, event)
	}

	return events, nil
}
