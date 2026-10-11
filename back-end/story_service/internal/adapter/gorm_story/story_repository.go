package gorm_story

import (
	"errors"
	"story_service/internal/core/domain"
	"time"

	"gorm.io/gorm"
)

type StoryRepository struct {
	db *gorm.DB
}

func NewStoryRepository(db *gorm.DB) *StoryRepository {
	return &StoryRepository{db: db}
}

func (r *StoryRepository) CreateStory(story *domain.Story) error {
	model := &StoryModel{
		UserID:    story.UserID,
		MediaURL:  story.MediaURL,
		MediaType: story.MediaType,
		Duration:  story.Duration,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := r.db.Create(model).Error; err != nil {
		return err
	}

	story.ID = model.ID
	story.CreatedAt = model.CreatedAt
	story.ExpiresAt = model.ExpiresAt

	return nil
}

func (r *StoryRepository) GetStoryByID(id string) (*domain.Story, error) {
	var model StoryModel
	if err := r.db.Where("id = ?", id).First(&model).Error; err != nil {
		return nil, err
	}

	return &domain.Story{
		ID:        model.ID,
		UserID:    model.UserID,
		MediaURL:  model.MediaURL,
		MediaType: model.MediaType,
		Duration:  model.Duration,
		CreatedAt: model.CreatedAt,
		ExpiresAt: model.ExpiresAt,
	}, nil
}

func (r *StoryRepository) GetStoriesByUserID(userID string) ([]*domain.Story, error) {
	var models []StoryModel
	if err := r.db.Where("user_id = ? AND expires_at > ?", userID, time.Now()).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}

	stories := make([]*domain.Story, len(models))
	for i, model := range models {
		stories[i] = &domain.Story{
			ID:        model.ID,
			UserID:    model.UserID,
			MediaURL:  model.MediaURL,
			MediaType: model.MediaType,
			Duration:  model.Duration,
			CreatedAt: model.CreatedAt,
			ExpiresAt: model.ExpiresAt,
		}
	}

	return stories, nil
}

func (r *StoryRepository) GetAllActiveStories() ([]*domain.Story, error) {
	var models []StoryModel
	if err := r.db.Where("expires_at > ?", time.Now()).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}

	stories := make([]*domain.Story, len(models))
	for i, model := range models {
		stories[i] = &domain.Story{
			ID:        model.ID,
			UserID:    model.UserID,
			MediaURL:  model.MediaURL,
			MediaType: model.MediaType,
			Duration:  model.Duration,
			CreatedAt: model.CreatedAt,
			ExpiresAt: model.ExpiresAt,
		}
	}

	return stories, nil
}

func (r *StoryRepository) DeleteStory(id string) error {
	return r.db.Delete(&StoryModel{}, "id = ?", id).Error
}

func (r *StoryRepository) DeleteExpiredStories() error {
	return r.db.Delete(&StoryModel{}, "expires_at <= ?", time.Now()).Error
}

func (r *StoryRepository) LikeStory(storyID, userID string) error {
	model := &StoryLikeModel{
		StoryID: storyID,
		UserID:  userID,
	}
	return r.db.Create(model).Error
}

func (r *StoryRepository) UnlikeStory(storyID, userID string) error {
	return r.db.Where("story_id = ? AND user_id = ?", storyID, userID).Delete(&StoryLikeModel{}).Error
}

func (r *StoryRepository) IsStoryLiked(storyID, userID string) (bool, error) {
	var count int64
	err := r.db.Model(&StoryLikeModel{}).Where("story_id = ? AND user_id = ?", storyID, userID).Count(&count).Error
	return count > 0, err
}

func (r *StoryRepository) GetStoryLikes(storyID string) ([]*domain.StoryLike, error) {
	var models []StoryLikeModel
	if err := r.db.Where("story_id = ?", storyID).Find(&models).Error; err != nil {
		return nil, err
	}

	likes := make([]*domain.StoryLike, len(models))
	for i, model := range models {
		likes[i] = &domain.StoryLike{
			ID:        model.ID,
			StoryID:   model.StoryID,
			UserID:    model.UserID,
			CreatedAt: model.CreatedAt,
		}
	}

	return likes, nil
}

