package repositories

import (
	"database/sql"
	"fmt"
)

// EventRSVPRepository handles event RSVP operations
type EventRSVPRepository struct {
	DB *sql.DB
}

// NewEventRSVPRepository creates a new instance of EventRSVPRepository
func NewEventRSVPRepository(db *sql.DB) *EventRSVPRepository {
	return &EventRSVPRepository{DB: db}
}

// RSVPToEvent lets users mark themselves as "going" or "not going"
func (repo *EventRSVPRepository) RSVPToEvent(eventID, userID int, status string) error {
	var groupID int

	// ✅ Retrieve group_id for the given event_id
	err := repo.DB.QueryRow(`SELECT group_id FROM group_events WHERE id = ?`, eventID).Scan(&groupID)
	if err != nil {
		return fmt.Errorf("failed to get group_id: %w", err)
	}

	_, err = repo.DB.Exec(`
        INSERT INTO group_event_attendees (event_id, member_id, group_id, status) 
        VALUES (?, ?, ?, ?) 
        ON CONFLICT(event_id, member_id) 
        DO UPDATE SET status = ?`, eventID, userID, groupID, status, status)

	return err
}

// GetRSVPCount returns the number of users who are "going" to an event
func (repo *EventRSVPRepository) GetRSVPCount(eventID int) (int, error) {
	var count int
	err := repo.DB.QueryRow(`
        SELECT COUNT(*) FROM event_rsvps WHERE event_id = ? AND status = 'going'`,
		eventID).Scan(&count)
	return count, err
}

func (repo *EventRSVPRepository) GetUserRSVPStatus(eventID, userID int) (string, error) {
	var status string
	err := repo.DB.QueryRow(`
        SELECT status FROM group_event_attendees 
        WHERE event_id = ? AND member_id = ?`, eventID, userID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // No existing RSVP found
		}
		return "", err // Database error
	}
	return status, nil
}
