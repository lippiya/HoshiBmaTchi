package service

import (
	"context"
	"encoding/json"
	"fmt"
	"hoshiBmaTchi/back-end/social_service/internal/adapter/gorm_user"
	"net/http"
	"os"
)

type NotificationService struct {
	repo *gorm_user.NotificationRepository
}

func NewNotificationService(repo *gorm_user.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (ns *NotificationService) GetNotifications(ctx context.Context, userID string, page, limit int) ([]gorm_user.NotificationResponse, int64, error) {
	notifications, total, err := ns.repo.GetNotifications(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]gorm_user.NotificationResponse, 0, len(notifications))
	for _, n := range notifications {
		userInfo, err := ns.getUserInfo(n.ActorID)
		if err != nil {
			userInfo = gorm_user.UserResponse{
				ID:       n.ActorID,
				Username: "Unknown",
			}
		}

		var postThumbnail *string
		if n.PostID != nil {
			thumbnail, err := ns.getPostThumbnail(*n.PostID)
			if err == nil && thumbnail != "" {
				postThumbnail = &thumbnail
			}
		}

		responses = append(responses, gorm_user.NotificationResponse{
			ID:            n.ID,
			Type:          n.Type,
			User:          userInfo,
			PostID:        n.PostID,
			PostThumbnail: postThumbnail,
			CommentText:   n.CommentText,
			IsRead:        n.IsRead,
			CreatedAt:     n.CreatedAt,
		})
	}

	return responses, total, nil
}

func (ns *NotificationService) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	return ns.repo.GetUnreadCount(ctx, userID)
}

func (ns *NotificationService) CreateNotification(ctx context.Context, recipientID, actorID, notificationType string, postID, commentText *string) error {
	if recipientID == actorID {
		return nil
	}

	exists, err := ns.repo.CheckDuplicateNotification(ctx, actorID, recipientID, notificationType, postID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	notification := &gorm_user.NotificationModel{
		RecipientID: recipientID,
		ActorID:     actorID,
		Type:        notificationType,
		PostID:      postID,
		CommentText: commentText,
		IsRead:      false,
	}

	if err := ns.repo.CreateNotification(ctx, notification); err != nil {
		return err
	}

	if err := ns.repo.DeleteOldNotifications(ctx, recipientID, 100); err != nil {
		return err
	}

	return nil
}

func (ns *NotificationService) MarkAsRead(ctx context.Context, notificationID, userID string) error {
	return ns.repo.MarkAsRead(ctx, notificationID, userID)
}

func (ns *NotificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	return ns.repo.MarkAllAsRead(ctx, userID)
}

func (ns *NotificationService) DeleteNotification(ctx context.Context, notificationID, userID string) error {
	return ns.repo.DeleteNotification(ctx, notificationID, userID)
}

func (ns *NotificationService) getUserInfo(userID string) (gorm_user.UserResponse, error) {
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

func (ns *NotificationService) getPostThumbnail(postID string) (string, error) {
	postServiceURL := os.Getenv("POST_SERVICE_URL")
	if postServiceURL == "" {
		postServiceURL = "http://post-service:8083"
	}

	resp, err := http.Get(fmt.Sprintf("%s/api/v1/posts/get-post-by-id/%s", postServiceURL, postID))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Post struct {
			MediaURLs []string `json:"media_urls"`
		} `json:"post"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Post.MediaURLs) > 0 {
		return result.Post.MediaURLs[0], nil
	}

	return "", nil
}
