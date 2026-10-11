package port

import "story_service/internal/core/domain"

type StoryRepository interface {
	CreateStory(story *domain.Story) error
	GetStoryByID(id string) (*domain.Story, error)
	GetStoriesByUserID(userID string) ([]*domain.Story, error)
	GetAllActiveStories() ([]*domain.Story, error)
	GetArchivedStoriesByUserID(userID string, limit, offset int) ([]*domain.Story, error)
	DeleteStory(id string) error
	DeleteExpiredStories() error

	LikeStory(storyID, userID string) error
	UnlikeStory(storyID, userID string) error
	IsStoryLiked(storyID, userID string) (bool, error)
	GetStoryLikes(storyID string) ([]*domain.StoryLike, error)

	CreateStoryReply(reply *domain.StoryReply) error
	GetStoryReplies(storyID string) ([]*domain.StoryReply, error)

	CreateStoryView(view *domain.StoryView) error
	HasViewedStory(storyID, userID string) (bool, error)
	GetStoryViews(storyID string) ([]*domain.StoryView, error)

	CreateStoryShare(share *domain.StoryShare) error
	GetStoryShares(storyID string) ([]*domain.StoryShare, error)
}

type StoryService interface {
	CreateStory(userID, mediaURL, mediaType string, duration int) (*domain.Story, error)
	GetStoryByID(id string) (*domain.Story, error)
	GetStoriesByUserID(userID string) ([]*domain.Story, error)
	GetAllActiveStories() ([]*domain.Story, error)
	GetArchivedStories(userID string, limit, offset int) ([]*domain.Story, error)
	DeleteStory(id, userID string) error
	CleanupExpiredStories() error

	LikeStory(storyID, userID string) error
	UnlikeStory(storyID, userID string) error
	IsStoryLiked(storyID, userID string) (bool, error)

	ReplyToStory(storyID, userID, reply string) error
	GetStoryReplies(storyID string) ([]*domain.StoryReply, error)

	ViewStory(storyID, userID string) error
	HasViewedStory(storyID, userID string) (bool, error)
	GetStoryViews(storyID string) ([]*domain.StoryView, error)

	ShareStory(storyID, senderID, recipientID string) error
}
