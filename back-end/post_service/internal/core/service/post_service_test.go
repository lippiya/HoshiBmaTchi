package service

import (
	"context"
	"errors"
	"hoshiBmaTchi/back-end/post_service/internal/core/domain"
	"testing"
)

type MockPostRepo struct {
	CreateFunc      func(ctx context.Context, post *domain.Post) error
	GetByIDFunc     func(ctx context.Context, id string) (*domain.Post, error)
	GetAllFunc      func(ctx context.Context) ([]*domain.Post, error)
	GetByUserIDFunc func(ctx context.Context, userID string) ([]*domain.Post, error)
	UpdateFunc      func(ctx context.Context, post *domain.Post) error
	DeleteFunc      func(ctx context.Context, id string) error
}

func (m *MockPostRepo) Create(ctx context.Context, post *domain.Post) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, post)
	}
	return nil
}

func (m *MockPostRepo) GetByID(ctx context.Context, id string) (*domain.Post, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *MockPostRepo) GetAll(ctx context.Context) ([]*domain.Post, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc(ctx)
	}
	return []*domain.Post{}, nil
}

func (m *MockPostRepo) GetByUserID(ctx context.Context, userID string) ([]*domain.Post, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return []*domain.Post{}, nil
}

func (m *MockPostRepo) Update(ctx context.Context, post *domain.Post) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, post)
	}
	return nil
}

func (m *MockPostRepo) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func TestCreatePost_Success(t *testing.T) {
	mockRepo := &MockPostRepo{
		CreateFunc: func(ctx context.Context, post *domain.Post) error {
			return nil
		},
	}

	service := NewPostService(mockRepo)
	post := &domain.Post{
		UserID:  "user123",
		Caption: "Test post",
	}
	err := service.Create(context.Background(), post)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestGetPostByID_Success(t *testing.T) {
	mockRepo := &MockPostRepo{
		GetByIDFunc: func(ctx context.Context, id string) (*domain.Post, error) {
			return &domain.Post{
				ID:      "post123",
				UserID:  "user123",
				Caption: "Test post",
			}, nil
		},
	}

	service := NewPostService(mockRepo)
	post, err := service.GetByID(context.Background(), "post123")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if post == nil {
		t.Error("Expected post, got nil")
	}
	if post.ID != "post123" {
		t.Errorf("Expected ID post123, got %s", post.ID)
	}
}

func TestGetPostByID_NotFound(t *testing.T) {
	mockRepo := &MockPostRepo{
		GetByIDFunc: func(ctx context.Context, id string) (*domain.Post, error) {
			return nil, errors.New("post not found")
		},
	}

	service := NewPostService(mockRepo)
	post, err := service.GetByID(context.Background(), "invalid")

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if post != nil {
		t.Error("Expected nil post, got post")
	}
}

func TestDeletePost_Success(t *testing.T) {
	mockRepo := &MockPostRepo{
		DeleteFunc: func(ctx context.Context, id string) error {
			return nil
		},
	}

	service := NewPostService(mockRepo)
	err := service.Delete(context.Background(), "post123")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestGetAllPosts_Success(t *testing.T) {
	mockRepo := &MockPostRepo{
		GetAllFunc: func(ctx context.Context) ([]*domain.Post, error) {
			return []*domain.Post{
				{ID: "post1", UserID: "user1", Caption: "Post 1"},
				{ID: "post2", UserID: "user2", Caption: "Post 2"},
			}, nil
		},
	}

	service := NewPostService(mockRepo)
	posts, err := service.GetAll(context.Background())

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(posts) != 2 {
		t.Errorf("Expected 2 posts, got %d", len(posts))
	}
}
