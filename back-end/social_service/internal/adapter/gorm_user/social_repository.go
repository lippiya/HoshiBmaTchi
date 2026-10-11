package gorm_user

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SocialRepository struct {
	db *gorm.DB
}

func NewSocialRepository(db *gorm.DB) *SocialRepository {
	return &SocialRepository{db: db}
}

func (sr *SocialRepository) GetCloseFriends(ctx context.Context, userID string) ([]CloseFriendModel, error) {
	var closeFriends []CloseFriendModel
	result := sr.db.WithContext(ctx).Where("user_id = ?", userID).Find(&closeFriends)
	return closeFriends, result.Error
}

func (sr *SocialRepository) AddCloseFriend(ctx context.Context, userID, friendID string) error {
	closeFriend := CloseFriendModel{
		ID:       uuid.New().String(),
		UserID:   userID,
		FriendID: friendID,
	}
	return sr.db.WithContext(ctx).Create(&closeFriend).Error
}

func (sr *SocialRepository) RemoveCloseFriend(ctx context.Context, userID, friendID string) error {
	return sr.db.WithContext(ctx).Where("user_id = ? AND friend_id = ?", userID, friendID).Delete(&CloseFriendModel{}).Error
}

func (sr *SocialRepository) GetBlockedUsers(ctx context.Context, userID string) ([]BlockedUserModel, error) {
	var blockedUsers []BlockedUserModel
	result := sr.db.WithContext(ctx).Where("user_id = ?", userID).Find(&blockedUsers)
	return blockedUsers, result.Error
}

func (sr *SocialRepository) BlockUser(ctx context.Context, userID, blockedID string) error {
	blockedUser := BlockedUserModel{
		ID:        uuid.New().String(),
		UserID:    userID,
		BlockedID: blockedID,
	}
	return sr.db.WithContext(ctx).Create(&blockedUser).Error
}

func (sr *SocialRepository) UnblockUser(ctx context.Context, userID, blockedID string) error {
	return sr.db.WithContext(ctx).Where("user_id = ? AND blocked_id = ?", userID, blockedID).Delete(&BlockedUserModel{}).Error
}

func (sr *SocialRepository) GetHiddenStoryViewers(ctx context.Context, userID string) ([]HiddenStoryViewerModel, error) {
	var hiddenViewers []HiddenStoryViewerModel
	result := sr.db.WithContext(ctx).Where("user_id = ?", userID).Find(&hiddenViewers)
	return hiddenViewers, result.Error
}

func (sr *SocialRepository) HideStoryViewer(ctx context.Context, userID, hiddenID string) error {
	hiddenViewer := HiddenStoryViewerModel{
		ID:       uuid.New().String(),
		UserID:   userID,
		HiddenID: hiddenID,
	}
	return sr.db.WithContext(ctx).Create(&hiddenViewer).Error
}

func (sr *SocialRepository) UnhideStoryViewer(ctx context.Context, userID, hiddenID string) error {
	return sr.db.WithContext(ctx).Where("user_id = ? AND hidden_id = ?", userID, hiddenID).Delete(&HiddenStoryViewerModel{}).Error
}
