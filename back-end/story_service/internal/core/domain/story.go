package domain

import "time"

type Story struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	MediaURL  string    `json:"media_url"`
	MediaType string    `json:"media_type"`
	Duration  int       `json:"duration"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type StoryLike struct {
	ID        string    `json:"id"`
	StoryID   string    `json:"story_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type StoryReply struct {
	ID        string    `json:"id"`
	StoryID   string    `json:"story_id"`
	UserID    string    `json:"user_id"`
	Reply     string    `json:"reply"`
	CreatedAt time.Time `json:"created_at"`
}

type StoryView struct {
	ID        string    `json:"id"`
	StoryID   string    `json:"story_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type StoryShare struct {
	ID          string    `json:"id"`
	StoryID     string    `json:"story_id"`
	SenderID    string    `json:"sender_id"`
	RecipientID string    `json:"recipient_id"`
	CreatedAt   time.Time `json:"created_at"`
}
