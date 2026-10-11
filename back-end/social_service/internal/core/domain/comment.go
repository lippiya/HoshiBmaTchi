package domain

import (
	"time"
)

type Comment struct {
	ID              string
	PostID          string
	ParentCommentID *string
	UserID          string
	Text            string
	GifUrl          string
	LikesCount      int
	RepliesCount    int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type CommentLike struct {
	ID        string
	CommentID string
	UserID    string
	CreatedAt time.Time
}

type CreateCommentRequest struct {
	PostID          string
	ParentCommentID *string
	UserID          string
	Text            string
	GifUrl          string
}

type CommentResponse struct {
	ID              string
	PostID          string
	ParentCommentID *string
	UserID          string
	Text            string
	GifUrl          string
	LikesCount      int
	RepliesCount    int
	CreatedAt       time.Time
	User            interface{}
}

type GetCommentsRequest struct {
	PostID string
	Page   int
	Limit  int
	Sort   string
}

type GetCommentRepliesRequest struct {
	CommentID string
}
