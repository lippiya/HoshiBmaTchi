package gorm_message

import "time"

type ConversationModel struct {
	ID            string                         `json:"id" gorm:"primaryKey"`
	IsGroup       bool                           `json:"is_group" gorm:"default:false"`
	GroupName     string                         `json:"group_name"`
	LastMessageID string                         `json:"last_message_id"`
	Participants  []ConversationParticipantModel `json:"participants" gorm:"foreignKey:ConversationID;constraint:OnDelete:CASCADE"`
	CreatedAt     time.Time                      `json:"created_at"`
	UpdatedAt     time.Time                      `json:"updated_at"`
}

func (ConversationModel) TableName() string {
	return "conversation_models"
}

type MessageModel struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	ConversationID string    `json:"conversation_id" gorm:"not null;index"`
	SenderID       string    `json:"sender_id" gorm:"not null;index"`
	Content        string    `json:"content" gorm:"type:text"`
	MediaURL       string    `json:"media_url"`
	MediaType      string    `json:"media_type"`
	Status         string    `json:"status" gorm:"default:'sent'"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (MessageModel) TableName() string {
	return "message_models"
}

type ConversationParticipantModel struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	ConversationID string    `json:"conversation_id" gorm:"not null;index"`
	UserID         string    `json:"user_id" gorm:"not null;index"`
	UnreadCount    int       `json:"unread_count" gorm:"default:0"`
	LastSeen       time.Time `json:"last_seen"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (ConversationParticipantModel) TableName() string {
	return "conversation_participant_models"
}
