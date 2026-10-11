package gorm_message

import (
	"context"
	"hoshiBmaTchi/back-end/message_service/internal/core/domain"
	"hoshiBmaTchi/back-end/message_service/internal/core/port"
	"gorm.io/gorm"
)

type messageRepository struct {
	db *gorm.DB
}

func NewGormMessageRepo(db *gorm.DB) port.MessageRepositoryPort {
	return &messageRepository{
		db: db,
	}
}

func (mr *messageRepository) CreateConversation(c context.Context, conversation *domain.Conversation) error {
	result := mr.db.WithContext(c).Create(conversation)
	return result.Error
}

func (mr *messageRepository) GetConversationByID(c context.Context, conversationID string) (*domain.Conversation, error) {
	var conversation domain.Conversation
	result := mr.db.WithContext(c).First(&conversation, "id = ?", conversationID)
	return &conversation, result.Error
}

func (mr *messageRepository) UpdateConversation(c context.Context, conversation *domain.Conversation) error {
	result := mr.db.WithContext(c).Save(conversation)
	return result.Error
}

func (mr *messageRepository) DeleteConversation(c context.Context, conversationID string) error {
	tx := mr.db.WithContext(c).Begin()

	if err := tx.Where("conversation_id = ?", conversationID).Delete(&domain.Message{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("conversation_id = ?", conversationID).Delete(&domain.ConversationParticipant{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&domain.Conversation{}, "id = ?", conversationID).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (mr *messageRepository) GetConversationsByUserID(c context.Context, userID string) ([]domain.Conversation, error) {
	var participants []domain.ConversationParticipant
	if err := mr.db.WithContext(c).Where("user_id = ?", userID).Find(&participants).Error; err != nil {
		return nil, err
	}

	var conversationIDs []string
	for _, p := range participants {
		conversationIDs = append(conversationIDs, p.ConversationID)
	}

	if len(conversationIDs) == 0 {
		return []domain.Conversation{}, nil
	}

	var conversations []domain.Conversation
	result := mr.db.WithContext(c).Where("id IN ?", conversationIDs).Order("updated_at DESC").Find(&conversations)
	return conversations, result.Error
}

func (mr *messageRepository) SearchConversations(c context.Context, userID string, query string) ([]domain.Conversation, error) {
	var participants []domain.ConversationParticipant
	if err := mr.db.WithContext(c).Where("user_id = ?", userID).Find(&participants).Error; err != nil {
		return nil, err
	}

	var conversationIDs []string
	for _, p := range participants {
		conversationIDs = append(conversationIDs, p.ConversationID)
	}

	if len(conversationIDs) == 0 {
		return []domain.Conversation{}, nil
	}

	var conversations []domain.Conversation
	result := mr.db.WithContext(c).
		Where("id IN ? AND (group_name ILIKE ? OR is_group = false)", conversationIDs, "%"+query+"%").
		Order("updated_at DESC").
		Find(&conversations)
	return conversations, result.Error
}

func (mr *messageRepository) CreateMessage(c context.Context, message *domain.Message) error {
	result := mr.db.WithContext(c).Create(message)
	return result.Error
}

func (mr *messageRepository) GetMessageByID(c context.Context, messageID string) (*domain.Message, error) {
	var message domain.Message
	result := mr.db.WithContext(c).First(&message, "id = ?", messageID)
	return &message, result.Error
}

func (mr *messageRepository) GetMessagesByConversationID(c context.Context, conversationID string) ([]domain.Message, error) {
	var messages []domain.Message
	result := mr.db.WithContext(c).Where("conversation_id = ?", conversationID).Order("created_at ASC").Find(&messages)
	return messages, result.Error
}

func (mr *messageRepository) DeleteMessage(c context.Context, messageID string) error {
	result := mr.db.WithContext(c).Delete(&domain.Message{}, "id = ?", messageID)
	return result.Error
}

func (mr *messageRepository) SearchMessagesInConversation(c context.Context, conversationID string, query string) ([]domain.Message, error) {
	var messages []domain.Message
	result := mr.db.WithContext(c).
		Where("conversation_id = ? AND content ILIKE ?", conversationID, "%"+query+"%").
		Order("created_at DESC").
		Find(&messages)
	return messages, result.Error
}

func (mr *messageRepository) UpdateMessageStatus(c context.Context, messageID string, status string) error {
	result := mr.db.WithContext(c).Model(&domain.Message{}).Where("id = ?", messageID).Update("status", status)
	return result.Error
}

func (mr *messageRepository) AddParticipant(c context.Context, participant *domain.ConversationParticipant) error {
	result := mr.db.WithContext(c).Create(participant)
	return result.Error
}

func (mr *messageRepository) RemoveParticipant(c context.Context, conversationID string, userID string) error {
	result := mr.db.WithContext(c).Where("conversation_id = ? AND user_id = ?", conversationID, userID).Delete(&domain.ConversationParticipant{})
	return result.Error
}

func (mr *messageRepository) GetParticipantsByConversationID(c context.Context, conversationID string) ([]domain.ConversationParticipant, error) {
	var participants []domain.ConversationParticipant
	result := mr.db.WithContext(c).Where("conversation_id = ?", conversationID).Find(&participants)
	return participants, result.Error
}

func (mr *messageRepository) GetParticipant(c context.Context, conversationID string, userID string) (*domain.ConversationParticipant, error) {
	var participant domain.ConversationParticipant
	result := mr.db.WithContext(c).Where("conversation_id = ? AND user_id = ?", conversationID, userID).First(&participant)
	return &participant, result.Error
}

func (mr *messageRepository) UpdateParticipant(c context.Context, participant *domain.ConversationParticipant) error {
	result := mr.db.WithContext(c).Save(participant)
	return result.Error
}

func (mr *messageRepository) UpdateUnreadCount(c context.Context, conversationID string, userID string, count int) error {
	result := mr.db.WithContext(c).
		Model(&domain.ConversationParticipant{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Update("unread_count", count)
	return result.Error
}

func (mr *messageRepository) ResetUnreadCount(c context.Context, conversationID string, userID string) error {
	result := mr.db.WithContext(c).
		Model(&domain.ConversationParticipant{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Update("unread_count", 0)
	return result.Error
}

func (mr *messageRepository) GetOrCreateDirectConversation(c context.Context, userID1 string, userID2 string) (*domain.Conversation, error) {
	var participants []domain.ConversationParticipant
	
	if err := mr.db.WithContext(c).Where("user_id IN ?", []string{userID1, userID2}).Find(&participants).Error; err != nil {
		return nil, err
	}

	conversationCounts := make(map[string]int)
	for _, p := range participants {
		conversationCounts[p.ConversationID]++
	}

	for convID, count := range conversationCounts {
		if count == 2 {
			var conv domain.Conversation
			if err := mr.db.WithContext(c).Where("id = ? AND is_group = false", convID).First(&conv).Error; err == nil {
				var allParticipants []domain.ConversationParticipant
				if err := mr.db.WithContext(c).Where("conversation_id = ?", convID).Find(&allParticipants).Error; err == nil {
					if len(allParticipants) == 2 {
						return &conv, nil
					}
				}
			}
		}
	}

	return nil, gorm.ErrRecordNotFound
}
