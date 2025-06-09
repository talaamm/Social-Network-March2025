package repositories

import (
	"database/sql"
	"time"

	"social-network/internal/models"
)

// GroupPostRepository handles group post-related operations
type GroupPostRepository struct {
	DB *sql.DB
}

// NewGroupPostRepository creates a new instance of GroupPostRepository
func NewGroupPostRepository(db *sql.DB) *GroupPostRepository {
	return &GroupPostRepository{DB: db}
}

func (repo *GroupPostRepository) CreateGroupPost(post *models.GroupPost) (*models.GroupPost, error) {
	query := `
        INSERT INTO group_posts (group_id, member_id, content, created_at, image, username)
        VALUES (?, ?, ?, ?, ? , ?)
		RETURNING id, username, group_id, member_id, content , created_at , image`

	// post.GroupID, post.Nickname, post.GroupID, post.MemberID, post.Content, post.CreatedAt, post.Image)
	var newPost models.GroupPost
	err := repo.DB.QueryRow(query, post.GroupID, post.MemberID, post.Content, time.Now(), post.Image, post.Nickname).
		Scan(&newPost.GroupID, &newPost.Nickname, &newPost.GroupID, &newPost.MemberID, &newPost.Content, &newPost.CreatedAt, &newPost.Image)
	if err != nil {
		return nil, err
	}
	return &newPost, nil
}
