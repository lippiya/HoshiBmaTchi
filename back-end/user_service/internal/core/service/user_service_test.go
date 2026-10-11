package service

import (
	"context"
	"errors"
	"hoshiBmaTchi/back-end/user_service/internal/core/domain"
	"testing"
)

type MockUserRepo struct {
	CreateFunc     func(ctx context.Context, user *domain.User) error
	GetByEmailFunc func(ctx context.Context, email string) (*domain.User, error)
	GetByIDFunc    func(ctx context.Context, id string) (*domain.User, error)
	UpdateFunc     func(ctx context.Context, user *domain.User) error
	SearchFunc     func(ctx context.Context, query string) ([]*domain.User, error)
}

func (m *MockUserRepo) Create(ctx context.Context, user *domain.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(ctx, email)
	}
	return nil, errors.New("not found")
}

func (m *MockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *MockUserRepo) Update(ctx context.Context, user *domain.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepo) Search(ctx context.Context, query string) ([]*domain.User, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, query)
	}
	return nil, nil
}

type MockRedisClient struct {
	GetFunc func(ctx context.Context, key string) (string, error)
	SetFunc func(ctx context.Context, key string, value string) error
	DelFunc func(ctx context.Context, key string) error
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key)
	}
	return "", errors.New("not found")
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value string) error {
	if m.SetFunc != nil {
		return m.SetFunc(ctx, key, value)
	}
	return nil
}

func (m *MockRedisClient) Del(ctx context.Context, key string) error {
	if m.DelFunc != nil {
		return m.DelFunc(ctx, key)
	}
	return nil
}

func TestGetByEmail_Success(t *testing.T) {
	mockRepo := &MockUserRepo{
		GetByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
			return &domain.User{
				ID:    "123",
				Email: "test@example.com",
				Name:  "Test User",
			}, nil
		},
	}
	mockRedis := &MockRedisClient{}

	service := NewUserService(mockRepo, mockRedis)
	user, err := service.GetByEmail(context.Background(), "test@example.com")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if user == nil {
		t.Error("Expected user, got nil")
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", user.Email)
	}
}

func TestGetByEmail_NotFound(t *testing.T) {
	mockRepo := &MockUserRepo{
		GetByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
			return nil, errors.New("user not found")
		},
	}
	mockRedis := &MockRedisClient{}

	service := NewUserService(mockRepo, mockRedis)
	user, err := service.GetByEmail(context.Background(), "notfound@example.com")

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if user != nil {
		t.Error("Expected nil user, got user")
	}
}

func TestGetByID_Success(t *testing.T) {
	mockRepo := &MockUserRepo{
		GetByIDFunc: func(ctx context.Context, id string) (*domain.User, error) {
			return &domain.User{
				ID:    "123",
				Email: "test@example.com",
				Name:  "Test User",
			}, nil
		},
	}
	mockRedis := &MockRedisClient{}

	service := NewUserService(mockRepo, mockRedis)
	user, err := service.GetByID(context.Background(), "123")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if user == nil {
		t.Error("Expected user, got nil")
	}
	if user.ID != "123" {
		t.Errorf("Expected ID 123, got %s", user.ID)
	}
}

func TestUpdateUser_Success(t *testing.T) {
	mockRepo := &MockUserRepo{
		UpdateFunc: func(ctx context.Context, user *domain.User) error {
			return nil
		},
	}
	mockRedis := &MockRedisClient{}

	service := NewUserService(mockRepo, mockRedis)
	user := &domain.User{
		ID:    "123",
		Email: "test@example.com",
		Name:  "Updated User",
	}
	err := service.UpdateUser(context.Background(), user)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
