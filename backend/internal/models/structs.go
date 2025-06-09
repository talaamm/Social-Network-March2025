package models

import "time"

type User struct {
	ID        int    `json:"id"`
	Nickname  string `json:"nickname"`
	Email     string `json:"email"`
	Password  string `json:"password,omitempty"` // Exclude from JSON output
	Age       int    `json:"age,omitempty"`      // ✅ Change to a pointer to handle NULL values
	Gender    string `json:"gender"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Birthdate string `json:"dbirth"`
	IsPrivate bool   `json:"isprivate"`
}

type Post struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	Nickname     string    `json:"nickname"`
	Content      string    `json:"content"`
	Image        *string   `json:"image"` // Nullable field
	Privacy      string    `json:"privacy"`
	CreatedAt    time.Time `json:"created_at"`
	LikeCount    int       `json:"likes"`
	DisLikeCount int       `json:"dislikes"`
}

type Notification struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	Type      string `json:"type"`
	Message   string `json:"message"`
	IsRead    bool   `json:"is_read"`
	CreatedAt string `json:"created_at"` // ✅ Ensure it's always included
}

type Message struct {
	ID         int    `json:"id"`
	SenderID   int    `json:"sender_id"`
	ReceiverID int    `json:"receiver_id"`
	Content    string `json:"content"`
	SentAt     string `json:"sent_at"`
}

type Follow struct {
	ID          int    `json:"id"`
	FollowerID  int    `json:"follower_id"`
	FollowingID int    `json:"following_id"`
	Status      string `json:"status"` // "pending" or "accepted"
}

type EventRSVP struct {
	ID        int    `json:"id"`
	EventID   int    `json:"event_id"`
	UserID    int    `json:"user_id"`
	Status    string `json:"status"` // "going" or "not going"
	CreatedAt string `json:"created_at"`
}

type Comment struct {
	ID       int    `json:"id"`
	PostID   int    `json:"post_id"`
	UserID   int    `json:"user_id"`
	Content  string `json:"content"`
	Image    string `json:"image"`
	Nickname string `json:"nickname"`
}

type ChatMessage struct {
	ID         int       `json:"id"`
	SenderID   int       `json:"sender_id"`
	ReceiverID *int      `json:"receiver_id,omitempty"` // Nullable for group messages
	GroupID    *int      `json:"group_id,omitempty"`    // Nullable for direct messages
	Content    string    `json:"content"`
	SentAt     time.Time `json:"sent_at"`
}

type Like struct {
	ID     int  `json:"id"`
	Postid int  `json:"postid"`
	IsLike bool `json:"islike"`
}

type PostVisibility struct {
	PostCreator int `json:"post_creator"` // The user who created the post
	UserID      int `json:"user_id"`      // The user allowed to see the post
}
