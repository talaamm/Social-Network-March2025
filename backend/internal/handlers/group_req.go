package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"social-network/internal/config"
	"social-network/internal/middlewars"
	"social-network/internal/repositories"
	"social-network/internal/websocket"
)

func RequestToJoinGroupHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupID, err := strconv.Atoi(r.URL.Query().Get("group_id"))
	if err != nil || groupID == 0 {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	db := config.GetDB()
	repo := repositories.NewGroupRepository(db)

	err = repo.RequestToJoinGroup(groupID, user.ID, user.Nickname)
	if err != nil {
		http.Error(w, "Failed to request to join group", http.StatusInternalServerError)
		return
	}
	gname, cid, _ := GetGroupNameAndCreator(db, groupID)
	websocket.SendNotification(cid, "request", user.Nickname+" has requested to join your group: "+gname)
	// w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Request sent"})
}

func GetGroupNameAndCreator(db *sql.DB, groupID int) (string, int, error) {
	var groupName string
	var creatorID int

	query := `SELECT group_name, creator_id FROM groups WHERE id = ?`
	err := db.QueryRow(query, groupID).Scan(&groupName, &creatorID)
	if err != nil {
		return "", 0, err
	}

	return groupName, creatorID, nil
}

func GetPendingGroupRequestsHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := config.GetDB()

	// Fetch all pending requests where the current user is the group creator
	query := `
		SELECT gm.id, gm.group_id, gm.username, g.group_name
		FROM group_members gm
		INNER JOIN groups g ON gm.group_id = g.id
		WHERE g.creator_id = ? AND gm.status = 'pending'
	`
	rows, err := db.Query(query, user.ID)
	if err != nil {
		log.Println("❌ Error fetching pending group join requests:", err)
		http.Error(w, "Failed to fetch requests", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var requests []map[string]any
	for rows.Next() {
		var requestID, groupID int
		var nickname, groupName string
		if err := rows.Scan(&requestID, &groupID, &nickname, &groupName); err != nil {
			log.Println("❌ Error scanning request:", err)
			continue
		}

		requests = append(requests, map[string]any{
			"id":         requestID,
			"nickname":   nickname,
			"group_name": groupName,
			"group_id":   groupID,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(requests)
}
