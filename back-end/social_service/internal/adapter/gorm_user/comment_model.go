package gorm_user

import (
	"time"
)

type CommentModel struct {
	ID              string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	PostID          string    `gorm:"type:uuid;index"`
	ParentCommentID *string   `gorm:"type:uuid;index"`
	UserID          string    `gorm:"type:uuid;not null;index"`
	Text            string    `gorm:"type:text"`
	GifUrl          string    `gorm:"type:varchar(500)"`
	LikesCount      int       `gorm:"default:0"`
	RepliesCount    int       `gorm:"default:0"`
	CreatedAt       time.Time `gorm:"type:timestamp;not null;default:now()"`
	UpdatedAt       time.Time `gorm:"type:timestamp;not null;default:now()"`
	DeletedAt       *time.Time `gorm:"index"`
}

func (CommentModel) TableName() string {
	return "comments"
}

type CommentLikeModel struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CommentID string    `gorm:"type:uuid;not null;index"`
	UserID    string    `gorm:"type:uuid;not null;index"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:now()"`
}

func (CommentLikeModel) TableName() string {
	return "comment_likes"
}
