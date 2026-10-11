package service

import (
	"context"
	"errors"
	"hoshiBmaTchi/back-end/social_service/internal/core/domain"
	"hoshiBmaTchi/back-end/social_service/internal/core/port"

	"github.com/google/uuid"
)

type FollowService struct {
	followRepo port.FollowRepositoryPort
}

func NewFollowService(followRepo port.FollowRepositoryPort) port.FollowServicePort {
	return &FollowService{
		followRepo: followRepo,
	}
}

func (fs *FollowService) FollowUser(ctx context.Context, followerID, followingID string) error {
	if followerID == followingID {
		return errors.New("cannot follow yourself")
	}

	isFollowing, err := fs.followRepo.IsFollowing(ctx, followerID, followingID)
	if err != nil {
		return err
	}

	if isFollowing {
		return errors.New("already following this user")
	}

	follow := &domain.FollowModel{
		ID:          uuid.New().String(),
		FollowerID:  followerID,
		FollowingID: followingID,
	}

	return fs.followRepo.CreateFollow(ctx, follow)
}

func (fs *FollowService) UnfollowUser(ctx context.Context, followerID, followingID string) error {
	if followerID == followingID {
		return errors.New("cannot unfollow yourself")
	}

	isFollowing, err := fs.followRepo.IsFollowing(ctx, followerID, followingID)
	if err != nil {
		return err
	}

	if !isFollowing {
		return errors.New("not following this user")
	}

	return fs.followRepo.DeleteFollow(ctx, followerID, followingID)
}

func (fs *FollowService) IsFollowing(ctx context.Context, followerID, followingID string) (bool, error) {
	return fs.followRepo.IsFollowing(ctx, followerID, followingID)
}

func (fs *FollowService) GetFollowersCount(ctx context.Context, userID string) (int, error) {
	return fs.followRepo.GetFollowersCount(ctx, userID)
}

func (fs *FollowService) GetFollowingCount(ctx context.Context, userID string) (int, error) {
	return fs.followRepo.GetFollowingCount(ctx, userID)
}

func (fs *FollowService) GetFollowersList(ctx context.Context, userID string, page, limit int) ([]domain.UserResponse, int, error) {
	return fs.followRepo.GetFollowersList(ctx, userID, page, limit)
}

func (fs *FollowService) GetFollowingList(ctx context.Context, userID string, page, limit int) ([]domain.UserResponse, int, error) {
	return fs.followRepo.GetFollowingList(ctx, userID, page, limit)
}

func (fs *FollowService) GetFollowing(ctx context.Context, userID string) ([]domain.UserResponse, error) {
	return fs.followRepo.GetFollowing(ctx, userID)
}
