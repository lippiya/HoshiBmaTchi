package port

import (
	"context"
	"hoshiBmaTchi/back-end/message_service/internal/core/domain"
)

type MessageRepositoryPort interface {
	CreateConversation(c context.Context, conversation *domain.Conversation) error
	GetConversationByID(c context.Context, conversationID string) (*domain.Conversation, error)
	UpdateConversation(c context.Context, conversation *domain.Conversation) error
	DeleteConversation(c context.Context, conversationID string) error
	GetConversationsByUserID(c context.Context, userID string) ([]domain.Conversation, error)
	SearchConversations(c context.Context, userID string, query string) ([]domain.Conversation, error)

	CreateMessage(c context.Context, message *domain.Message) error
	GetMessageByID(c context.Context, messageID string) (*domain.Message, error)
	GetMessagesByConversationID(c context.Context, conversationID string) ([]domain.Message, error)
	DeleteMessage(c context.Context, messageID string) error
	SearchMessagesInConversation(c context.Context, conversationID string, query string) ([]domain.Message, error)
	UpdateMessageStatus(c context.Context, messageID string, status string) error

	AddParticipant(c context.Context, participant *domain.ConversationParticipant) error
	RemoveParticipant(c context.Context, conversationID string, userID string) error
	GetParticipantsByConversationID(c context.Context, conversationID string) ([]domain.ConversationParticipant, error)
	GetParticipant(c context.Context, conversationID string, userID string) (*domain.ConversationParticipant, error)
	UpdateParticipant(c context.Context, participant *domain.ConversationParticipant) error
	UpdateUnreadCount(c context.Context, conversationID string, userID string, count int) error
	ResetUnreadCount(c context.Context, conversationID string, userID string) error

	GetOrCreateDirectConversation(c context.Context, userID1 string, userID2 string) (*domain.Conversation, error)
}

type MessageServicePort interface {
	CreateConversation(c context.Context, conversation *domain.Conversation, participantIDs []string) error
	GetConversationByID(c context.Context, conversationID string) (*domain.Conversation, error)
	UpdateConversation(c context.Context, conversation *domain.Conversation) error
	DeleteConversation(c context.Context, conversationID string) error
	GetConversationsByUserID(c context.Context, userID string) ([]domain.Conversation, error)
	SearchConversations(c context.Context, userID string, query string) ([]domain.Conversation, error)

	SendMessage(c context.Context, message *domain.Message) error
	GetMessagesByConversationID(c context.Context, conversationID string) ([]domain.Message, error)
	UnsendMessage(c context.Context, conversationID string, messageID string, userID string) error
	SearchMessagesInConversation(c context.Context, conversationID string, query string) ([]domain.Message, error)
	MarkAsSeen(c context.Context, conversationID string, userID string) error

	AddParticipants(c context.Context, conversationID string, participantIDs []string) error
	RemoveParticipant(c context.Context, conversationID string, userID string) error
	GetParticipantsByConversationID(c context.Context, conversationID string) ([]domain.ConversationParticipant, error)

	GetOrCreateDirectConversation(c context.Context, userID1 string, userID2 string) (*domain.Conversation, error)
}

type PhotoClientPort interface {
	UploadPhoto(file interface{}, userID string) (string, error)
	DeletePhoto(filePath string) error
	GetPhotoURL(filePath string) (string, error)
}

type UserClientPort interface {
	GetUserByID(userID string) (interface{}, error)
	SearchUsers(query string) ([]interface{}, error)
}
