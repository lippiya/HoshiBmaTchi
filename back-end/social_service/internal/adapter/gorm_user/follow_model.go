package gorm_user

import (
	"time"
)

type FollowModel struct {
	ID          string    `gorm:"primaryKey;type:uuid"`
	FollowerID  string    `gorm:"type:uuid;not null;index"` 
	FollowingID string    `gorm:"type:uuid;not null;index"` 
	CreatedAt   time.Time `gorm:"type:timestamp;not null;default:now()"`
	UpdatedAt   time.Time `gorm:"type:timestamp;not null;default:now()"`
}

func (FollowModel) TableName() string {
	return "follow_models"
}


type FollowRequest struct {
	UserID       string `json:"user_id"`       
	CurrentUserID string `json:"current_user_id"` 
}


type FollowResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	IsFollowing   bool   `json:"is_following"`
	FollowersCount int    `json:"followers_count"`
	FollowingCount int    `json:"following_count"`
}


type GetFollowStatusRequest struct {
	UserID        string `json:"user_id"`        
	CurrentUserID string `json:"current_user_id"` 
}

type GetFollowStatusResponse struct {
	IsFollowing   bool `json:"is_following"`
	FollowersCount int  `json:"followers_count"`
	FollowingCount int  `json:"following_count"`
}


type GetFollowersRequest struct {
	UserID string `json:"user_id"`
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
}


type GetFollowingRequest struct {
	UserID string `json:"user_id"`
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
}

type UserResponse struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	Name           string `json:"name"`
	ProfilePicture string `json:"profile_picture"`
	IsVerified     bool   `json:"is_verified"`
}

type FollowersResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
}

type FollowingResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
}