package port

import (
	"context"
	"hoshiBmaTchi/back-end/social_service/internal/core/domain"
)

type FollowRepositoryPort interface {
	CreateFollow(ctx context.Context, follow *domain.FollowModel) error
	DeleteFollow(ctx context.Context, followerID, followingID string) error
	IsFollowing(ctx context.Context, followerID, followingID string) (bool, error)
	GetFollowersCount(ctx context.Context, userID string) (int, error)
	GetFollowingCount(ctx context.Context, userID string) (int, error)
	GetFollowersList(ctx context.Context, userID string, page, limit int) ([]domain.UserResponse, int, error)
	GetFollowingList(ctx context.Context, userID string, page, limit int) ([]domain.UserResponse, int, error)
	GetFollowing(ctx context.Context, userID string) ([]domain.UserResponse, error)
}

type FollowServicePort interface {
	FollowUser(ctx context.Context, followerID, followingID string) error
	UnfollowUser(ctx context.Context, followerID, followingID string) error
	IsFollowing(ctx context.Context, followerID, followingID string) (bool, error)
	GetFollowersCount(ctx context.Context, userID string) (int, error)
	GetFollowingCount(ctx context.Context, userID string) (int, error)
	GetFollowersList(ctx context.Context, userID string, page, limit int) ([]domain.UserResponse, int, error)
	GetFollowingList(ctx context.Context, userID string, page, limit int) ([]domain.UserResponse, int, error)
	GetFollowing(ctx context.Context, userID string) ([]domain.UserResponse, error)
}

type CommentRepositoryPort interface {
	CreateComment(req domain.CreateCommentRequest) (*domain.Comment, error)
	GetCommentsByPostID(postID string, page, limit int, sort string) ([]domain.Comment, error)
	GetCommentReplies(commentID string) ([]domain.Comment, error)
	DeleteComment(commentID, userID string) error
	LikeComment(commentID, userID string) (bool, error)
	CheckUserCommented(postID, userID string) (bool, error)
	GetCommentsCount(postID string) (int, error)
}

type CommentServicePort interface {
	CreateComment(req domain.CreateCommentRequest) (*domain.Comment, error)
	GetComments(postID string, page, limit int, sort string) ([]domain.Comment, error)
	GetCommentReplies(commentID string) ([]domain.Comment, error)
	DeleteComment(commentID, userID string) error
	LikeComment(commentID, userID string) (bool, error)
	CheckUserCommented(postID, userID string) (bool, error)
	GetCommentsCount(postID string) (int, error)
}