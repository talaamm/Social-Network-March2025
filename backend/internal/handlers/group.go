package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"social-network/internal/config"
	"social-network/internal/middlewars"
	"social-network/internal/models"
	"social-network/internal/repositories"
)

func GetGroupDetailsHandler(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.Atoi(r.URL.Query().Get("group_id"))
	if err != nil || groupID == 0 {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	db := config.GetDB()

	var group models.Group
	query := "SELECT id, group_name, description, creator_id FROM groups WHERE id = ?"
	err = db.QueryRow(query, groupID).Scan(&group.ID, &group.Name, &group.Description, &group.CreatorID)
	if err != nil {
		log.Println("Error fetching group details:", err)
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}
	group.CreatorName, err = GetGroupCreatorNick(db, group.CreatorID)
	if err != nil {
		log.Println("Error getting nickname, id:", group.CreatorID, err)
	}
	group.MemberCount, _ = GetGroupMemberCount(db, groupID)
	log.Println("fetched Group datails successfully!")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}

func GetGroupCreatorNick(db *sql.DB, CreatorID int) (string, error) {
	var CreatorName string
	query := "SELECT nickname FROM users WHERE id = ?"
	err := db.QueryRow(query, CreatorID).Scan(&CreatorName)
	if err != nil {
		log.Println("Error fetching group creator nickname:", err)
		// http.Error(w, "Group creator not found", http.StatusNotFound)
		return "", err
	}
	log.Println("nickname to be returned:", CreatorName)
	return CreatorName, nil
}

func GetUserCreatedGroupsHandler(w http.ResponseWriter, r *http.Request) {
	userID := middlewars.GetUserIDFromSession(w, r)
	db := config.GetDB()
	var groups []models.Group
	query := "SELECT id, group_name, description FROM groups WHERE creator_id = ? ORDER BY id DESC "
	rows, err := db.Query(query, userID)
	if err != nil {
		log.Println("Error fetching created groups:", err)
		http.Error(w, "Failed to fetch created groups", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(&group.ID, &group.Name, &group.Description); err != nil {
			log.Println("Error scanning group row:", err)
			continue
		}
		group.MemberCount, err = GetGroupMemberCount(db, group.ID)
		if err != nil {
			log.Println("error counting group members")
			return
		}
		groups = append(groups, group)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

func GetUserGroupsHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := config.GetDB()

	query := `
        SELECT g.id, g.group_name, g.description , g.creator_id
        FROM groups g
        INNER JOIN group_members gm ON g.id = gm.group_id
        WHERE gm.id = ? AND gm.status = 'approved'
        AND g.creator_id != ?`

	rows, err := db.Query(query, user.ID, user.ID)
	if err != nil {
		log.Println("Error fetching user groups:", err)
		http.Error(w, "Error fetching groups", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(&group.ID, &group.Name, &group.Description, &group.CreatorID); err != nil {
			log.Println("Error scanning group:", err)
			continue
		}
		group.MemberCount, err = GetGroupMemberCount(db, group.ID)
		if err != nil {
			log.Println("Error counting group members:", err)
			continue
		}
		group.CreatorName, err = GetGroupCreatorNick(db, group.CreatorID)
		if err != nil {
			log.Println("Error getting nickname, id:", group.CreatorID, err)
			continue
		}
		groups = append(groups, group)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(groups)
}

func GetNonMemberGroupsHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := config.GetDB()

	query := `
        SELECT 
            g.id, g.group_name, g.description, g.creator_id, 
            u.nickname AS creator_nickname,
            (SELECT COUNT(*) FROM group_members WHERE group_id = g.id AND status = 'approved') AS member_count,
            gm.status AS membership_status
        FROM groups g
        LEFT JOIN group_members gm ON g.id = gm.group_id AND gm.id = ?
        JOIN users u ON g.creator_id = u.id
        WHERE gm.id IS NULL OR gm.status = 'pending'
    `

	rows, err := db.Query(query, user.ID)
	if err != nil {
		log.Println("Error fetching groups:", err)
		http.Error(w, "Error fetching groups", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var availableGroups []map[string]any
	var pendingGroups []map[string]any

	for rows.Next() {
		var groupID, creatorID, memberCount int
		var name, description, creatorNickname, membershipStatus sql.NullString

		err := rows.Scan(&groupID, &name, &description, &creatorID, &creatorNickname, &memberCount, &membershipStatus)
		if err != nil {
			log.Println("Error scanning group:", err)
			continue
		}

		groupData := map[string]any{
			"id":               groupID,
			"name":             name.String,
			"description":      description.String,
			"creator_id":       creatorID,
			"creator_nickname": creatorNickname.String,
			"member_count":     memberCount,
		}

		if membershipStatus.Valid && membershipStatus.String == "pending" {
			pendingGroups = append(pendingGroups, groupData)
		} else {
			availableGroups = append(availableGroups, groupData)
		}
	}

	response := map[string]any{
		"pending_groups":   pendingGroups,
		"available_groups": availableGroups,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func CreateGroupHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var group models.Group
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		log.Println(err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	group.CreatorID = user.ID

	db := config.GetDB()
	repo := repositories.NewGroupRepository(db)
	var err error
	group.ID, err = repo.CreateGroup(&group)
	if err != nil {
		log.Println("Error creating group:", err)
		http.Error(w, "Failed to create group", http.StatusInternalServerError)
		return
	}
	memRepo := repositories.NewGroupMemberRepository(db)
	err = memRepo.AddUserToGroup(group.ID, user.ID, "approved", user.Nickname)
	fmt.Println(err)
	fmt.Println("admin added to group", user.ID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"message": "Group created successfully", "group_id": group.ID})
}
