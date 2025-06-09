package repositories

import (
	"database/sql"
)

// GroupMemberRepository handles user-group relationships
type GroupMemberRepository struct {
	DB *sql.DB
}

// NewGroupMemberRepository creates a new instance of GroupMemberRepository
func NewGroupMemberRepository(db *sql.DB) *GroupMemberRepository {
	return &GroupMemberRepository{DB: db}
}

// RemoveUserFromGroup removes a user from a group
func (repo *GroupMemberRepository) RemoveUserFromGroup(groupID, userID int) error {
	_, err := repo.DB.Exec(`DELETE FROM group_members WHERE group_id = ? AND id = ?`, groupID, userID)
	return err
}

func (repo *GroupMemberRepository) AddUserToGroup(groupID, userID int, role, username string) error {
	_, err := repo.DB.Exec(`
		INSERT INTO group_members (group_id, id, status , username) 
		VALUES (?, ?, ? , ?)`, groupID, userID, role , username)
	return err
}
