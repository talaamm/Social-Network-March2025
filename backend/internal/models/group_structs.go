package models

import "time"

type Group struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatorID   int    `json:"creator_id"`
	CreatorName string `json:"creator_nickname"` // New field
	MemberCount int    `json:"member_count"`
}

type GroupMember struct {
	GroupID  int    `json:"group_id"`
	ID       int    `json:"user_id"`
	Nickname string `json:"nickname"`
	Status   string `json:"status"` // "pending", "approved", or "rejected"
}

type GroupComment struct {
	ID       int     `json:"id"`
	MemberID int     `json:"member_id"`
	GPostID  int     `json:"g_post_id"`
	Content  string  `json:"content"`
	Image    *string `json:"image,omitempty"` // Optional field
	Nickname string  `json:"nickname"`
	GroupID  int     `json:"group_id"`
}

type GroupPost struct {
	ID           int     `json:"id"`
	GroupID      int     `json:"group_id"`
	MemberID     int     `json:"member_id"`
	Content      string  `json:"content"`
	Image        *string `json:"image,omitempty"` // Optional field
	CreatedAt    string  `json:"created_at"`
	Nickname     string  `json:"nickname"`
	LikeCount    int     `json:"likes"`
	DisLikeCount int     `json:"dislikes"`
}

type GroupLike struct {
	ID       int  `json:"id"`
	PostID   int  `json:"post_id"`
	MemberID int  `json:"member_id"`
	IsLike   bool `json:"is_like"` // true for like, false for dislike
}

type GroupChatMessage struct {
	ID       int       `json:"id"`
	GroupID  int       `json:"group_id"`
	SenderID int       `json:"sender_id"`
	Content  string    `json:"content"`
	SentAt   time.Time `json:"sent_at"`
}

type GroupEvent struct {
	ID          int    `json:"id"`
	GroupID     int    `json:"group_id"`
	Creator     string `json:"creator"`
	CreatorID   int    `json:"creator_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	EventDate   string `json:"event_date"`
	CreatedAt   string `json:"created_at"`
	Going       int    `json:"going"`
	NotGoing    int    `json:"not_going"`
}
