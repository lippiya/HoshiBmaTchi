package gorm_user

import "time"

type NotificationModel struct {
	ID          string    `gorm:"primaryKey;type:uuid" json:"id"`
	RecipientID string    `gorm:"type:uuid;not null;index" json:"recipient_id"`
	ActorID     string    `gorm:"type:uuid;not null;index" json:"actor_id"`
	Type        string    `gorm:"type:varchar(50);not null" json:"type"` 
	PostID      *string   `gorm:"type:uuid;index" json:"post_id,omitempty"`
	CommentText *string   `gorm:"type:text" json:"comment_text,omitempty"`
	IsRead      bool      `gorm:"type:boolean;default:false" json:"is_read"`
	CreatedAt   time.Time `gorm:"type:timestamp;not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"type:timestamp;not null;default:now()" json:"updated_at"`
}

func (NotificationModel) TableName() string {
	return "notification_models"
}

type NotificationResponse struct {
	ID            string        `json:"id"`
	Type          string        `json:"type"`
	User          UserResponse  `json:"user"`
	PostID        *string       `json:"post_id,omitempty"`
	PostThumbnail *string       `json:"post_thumbnail,omitempty"`
	CommentText   *string       `json:"comment_text,omitempty"`
	IsRead        bool          `json:"is_read"`
	CreatedAt     time.Time     `json:"created_at"`
}
