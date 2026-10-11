package gorm_user

import "time"

type CloseFriendModel struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	UserID    string    `gorm:"type:uuid;not null;index"`
	FriendID  string    `gorm:"type:uuid;not null;index"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:now()"`
	UpdatedAt time.Time `gorm:"type:timestamp;not null;default:now()"`
}

func (CloseFriendModel) TableName() string {
	return "close_friend_models"
}

type BlockedUserModel struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	UserID    string    `gorm:"type:uuid;not null;index"`
	BlockedID string    `gorm:"type:uuid;not null;index"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:now()"`
	UpdatedAt time.Time `gorm:"type:timestamp;not null;default:now()"`
}

func (BlockedUserModel) TableName() string {
	return "blocked_user_models"
}

type HiddenStoryViewerModel struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	UserID    string    `gorm:"type:uuid;not null;index"`
	HiddenID  string    `gorm:"type:uuid;not null;index"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:now()"`
	UpdatedAt time.Time `gorm:"type:timestamp;not null;default:now()"`
}

func (HiddenStoryViewerModel) TableName() string {
	return "hidden_story_viewer_models"
}

type CloseFriendsResponse struct {
	CloseFriends []UserResponse `json:"close_friends"`
}

type BlockedUsersResponse struct {
	BlockedUsers []UserResponse `json:"blocked_users"`
}

type HiddenStoryViewersResponse struct {
	HiddenViewers []UserResponse `json:"hidden_viewers"`
}
