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
	"social-network/internal/websocket"
)

func GetGroupMemberCount(db *sql.DB, groupID int) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM group_members WHERE group_id = ? AND status = 'approved'"

	err := db.QueryRow(query, groupID).Scan(&count)
	if err != nil {
		log.Println("❌ Error counting group members:", err)
		return 0, err
	}

	return count, nil
}

func LeaveGroupHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupID, err := strconv.Atoi(r.URL.Query().Get("group_id"))
	if err != nil {
		log.Println("invalid group id:", err)
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	db := config.GetDB()
	repo := repositories.NewGroupMemberRepository(db)
	groupRep := repositories.NewGroupRepository(db)
	err = repo.RemoveUserFromGroup(groupID, user.ID)
	if err != nil {
		log.Println("error while removing user from db,", err)
		http.Error(w, "Failed to leave group", http.StatusInternalServerError)
		return
	}
	err = groupRep.DeleteGroup(groupID, user.ID)
	if err != nil {
		log.Println("user removed from repo but group still exist!", err)
		var creatorID int
		var groupName string
		db.QueryRow(`SELECT creator_id FROM groups WHERE id = ?`, groupID).Scan(&creatorID)
		db.QueryRow(`SELECT name FROM groups WHERE id = ?`, groupID).Scan(&groupName)
		websocket.SendNotification(creatorID, "NOTICE", user.Nickname+" has left your group: "+groupName)
	}
	log.Println("user removed successfully")
	// w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Left group successfully"})
}

func InviteUserToGroupHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupID, err1 := strconv.Atoi(r.URL.Query().Get("group_id"))
	invitedUserID, err2 := strconv.Atoi(r.URL.Query().Get("invited_user_id"))
	if err1 != nil || err2 != nil || groupID == 0 || invitedUserID == 0 {
		http.Error(w, "Invalid request parameters", http.StatusBadRequest)
		return
	}

	db := config.GetDB()
	repo := repositories.NewGroupRepository(db)

	// ✅ Check if the invited user is already in the group
	if repo.IsUserInGroup(groupID, invitedUserID) {
		http.Error(w, "User is already a group member", http.StatusConflict)
		return
	}

	// ✅ Check if the invitation already exists
	if repo.IsInvitationPending(groupID, invitedUserID) {
		http.Error(w, "Invitation already sent", http.StatusConflict)
		return
	}

	// ✅ Insert invitation into the database
	err := repo.InviteUserToGroup(groupID, user.ID, invitedUserID)
	if err != nil {
		log.Println("Error inviting user to group:", err)
		http.Error(w, "Failed to send invitation", http.StatusInternalServerError)
		return
	}
	gname, _, _ := GetGroupNameAndCreator(db, groupID)
	websocket.SendNotification(invitedUserID, "invitation", user.Nickname+" has invited you to join group: "+gname)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Invitation sent successfully"})
}

