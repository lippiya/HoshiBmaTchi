package gorm_post

import "time"

type PostModel struct {
	ID string      `json:"id" gorm:"primaryKey"`
	UserID string      `json:"user_id" gorm:"not null;index"`
	Caption        string      `json:"caption"`
	IsCommentOff   bool        `json:"is_comment_off" gorm:"default:false"`
	IsHideLikes    bool        `json:"is_hide_likes" gorm:"default:false"`
	IsReel         bool        `json:"is_reel" gorm:"default:false"`
	LikesCount     int         `json:"likes_count" gorm:"default:0"`
	CommentsCount  int         `json:"comments_count" gorm:"default:0"`
	Media          []PostMediaModel `json:"media" gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE"`
	Collaborators  []PostCollaboratorModel `json:"collaborators" gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE"`
	LikedByUsers   string      `json:"liked_by_users" gorm:"type:text"`
	SavedByUsers   string      `json:"saved_by_users" gorm:"type:text"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

func (PostModel) TableName() string {
    return "post_models"
}

type CommentModel struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	PostID    string    `json:"post_id" gorm:"not null;index"`
	UserID    string    `json:"user_id" gorm:"not null;index"`
	Content   string    `json:"content" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PostMediaModel struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	PostID    string    `json:"post_id" gorm:"not null;index"`
	FilePath  string    `json:"file_path" gorm:"not null"`
	MediaType string    `json:"media_type" gorm:"not null"`
	OrderNum  int       `json:"order_num" gorm:"default:0"` 
	Duration  int       `json:"duration,omitempty"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (PostMediaModel) TableName() string {
	return "post_media_models"
}

func (PostCollaboratorModel) TableName() string {
	return "post_collaborator_models"
}

type PostCollaboratorModel struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	PostID    string    `json:"post_id" gorm:"not null;index"`
	UserID    string    `json:"user_id" gorm:"not null;index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}


type CreatePostMediaRequest struct{
    PostID string
    FilePath string
    MediaType string
    OrderNum int
    Duration int
    Width int
    Height int
    CreatedAt time.Time
	UpdatedAt time.Time
}

type CreatePostCollaboratorRequest struct{
    PostID string
    UserID string
    CreatedAt time.Time
    UpdatedAt time.Time
}

type CreatePostRequest struct{
    UserID string
    Caption string
    IsCommentOff bool
    IsHideLikes bool
    IsReel bool
    LikesCount int
	CommentsCount int
    Media []CreatePostMediaRequest
    Collaborators []CreatePostCollaboratorRequest
    CreatedAt time.Time
    UpdatedAt time.Time
}

type CollectionModel struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"not null;index"`
	Name      string    `json:"name" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (CollectionModel) TableName() string {
	return "collection_models"
}

type SavedPostModel struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	CollectionID string    `json:"collection_id" gorm:"not null;index"`
	PostID       string    `json:"post_id" gorm:"not null;index"`
	CreatedAt    time.Time `json:"created_at"`
}

func (SavedPostModel) TableName() string {
	return "saved_post_models"
}
