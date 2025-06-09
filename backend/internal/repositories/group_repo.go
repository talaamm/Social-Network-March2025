package repositories

import (
	"database/sql"
	"errors"
	"fmt"

	"social-network/internal/models"
)

// GroupRepository handles database operations related to groups
type GroupRepository struct {
	DB *sql.DB
}

// NewGroupRepository creates a new instance of GroupRepository
func NewGroupRepository(db *sql.DB) *GroupRepository {
	return &GroupRepository{DB: db}
}

func (repo *GroupRepository) DeleteGroup(groupID, userID int) error {
	// Check if the user is the creator of the group
	var creatorID int
	err := repo.DB.QueryRow(`SELECT creator_id FROM groups WHERE id = ?`, groupID).Scan(&creatorID)
	if err != nil {
		return fmt.Errorf("group not found or error retrieving group: %v", err)
	}

	// Ensure only the creator can delete the group
	if creatorID != userID {
		return fmt.Errorf("unauthorized: only the group creator can delete this group")
	}

	// Delete the group (all related records will be removed due to ON DELETE CASCADE)
	_, err = repo.DB.Exec(`DELETE FROM groups WHERE id = ?`, groupID)
	if err != nil {
		return fmt.Errorf("failed to delete group: %v", err)
	}

	return nil
}

func (repo *GroupRepository) RequestToJoinGroup(groupID, userID int, nickanme string) error {
	// Check if the user is already a member or pending approval
	var exists bool
	err := repo.DB.QueryRow("SELECT EXISTS (SELECT 1 FROM group_members WHERE id = ? AND group_id = ?)", userID, groupID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("user is already in the group or has a pending request")
	}

	// Insert the membership request with status "pending"
	_, err = repo.DB.Exec(`
        INSERT INTO group_members (group_id, id, status, username) 
        VALUES (?, ?, 'pending', ?)`,
		groupID, userID, nickanme)

	return err
}

func (repo *GroupRepository) CreateGroup(group *models.Group) (int, error) {
	var groupID int

	// ✅ SQLite & MySQL: Use LastInsertId
	result, err := repo.DB.Exec(`
        INSERT INTO groups (group_name, description, creator_id) 
        VALUES (?, ?, ?)`,
		group.Name, group.Description, group.CreatorID)
	if err != nil {
		return 0, err
	}

	// ✅ Get the last inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	groupID = int(id)
	return groupID, nil
}

func (repo *GroupRepository) GetGroupMembers(groupID int) ([]models.GroupMember, error) {
	rows, err := repo.DB.Query(`
        SELECT gm.id, u.nickname, gm.status
        FROM group_members gm
        JOIN users u ON gm.id = u.id
        WHERE gm.group_id = ? AND (gm.status = 'approved')`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.GroupMember
	for rows.Next() {
		var member models.GroupMember
		err := rows.Scan(&member.ID, &member.Nickname, &member.Status)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return members, nil
}

// IsUserInGroup checks if a user is already in a group
func (repo *GroupRepository) IsUserInGroup(groupID, userID int) bool {
	var exists bool
	err := repo.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = ? AND id = ? AND status = 'approved')`, groupID, userID).Scan(&exists)
	return err == nil && exists
}

func (repo *GroupRepository) IsInvitationPending(groupID, invitedUserID int) bool {
	var exists bool
	err := repo.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM group_invitations WHERE group_id = ? AND invited_user_id = ?)`, groupID, invitedUserID).Scan(&exists)
	return err == nil && exists
}

func (repo *GroupRepository) InviteUserToGroup(groupID, memberID, invitedUserID int) error {
	_, err := repo.DB.Exec(`INSERT INTO group_invitations (group_id, member_id, invited_user_id) VALUES (?, ?, ?)`, groupID, memberID, invitedUserID)
	return err
}

func (repo *GroupRepository) ApproveMembership(groupID, userID, adminID int) error {
	isAdmin, err := repo.IsUserGroupAdmin(adminID, groupID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New("only the group admin can approve members")
	}

	_, err = repo.DB.Exec("UPDATE group_members SET status = 'approved' WHERE group_id = ? AND id = ? AND status = 'pending'", groupID, userID)
	return err
}

// RejectMembership removes the user from pending requests
func (repo *GroupRepository) RejectMembership(groupID, userID, adminID int) error {
	isAdmin, err := repo.IsUserGroupAdmin(adminID, groupID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New("only the group admin can reject members")
	}

	_, err = repo.DB.Exec("DELETE FROM group_members WHERE group_id = ? AND id = ? AND status = 'pending'", groupID, userID)
	return err
}

func (repo *GroupRepository) IsUserGroupAdmin(userID, groupID int) (bool, error) {
	var isAdmin bool
	err := repo.DB.QueryRow("SELECT EXISTS (SELECT 1 FROM groups WHERE id = ? AND creator_id = ?)", groupID, userID).Scan(&isAdmin)
	return isAdmin, err
}
