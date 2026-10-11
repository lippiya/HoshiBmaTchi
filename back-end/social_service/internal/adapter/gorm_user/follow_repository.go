package gorm_user

import (
	"context"
	"hoshiBmaTchi/back-end/social_service/internal/core/domain"
	"hoshiBmaTchi/back-end/social_service/internal/core/port"

	"gorm.io/gorm"
)

type followRepository struct {
	db *gorm.DB
}

func NewFollowRepository(db *gorm.DB) port.FollowRepositoryPort {
	return &followRepository{
		db: db,
	}
}

func (fr *followRepository) CreateFollow(ctx context.Context, follow *domain.FollowModel) error {
	result := fr.db.WithContext(ctx).Create(follow)
	return result.Error
}

func (fr *followRepository) DeleteFollow(ctx context.Context, followerID, followingID string) error {
	result := fr.db.WithContext(ctx).Where("follower_id = ? AND following_id = ?", followerID, followingID).Delete(&domain.FollowModel{})
	return result.Error
}

func (fr *followRepository) IsFollowing(ctx context.Context, followerID, followingID string) (bool, error) {
	var count int64
	result := fr.db.WithContext(ctx).Model(&domain.FollowModel{}).Where("follower_id = ? AND following_id = ?", followerID, followingID).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

func (fr *followRepository) GetFollowersCount(ctx context.Context, userID string) (int, error) {
	var count int64
	result := fr.db.WithContext(ctx).Model(&domain.FollowModel{}).Where("following_id = ?", userID).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return int(count), nil
}

func (fr *followRepository) GetFollowingCount(ctx context.Context, userID string) (int, error) {
	var count int64
	result := fr.db.WithContext(ctx).Model(&domain.FollowModel{}).Where("follower_id = ?", userID).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return int(count), nil
}

func (fr *followRepository) GetFollowersList(ctx context.Context, userID string, page, limit int) ([]domain.UserResponse, int, error) {
	var follows []domain.FollowModel
	var count int64

	fr.db.WithContext(ctx).Model(&domain.FollowModel{}).Where("following_id = ?", userID).Count(&count)

	offset := (page - 1) * limit
	result := fr.db.WithContext(ctx).Where("following_id = ?", userID).Offset(offset).Limit(limit).Find(&follows)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return nil, int(count), nil
}

func (fr *followRepository) GetFollowingList(ctx context.Context, userID string, page, limit int) ([]domain.UserResponse, int, error) {
	var follows []domain.FollowModel
	var count int64

	fr.db.WithContext(ctx).Model(&domain.FollowModel{}).Where("follower_id = ?", userID).Count(&count)

	offset := (page - 1) * limit
	result := fr.db.WithContext(ctx).Where("follower_id = ?", userID).Offset(offset).Limit(limit).Find(&follows)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return nil, int(count), nil
}

func (fr *followRepository) GetFollowing(ctx context.Context, userID string) ([]domain.UserResponse, error) {
	var follows []domain.FollowModel
	result := fr.db.WithContext(ctx).Where("follower_id = ?", userID).Find(&follows)
	if result.Error != nil {
		return nil, result.Error
	}

	var userIDs []string
	for _, f := range follows {
		userIDs = append(userIDs, f.FollowingID)
	}

	var users []domain.UserResponse
	if len(userIDs) > 0 {
	
		for _, id := range userIDs {
			users = append(users, domain.UserResponse{ID: id})
		}
	}

	return users, nil
}
