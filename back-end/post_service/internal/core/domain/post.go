package domain

import "time"

type Post struct {
	ID            string
	UserID        string
	Caption       string
	IsCommentOff  bool
	IsHideLikes   bool
	IsReel        bool
	LikesCount    int
	CommentsCount int
	Media         []PostMedia
	Collaborators []PostCollaborator
	LikedByUsers  string
	SavedByUsers  string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type PostMedia struct {
	ID        string
	PostID    string
	FilePath  string
	MediaType string
	OrderNum  int
	Duration  int
	Width     int
	Height    int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PostCollaborator struct {
	ID        string
	PostID    string
	UserID    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Post) TableName() string {
	return "post_models"
}

func (PostMedia) TableName() string {
	return "post_media_models"
}

func (PostCollaborator) TableName() string {
	return "post_collaborator_models"
}

type Collection struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SavedPost struct {
	ID           string    `json:"id"`
	CollectionID string    `json:"collection_id"`
	PostID       string    `json:"post_id"`
	CreatedAt    time.Time `json:"created_at"`
}

func (Collection) TableName() string {
	return "collection_models"
}

func (SavedPost) TableName() string {
	return "saved_post_models"
}
