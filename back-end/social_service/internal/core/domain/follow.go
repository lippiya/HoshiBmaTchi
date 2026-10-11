package domain

import (
	"time"
)

type FollowModel struct {
	ID          string 
	FollowerID  string 
	FollowingID string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type FollowRequest struct {
	UserID       string 
	CurrentUserID string
}


type FollowResponse struct {
	Success       bool  
	Message       string
	IsFollowing   bool  
	FollowersCount int  
	FollowingCount int  
}


type GetFollowStatusRequest struct {
	UserID        string
	CurrentUserID string
}

type GetFollowStatusResponse struct {
	IsFollowing   bool 
	FollowersCount int  
	FollowingCount int  
}


type GetFollowersRequest struct {
	UserID string 
	Page   int    
	Limit  int    
}


type GetFollowingRequest struct {
	UserID string 
	Page   int    
	Limit  int    
}

type UserResponse struct {
	ID             string 
	Username       string 
	Name           string 
	ProfilePicture string 
	IsVerified     bool  
}

type FollowersResponse struct {
	Users []UserResponse 
	Total int            
	Page  int   
}

type FollowingResponse struct {
	Users []UserResponse 
	Total int            
	Page  int            
}