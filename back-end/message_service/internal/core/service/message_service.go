package service

import (
	"context"
	"hoshiBmaTchi/back-end/message_service/internal/core/domain"
	"hoshiBmaTchi/back-end/message_service/internal/core/port"
	"time"
	"github.com/google/uuid"
)

type MessageServicePort struct {
	MessageRepository port.MessageRepositoryPort
}

func NewMessageService(messageRepo port.MessageRepositoryPort) port.MessageServicePort {
	return &MessageServicePort{
		MessageRepository: messageRepo,
	}
}

func (ms *MessageServicePort) CreateConversation(c context.Context, conversation *domain.Conversation, participantIDs []string) error {
	conversation.ID = uuid.New().String()
	conversation.CreatedAt = time.Now()
	conversation.UpdatedAt = time.Now()

	if err := ms.MessageRepository.CreateConversation(c, conversation); err != nil {
		return err
	}

	for _, userID := range participantIDs {
		participant := &domain.ConversationParticipant{
			ID:             uuid.New().String(),
			ConversationID: conversation.ID,
			UserID:         userID,
			UnreadCount:    0,
			LastSeen:       time.Now(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		if err := ms.MessageRepository.AddParticipant(c, participant); err != nil {
			return err
		}
	}

	return nil
}

func (ms *MessageServicePort) GetConversationByID(c context.Context, conversationID string) (*domain.Conversation, error) {
	return ms.MessageRepository.GetConversationByID(c, conversationID)
}

func (ms *MessageServicePort) UpdateConversation(c context.Context, conversation *domain.Conversation) error {
	conversation.UpdatedAt = time.Now()
	return ms.MessageRepository.UpdateConversation(c, conversation)
}

func (ms *MessageServicePort) DeleteConversation(c context.Context, conversationID string) error {
	return ms.MessageRepository.DeleteConversation(c, conversationID)
}

func (ms *MessageServicePort) GetConversationsByUserID(c context.Context, userID string) ([]domain.Conversation, error) {
	return ms.MessageRepository.GetConversationsByUserID(c, userID)
}

func (ms *MessageServicePort) SearchConversations(c context.Context, userID string, query string) ([]domain.Conversation, error) {
	return ms.MessageRepository.SearchConversations(c, userID, query)
}

func (ms *MessageServicePort) SendMessage(c context.Context, message *domain.Message) error {
	message.ID = uuid.New().String()
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()
	message.Status = "sent"

	if err := ms.MessageRepository.CreateMessage(c, message); err != nil {
		return err
	}

	conversation, err := ms.MessageRepository.GetConversationByID(c, message.ConversationID)
	if err != nil {
		return err
	}

	conversation.LastMessageID = message.ID
	conversation.UpdatedAt = time.Now()
	if err := ms.MessageRepository.UpdateConversation(c, conversation); err != nil {
		return err
	}

	participants, err := ms.MessageRepository.GetParticipantsByConversationID(c, message.ConversationID)
	if err != nil {
		return err
	}

	for _, participant := range participants {
		if participant.UserID != message.SenderID {
			participant.UnreadCount++
			participant.UpdatedAt = time.Now()
			if err := ms.MessageRepository.UpdateParticipant(c, &participant); err != nil {
				return err
			}
		}
	}

	return nil
}

func (ms *MessageServicePort) GetMessagesByConversationID(c context.Context, conversationID string) ([]domain.Message, error) {
	return ms.MessageRepository.GetMessagesByConversationID(c, conversationID)
}

func (ms *MessageServicePort) UnsendMessage(c context.Context, conversationID string, messageID string, userID string) error {
	message, err := ms.MessageRepository.GetMessageByID(c, messageID)
	if err != nil {
		return err
	}

	if message.SenderID != userID {
		return nil
	}

	if time.Since(message.CreatedAt) > time.Minute {
		return nil
	}

	return ms.MessageRepository.DeleteMessage(c, messageID)
}

func (ms *MessageServicePort) SearchMessagesInConversation(c context.Context, conversationID string, query string) ([]domain.Message, error) {
	return ms.MessageRepository.SearchMessagesInConversation(c, conversationID, query)
}

func (ms *MessageServicePort) MarkAsSeen(c context.Context, conversationID string, userID string) error {
	if err := ms.MessageRepository.ResetUnreadCount(c, conversationID, userID); err != nil {
		return err
	}

	messages, err := ms.MessageRepository.GetMessagesByConversationID(c, conversationID)
	if err != nil {
		return err
	}

	for _, msg := range messages {
		if msg.SenderID != userID && msg.Status != "seen" {
			if err := ms.MessageRepository.UpdateMessageStatus(c, msg.ID, "seen"); err != nil {
				return err
			}
		}
	}

	participant, err := ms.MessageRepository.GetParticipant(c, conversationID, userID)
	if err != nil {
		return err
	}

	participant.LastSeen = time.Now()
	participant.UpdatedAt = time.Now()
	return ms.MessageRepository.UpdateParticipant(c, participant)
}

func (ms *MessageServicePort) AddParticipants(c context.Context, conversationID string, participantIDs []string) error {
	for _, userID := range participantIDs {
		participant := &domain.ConversationParticipant{
			ID:             uuid.New().String(),
			ConversationID: conversationID,
			UserID:         userID,
			UnreadCount:    0,
			LastSeen:       time.Now(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		if err := ms.MessageRepository.AddParticipant(c, participant); err != nil {
			return err
		}
	}

	conversation, err := ms.MessageRepository.GetConversationByID(c, conversationID)
	if err != nil {
		return err
	}

	conversation.UpdatedAt = time.Now()
	return ms.MessageRepository.UpdateConversation(c, conversation)
}

func (ms *MessageServicePort) RemoveParticipant(c context.Context, conversationID string, userID string) error {
	return ms.MessageRepository.RemoveParticipant(c, conversationID, userID)
}

func (ms *MessageServicePort) GetParticipantsByConversationID(c context.Context, conversationID string) ([]domain.ConversationParticipant, error) {
	return ms.MessageRepository.GetParticipantsByConversationID(c, conversationID)
}

func (ms *MessageServicePort) GetOrCreateDirectConversation(c context.Context, userID1 string, userID2 string) (*domain.Conversation, error) {
	conversation, err := ms.MessageRepository.GetOrCreateDirectConversation(c, userID1, userID2)
	if err == nil {
		return conversation, nil
	}

	newConversation := &domain.Conversation{
		IsGroup:   false,
		GroupName: "",
	}

	if err := ms.CreateConversation(c, newConversation, []string{userID1, userID2}); err != nil {
		return nil, err
	}

	return newConversation, nil
}
