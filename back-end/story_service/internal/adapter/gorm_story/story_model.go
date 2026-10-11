package gorm_story

import (
	"gorm.io/gorm"
	"time"
)

type StoryModel struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID    string    `gorm:"type:uuid;not null;index"`
	MediaURL  string    `gorm:"type:text;not null"`
	MediaType string    `gorm:"type:varchar(50);not null"`
	Duration  int       `gorm:"type:int;default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	ExpiresAt time.Time `gorm:"index"`
}

func (StoryModel) TableName() string {
	return "stories"
}

type StoryLikeModel struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	StoryID   string    `gorm:"type:uuid;not null;index"`
	UserID    string    `gorm:"type:uuid;not null;index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (StoryLikeModel) TableName() string {
	return "story_likes"
}

type StoryReplyModel struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	StoryID   string    `gorm:"type:uuid;not null;index"`
	UserID    string    `gorm:"type:uuid;not null;index"`
	Reply     string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (StoryReplyModel) TableName() string {
	return "story_replies"
}

type StoryViewModel struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	StoryID   string    `gorm:"type:uuid;not null;index"`
	UserID    string    `gorm:"type:uuid;not null;index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (StoryViewModel) TableName() string {
	return "story_views"
}

type StoryShareModel struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	StoryID     string    `gorm:"type:uuid;not null;index"`
	SenderID    string    `gorm:"type:uuid;not null;index"`
	RecipientID string    `gorm:"type:uuid;not null;index"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}

func (StoryShareModel) TableName() string {
	return "story_shares"
}

func MigrateStoryTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&StoryModel{},
		&StoryLikeModel{},
		&StoryReplyModel{},
		&StoryViewModel{},
		&StoryShareModel{},
	)
}
