package domain

import "time"

type Conversation struct {
	ID           string
	IsGroup      bool
	GroupName    string
	LastMessageID string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Message struct {
	ID             string
	ConversationID string
	SenderID       string
	Content        string
	MediaURL       string
	MediaType      string
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ConversationParticipant struct {
	ID             string
	ConversationID string
	UserID         string
	UnreadCount    int
	LastSeen       time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (Conversation) TableName() string {
	return "conversation_models"
}

func (Message) TableName() string {
	return "message_models"
}

func (ConversationParticipant) TableName() string {
	return "conversation_participant_models"
}
