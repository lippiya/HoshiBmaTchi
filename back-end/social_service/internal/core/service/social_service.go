package service

import (
	"context"
	"encoding/json"
	"fmt"
	"hoshiBmaTchi/back-end/social_service/internal/adapter/gorm_user"
	"net/http"
	"os"
)

type SocialService struct {
	repo *gorm_user.SocialRepository
}

func NewSocialService(repo *gorm_user.SocialRepository) *SocialService {
	return &SocialService{repo: repo}
}

func (ss *SocialService) GetCloseFriends(ctx context.Context, userID string) ([]gorm_user.UserResponse, error) {
	closeFriends, err := ss.repo.GetCloseFriends(ctx, userID)
	if err != nil {
		return nil, err
	}

	var users []gorm_user.UserResponse
	for _, cf := range closeFriends {
		user, err := ss.getUserInfo(cf.FriendID)
		if err == nil {
			users = append(users, user)
		}
	}

	return users, nil
}

func (ss *SocialService) GetBlockedUsers(ctx context.Context, userID string) ([]gorm_user.UserResponse, error) {
	blockedUsers, err := ss.repo.GetBlockedUsers(ctx, userID)
	if err != nil {
		return nil, err
	}

	var users []gorm_user.UserResponse
	for _, bu := range blockedUsers {
		user, err := ss.getUserInfo(bu.BlockedID)
		if err == nil {
			users = append(users, user)
		}
	}

	return users, nil
}

func (ss *SocialService) GetHiddenStoryViewers(ctx context.Context, userID string) ([]gorm_user.UserResponse, error) {
	hiddenViewers, err := ss.repo.GetHiddenStoryViewers(ctx, userID)
	if err != nil {
		return nil, err
	}

	var users []gorm_user.UserResponse
	for _, hv := range hiddenViewers {
		user, err := ss.getUserInfo(hv.HiddenID)
		if err == nil {
			users = append(users, user)
		}
	}

	return users, nil
}

func (ss *SocialService) HideStoryViewer(ctx context.Context, userID, hiddenID string) error {
	return ss.repo.HideStoryViewer(ctx, userID, hiddenID)
}

func (ss *SocialService) UnhideStoryViewer(ctx context.Context, userID, hiddenID string) error {
	return ss.repo.UnhideStoryViewer(ctx, userID, hiddenID)
}

func (ss *SocialService) BlockUser(ctx context.Context, userID, blockedID string) error {
	return ss.repo.BlockUser(ctx, userID, blockedID)
}

func (ss *SocialService) UnblockUser(ctx context.Context, userID, blockedID string) error {
	return ss.repo.UnblockUser(ctx, userID, blockedID)
}

func (ss *SocialService) CheckBlockStatus(ctx context.Context, userID, targetID string) (bool, error) {
	blockedUsers, err := ss.repo.GetBlockedUsers(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, bu := range blockedUsers {
		if bu.BlockedID == targetID {
			return true, nil
		}
	}

	return false, nil
}

func (ss *SocialService) AddCloseFriend(ctx context.Context, userID, friendID string) error {
	return ss.repo.AddCloseFriend(ctx, userID, friendID)
}

func (ss *SocialService) RemoveCloseFriend(ctx context.Context, userID, friendID string) error {
	return ss.repo.RemoveCloseFriend(ctx, userID, friendID)
}

func (ss *SocialService) getUserInfo(userID string) (gorm_user.UserResponse, error) {
	userServiceURL := os.Getenv("USER_SERVICE_URL")
	if userServiceURL == "" {
		userServiceURL = "http://user-service:8080"
	}

	resp, err := http.Get(fmt.Sprintf("%s/api/v1/users/get-user-by-id/%s", userServiceURL, userID))
	if err != nil {
		return gorm_user.UserResponse{}, err
	}
	defer resp.Body.Close()

	var result struct {
		User gorm_user.UserResponse `json:"user"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return gorm_user.UserResponse{}, err
	}

	return result.User, nil
}
