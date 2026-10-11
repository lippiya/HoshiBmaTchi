package service

import (
	"errors"
	"story_service/internal/core/domain"
	"story_service/internal/core/port"
)

type StoryService struct {
	repo port.StoryRepository
}

func NewStoryService(repo port.StoryRepository) *StoryService {
	return &StoryService{repo: repo}
}

func (s *StoryService) CreateStory(userID, mediaURL, mediaType string, duration int) (*domain.Story, error) {
	story := &domain.Story{
		UserID:    userID,
		MediaURL:  mediaURL,
		MediaType: mediaType,
		Duration:  duration,
	}

	if err := s.repo.CreateStory(story); err != nil {
		return nil, err
	}

	return story, nil
}

func (s *StoryService) GetStoryByID(id string) (*domain.Story, error) {
	return s.repo.GetStoryByID(id)
}

func (s *StoryService) GetStoriesByUserID(userID string) ([]*domain.Story, error) {
	return s.repo.GetStoriesByUserID(userID)
}

func (s *StoryService) GetAllActiveStories() ([]*domain.Story, error) {
	return s.repo.GetAllActiveStories()
}

func (s *StoryService) DeleteStory(id, userID string) error {
	story, err := s.repo.GetStoryByID(id)
	if err != nil {
		return err
	}

	if story.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.repo.DeleteStory(id)
}

func (s *StoryService) CleanupExpiredStories() error {
	return s.repo.DeleteExpiredStories()
}

func (s *StoryService) LikeStory(storyID, userID string) error {
	isLiked, err := s.repo.IsStoryLiked(storyID, userID)
	if err != nil {
		return err
	}

	if isLiked {
		return s.repo.UnlikeStory(storyID, userID)
	}

	return s.repo.LikeStory(storyID, userID)
}

func (s *StoryService) UnlikeStory(storyID, userID string) error {
	return s.repo.UnlikeStory(storyID, userID)
}

func (s *StoryService) IsStoryLiked(storyID, userID string) (bool, error) {
	return s.repo.IsStoryLiked(storyID, userID)
}

func (s *StoryService) ReplyToStory(storyID, userID, reply string) error {
	storyReply := &domain.StoryReply{
		StoryID: storyID,
		UserID:  userID,
		Reply:   reply,
	}

	return s.repo.CreateStoryReply(storyReply)
}

func (s *StoryService) GetStoryReplies(storyID string) ([]*domain.StoryReply, error) {
	return s.repo.GetStoryReplies(storyID)
}

func (s *StoryService) ViewStory(storyID, userID string) error {
	view := &domain.StoryView{
		StoryID: storyID,
		UserID:  userID,
	}

	return s.repo.CreateStoryView(view)
}

func (s *StoryService) HasViewedStory(storyID, userID string) (bool, error) {
	return s.repo.HasViewedStory(storyID, userID)
}

func (s *StoryService) GetStoryViews(storyID string) ([]*domain.StoryView, error) {
	return s.repo.GetStoryViews(storyID)
}

func (s *StoryService) ShareStory(storyID, senderID, recipientID string) error {
	share := &domain.StoryShare{
		StoryID:     storyID,
		SenderID:    senderID,
		RecipientID: recipientID,
	}

	return s.repo.CreateStoryShare(share)
}

func (s *StoryService) GetArchivedStories(userID string, limit, offset int) ([]*domain.Story, error) {
	return s.repo.GetArchivedStoriesByUserID(userID, limit, offset)
}
