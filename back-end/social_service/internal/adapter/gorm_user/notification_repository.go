package gorm_user

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (nr *NotificationRepository) GetNotifications(ctx context.Context, userID string, page, limit int) ([]NotificationModel, int64, error) {
	var notifications []NotificationModel
	var total int64

	offset := (page - 1) * limit

	if err := nr.db.WithContext(ctx).Model(&NotificationModel{}).Where("recipient_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	result := nr.db.WithContext(ctx).
		Where("recipient_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&notifications)

	return notifications, total, result.Error
}

func (nr *NotificationRepository) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	var count int64
	result := nr.db.WithContext(ctx).
		Model(&NotificationModel{}).
		Where("recipient_id = ? AND is_read = ?", userID, false).
		Count(&count)
	return count, result.Error
}

func (nr *NotificationRepository) CreateNotification(ctx context.Context, notification *NotificationModel) error {
	notification.ID = uuid.New().String()
	return nr.db.WithContext(ctx).Create(notification).Error
}

func (nr *NotificationRepository) MarkAsRead(ctx context.Context, notificationID, userID string) error {
	return nr.db.WithContext(ctx).
		Model(&NotificationModel{}).
		Where("id = ? AND recipient_id = ?", notificationID, userID).
		Update("is_read", true).Error
}

func (nr *NotificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	return nr.db.WithContext(ctx).
		Model(&NotificationModel{}).
		Where("recipient_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}

func (nr *NotificationRepository) DeleteNotification(ctx context.Context, notificationID, userID string) error {
	return nr.db.WithContext(ctx).
		Where("id = ? AND recipient_id = ?", notificationID, userID).
		Delete(&NotificationModel{}).Error
}

func (nr *NotificationRepository) DeleteOldNotifications(ctx context.Context, userID string, keepCount int) error {
	subQuery := nr.db.WithContext(ctx).
		Model(&NotificationModel{}).
		Select("id").
		Where("recipient_id = ?", userID).
		Order("created_at DESC").
		Limit(keepCount)

	return nr.db.WithContext(ctx).
		Where("recipient_id = ? AND id NOT IN (?)", userID, subQuery).
		Delete(&NotificationModel{}).Error
}

func (nr *NotificationRepository) GetNotificationByID(ctx context.Context, notificationID string) (*NotificationModel, error) {
	var notification NotificationModel
	result := nr.db.WithContext(ctx).Where("id = ?", notificationID).First(&notification)
	if result.Error != nil {
		return nil, result.Error
	}
	return &notification, nil
}

func (nr *NotificationRepository) CheckDuplicateNotification(ctx context.Context, actorID, recipientID, notificationType string, postID *string) (bool, error) {
	var count int64
	query := nr.db.WithContext(ctx).
		Model(&NotificationModel{}).
		Where("actor_id = ? AND recipient_id = ? AND type = ?", actorID, recipientID, notificationType)

	if postID != nil {
		query = query.Where("post_id = ?", *postID)
	} else {
		query = query.Where("post_id IS NULL")
	}

	result := query.Count(&count)
	return count > 0, result.Error
}