func (r *StoryRepository) CreateStoryReply(reply *domain.StoryReply) error {
	model := &StoryReplyModel{
		StoryID: reply.StoryID,
		UserID:  reply.UserID,
		Reply:   reply.Reply,
	}

	if err := r.db.Create(model).Error; err != nil {
		return err
	}

	reply.ID = model.ID
	reply.CreatedAt = model.CreatedAt

	return nil
}

func (r *StoryRepository) GetStoryReplies(storyID string) ([]*domain.StoryReply, error) {
	var models []StoryReplyModel
	if err := r.db.Where("story_id = ?", storyID).Order("created_at ASC").Find(&models).Error; err != nil {
		return nil, err
	}

	replies := make([]*domain.StoryReply, len(models))
	for i, model := range models {
		replies[i] = &domain.StoryReply{
			ID:        model.ID,
			StoryID:   model.StoryID,
			UserID:    model.UserID,
			Reply:     model.Reply,
			CreatedAt: model.CreatedAt,
		}
	}

	return replies, nil
}

func (r *StoryRepository) CreateStoryView(view *domain.StoryView) error {
	var count int64
	r.db.Model(&StoryViewModel{}).Where("story_id = ? AND user_id = ?", view.StoryID, view.UserID).Count(&count)

	if count > 0 {
		return errors.New("already viewed")
	}

	model := &StoryViewModel{
		StoryID: view.StoryID,
		UserID:  view.UserID,
	}

	if err := r.db.Create(model).Error; err != nil {
		return err
	}

	view.ID = model.ID
	view.CreatedAt = model.CreatedAt

	return nil
}

func (r *StoryRepository) HasViewedStory(storyID, userID string) (bool, error) {
	var count int64
	err := r.db.Model(&StoryViewModel{}).Where("story_id = ? AND user_id = ?", storyID, userID).Count(&count).Error
	return count > 0, err
}

func (r *StoryRepository) GetStoryViews(storyID string) ([]*domain.StoryView, error) {
	var models []StoryViewModel
	if err := r.db.Where("story_id = ?", storyID).Find(&models).Error; err != nil {
		return nil, err
	}

	views := make([]*domain.StoryView, len(models))
	for i, model := range models {
		views[i] = &domain.StoryView{
			ID:        model.ID,
			StoryID:   model.StoryID,
			UserID:    model.UserID,
			CreatedAt: model.CreatedAt,
		}
	}

	return views, nil
}

func (r *StoryRepository) CreateStoryShare(share *domain.StoryShare) error {
	model := &StoryShareModel{
		StoryID:     share.StoryID,
		SenderID:    share.SenderID,
		RecipientID: share.RecipientID,
	}

	if err := r.db.Create(model).Error; err != nil {
		return err
	}

	share.ID = model.ID
	share.CreatedAt = model.CreatedAt

	return nil
}

func (r *StoryRepository) GetStoryShares(storyID string) ([]*domain.StoryShare, error) {
	var models []StoryShareModel
	if err := r.db.Where("story_id = ?", storyID).Find(&models).Error; err != nil {
		return nil, err
	}

	shares := make([]*domain.StoryShare, len(models))
	for i, model := range models {
		shares[i] = &domain.StoryShare{
			ID:          model.ID,
			StoryID:     model.StoryID,
			SenderID:    model.SenderID,
			RecipientID: model.RecipientID,
			CreatedAt:   model.CreatedAt,
		}
	}

	return shares, nil
}

func (r *StoryRepository) GetArchivedStoriesByUserID(userID string, limit, offset int) ([]*domain.Story, error) {
	var models []StoryModel
	if err := r.db.Where("user_id = ? AND expires_at <= ?", userID, time.Now()).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error; err != nil {
		return nil, err
	}

	stories := make([]*domain.Story, len(models))
	for i, model := range models {
		stories[i] = &domain.Story{
			ID:        model.ID,
			UserID:    model.UserID,
			MediaURL:  model.MediaURL,
			MediaType: model.MediaType,
			Duration:  model.Duration,
			CreatedAt: model.CreatedAt,
			ExpiresAt: model.ExpiresAt,
		}
	}

	return stories, nil
}
