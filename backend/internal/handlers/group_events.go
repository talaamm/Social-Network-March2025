package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"social-network/internal/config"
	"social-network/internal/middlewars"
	"social-network/internal/models"
	"social-network/internal/repositories"
	"social-network/internal/websocket"
)

// CreateGroupEventHandler handles event creation inside a group
func CreateGroupEventHandler(w http.ResponseWriter, r *http.Request) {
	userID := middlewars.GetUserIDFromSession(w, r)
	if userID == 0 {
		log.Println("\nUser id missing for creating event")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var event models.GroupEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Println("\nerror decoding event req body", r.Body, "\n", err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if event.Title == "" || event.Description == "" || event.GroupID == 0 || event.EventDate == "" {
		log.Println("\nsome fields are missing", event.Title, event.Description, event.GroupID, event.EventDate)
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}
	/*
	   type GroupEvent struct {
	   	ID          int    `json:"id"`
	   	GroupID     int    `json:"group_id"`
	   	CreatorID   int    `json:"creator_id"`
	   	Title       string `json:"title"`
	   	Description string `json:"description"`
	   	EventDate   string `json:"event_date"`
	   	CreatedAt   string `json:"created_at"`
	   }
	*/
	event.CreatorID = userID
	db := config.GetDB()
	repo := repositories.NewGroupEventRepository(db)

	err := repo.CreateGroupEvent(&event)
	if err != nil {
		log.Println("Error creating group event:", err)
		http.Error(w, "Failed to create event", http.StatusInternalServerError)
		return
	}
	websocket.BroadcastGroupEvents(event.GroupID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Group event created successfully"})
}

// GetGroupEventsHandler retrieves all events for a given group
func GetGroupEventsHandler(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.Atoi(r.URL.Query().Get("group_id"))
	if err != nil || groupID == 0 {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	db := config.GetDB()
	repo := repositories.NewGroupEventRepository(db)

	events, err := repo.GetGroupEvents(groupID)
	if err != nil {
		log.Println("Error fetching group events:", err)
		http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(events)
}

// RSVPToEventHandler handles user RSVPs to an event
func RSVPEventHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	eventID, err := strconv.Atoi(r.URL.Query().Get("event_id"))
	if err != nil || eventID == 0 {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	var requestBody struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	db := config.GetDB()
	repo := repositories.NewEventRSVPRepository(db)

	// Check if user already has an RSVP
	currentStatus, err := repo.GetUserRSVPStatus(eventID, user.ID)
	if err != nil {
		log.Println("Error checking RSVP status:", err)
		http.Error(w, "Error checking RSVP", http.StatusInternalServerError)
		return
	}

	// Only update if the user is changing their RSVP
	if currentStatus != requestBody.Status {
		err = repo.RSVPToEvent(eventID, user.ID, requestBody.Status)
		if err != nil {
			log.Println("❌ Failed to RSVP:", err)
			http.Error(w, "Failed to RSVP", http.StatusInternalServerError)
			return
		}
	}
	var creator_id, gid int
	var title string

	err = db.QueryRow("SELECT creator_id, group_id, title FROM group_events WHERE id = ?", eventID).Scan(&creator_id, &gid, &title)
	fmt.Println(err)
	// Send WebSocket notification (optional)
	fmt.Println("notification sent to event creator: ", creator_id)
	message := fmt.Sprintf("%s is %s to event %s", user.Nickname, requestBody.Status, title)
	websocket.SendNotification(creator_id, "event_rsvp", message)
	// w.WriteHeader(http.StatusOK)
	websocket.BroadcastGroupEvents(gid)
	json.NewEncoder(w).Encode(map[string]string{"message": "RSVP updated successfully"})
}

// GetRSVPCountHandler returns the number of users attending an event
func GetRSVPCountHandler(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.Atoi(r.URL.Query().Get("event_id"))
	if err != nil || eventID == 0 {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	db := config.GetDB()
	repo := repositories.NewEventRSVPRepository(db)

	count, err := repo.GetRSVPCount(eventID)
	if err != nil {
		log.Println("Error retrieving RSVP count:", err)
		http.Error(w, "Failed to retrieve RSVP count", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"going_count": count})
}