func GetFollowersToInviteHandler(w http.ResponseWriter, r *http.Request) {
	// ✅ Get logged-in user
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// ✅ Get Group ID from request
	groupID, err := strconv.Atoi(r.URL.Query().Get("group_id"))
	if err != nil || groupID == 0 {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	db := config.GetDB()

	query := `
        SELECT DISTINCT u.id, u.nickname, u.first_name, u.last_name 
        FROM followers f
        JOIN users u ON u.id = f.follower_id OR u.id = f.following_id
        WHERE (f.follower_id = ? OR f.following_id = ?) 
          AND f.status = 'accepted'
          AND u.id != ? -- Prevent self-invitation
          AND u.id NOT IN (SELECT id FROM group_members WHERE group_id = ? AND status = 'approved')
          AND u.id NOT IN (SELECT invited_user_id FROM group_invitations WHERE group_id = ?)
    `

	rows, err := db.Query(query, user.ID, user.ID, user.ID, groupID, groupID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to retrieve users to invite", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// DISTINCT in SQL: Ensures the same user is not selected multiple times.
	// ✅ Use a map to prevent duplicates in Go (as an extra safeguard)
	uniqueUsers := make(map[int]models.User)
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Nickname, &u.FirstName, &u.LastName)
		if err != nil {
			log.Println("Error scanning user:", err)
			continue
		}
		uniqueUsers[u.ID] = u
	}

	// ✅ Convert map to slice
	var users []models.User
	for _, u := range uniqueUsers {
		users = append(users, u)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func GetGroupInvitationsHandler(w http.ResponseWriter, r *http.Request) {
	// ✅ Get logged-in user
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := config.GetDB()

	// ✅ Fetch invitations where the user is invited
	query := `
        SELECT gi.id, gi.group_id, gi.member_id, g.group_name, u.nickname 
        FROM group_invitations gi
        JOIN groups g ON gi.group_id = g.id
        JOIN users u ON gi.member_id = u.id
        WHERE gi.invited_user_id = ?
    `

	rows, err := db.Query(query, user.ID)
	if err != nil {
		http.Error(w, "Failed to retrieve group invitations", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var invitations []map[string]any
	for rows.Next() {
		var invitation map[string]any
		var id, groupID, inviterID int
		var groupName, inviterNickname string

		err := rows.Scan(&id, &groupID, &inviterID, &groupName, &inviterNickname)
		if err != nil {
			log.Println("Error scanning invitation:", err)
			continue
		}

		invitation = map[string]any{
			"id":               id,
			"group_id":         groupID,
			"inviter_id":       inviterID,
			"group_name":       groupName,
			"inviter_nickname": inviterNickname,
		}

		invitations = append(invitations, invitation)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(invitations)
}

func AcceptGroupInvitationHandler(w http.ResponseWriter, r *http.Request) {
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

	// ✅ Add user as approved member
	_, err = db.Exec(`INSERT INTO group_members (id, group_id, username, status) VALUES (?, ?, ?, 'approved')`,
		user.ID, groupID, user.Nickname)
	if err != nil {
		http.Error(w, "Failed to accept invitation", http.StatusInternalServerError)
		return
	}

	// ✅ Remove invitation after acceptance
	_, _ = db.Exec(`DELETE FROM group_invitations WHERE group_id = ? AND invited_user_id = ?`, groupID, user.ID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Invitation accepted"})
}

func RejectGroupInvitationHandler(w http.ResponseWriter, r *http.Request) {
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

	// ✅ Remove invitation
	_, err = db.Exec(`DELETE FROM group_invitations WHERE group_id = ? AND invited_user_id = ?`, groupID, user.ID)
	if err != nil {
		http.Error(w, "Failed to reject invitation", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Invitation rejected"})
}

func GetUserMemberGroupsHandler(w http.ResponseWriter, r *http.Request) {
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
        WHERE gm.id = ? AND gm.status = 'approved'`

	rows, err := db.Query(query, user.ID)
	if err != nil {
		log.Println("❌ Error fetching user groups:", err)
		http.Error(w, "Error fetching groups", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(&group.ID, &group.Name, &group.Description, &group.CreatorID); err != nil {
			log.Println("❌ Error scanning group:", err)
			continue
		}
		group.MemberCount, err = GetGroupMemberCount(db, group.ID)
		if err != nil {
			log.Println("❌ Error counting group members:", err)
			continue
		}
		group.CreatorName, err = GetGroupCreatorNick(db, group.CreatorID)
		if err != nil {
			log.Println("❌ Error getting nickname for creator:", err)
			continue
		}
		groups = append(groups, group)
	}
	log.Println("Retreived user groups successfully")
	// w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(groups)
}

func ApproveMembershipHandler(w http.ResponseWriter, r *http.Request) {
	adminID := middlewars.GetUserIDFromSession(w, r)
	if adminID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupID, err := strconv.Atoi(r.URL.Query().Get("group_id"))
	userID, err2 := strconv.Atoi(r.URL.Query().Get("user_id"))
	if err != nil || err2 != nil || groupID == 0 || userID == 0 {
		http.Error(w, "Invalid group or user ID", http.StatusBadRequest)
		return
	}

	db := config.GetDB()
	repo := repositories.NewGroupRepository(db)

	err = repo.ApproveMembership(groupID, userID, adminID)
	if err != nil {
		log.Println("Error approving member:", err)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	gname, _, _ := GetGroupNameAndCreator(db, groupID)
	websocket.SendNotification(userID, "approved", "Your request to joing group: "+gname+" has been approved")
	// w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Membership approved"})
}

// GetGroupMembersHandler retrieves all approved members of a group
func GetGroupMembersHandler(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.Atoi(r.URL.Query().Get("group_id"))
	if err != nil || groupID == 0 {
		log.Println("id missing", groupID)
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	db := config.GetDB()
	repo := repositories.NewGroupRepository(db)

	members, err := repo.GetGroupMembers(groupID)
	if err != nil {
		log.Println("Error getting group members:", err)
		http.Error(w, "Failed to fetch members", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(members)
}

// RejectMembershipHandler allows group admins to reject members
func RejectMembershipHandler(w http.ResponseWriter, r *http.Request) {
	adminID := middlewars.GetUserIDFromSession(w, r)
	if adminID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupID, err := strconv.Atoi(r.URL.Query().Get("group_id"))
	userID, err2 := strconv.Atoi(r.URL.Query().Get("user_id"))
	if err != nil || err2 != nil || groupID == 0 || userID == 0 {
		http.Error(w, "Invalid group or user ID", http.StatusBadRequest)
		return
	}

	db := config.GetDB()
	repo := repositories.NewGroupRepository(db)

	err = repo.RejectMembership(groupID, userID, adminID)
	if err != nil {
		log.Println("Error rejecting member:", err)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	gname, _, _ := GetGroupNameAndCreator(db, groupID)
	websocket.SendNotification(userID, "rejection", "Your request to join group: "+gname+" has been rejected!")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Membership rejected"})
}
