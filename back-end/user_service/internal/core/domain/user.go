package domain

import "time"

type User struct{
	ID string
	Name string
	Username string
	Email string
	Password string
	Gender string
	ProfilePicture string
	DoB time.Time
	Bio string
	FollowersCount int
	FollowingCount int
	PostsCount int

	NewsletterSubscribed bool
	TwoFAEnabled bool
	TwoFASecret string
	IsVerified bool
	VerifiedAt time.Time
	
	OAuthProvider string
	OAuthID string
	ProfileCompleted bool

	IsBanned bool
	IsDeactivated bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (User) TableName() string {
    return "user_models"
}

type OAuthUserData struct {
    ID string
    Email string
    Name string
    Picture string
    VerifiedEmail bool
}

type CompleteRegistrationRequest struct {
	Email string `json:"email"`
	ProfilePicture string `json:"profile_picture"`
	NewsletterSubscribed bool `json:"newsletter_subscribed"`
	TwoFAEnabled bool `json:"two_fa_enabled"`
}

type NotificationSettings struct {
	UserID            string 
	PushNotification  bool
	EmailNotification bool
}

type NotificationSettingsResponse struct {
	PushNotification  bool 
	EmailNotification bool 
}

type PrivacySettings struct {
	UserID    string `gorm:"primaryKey"`
	IsPrivate bool
}

type PrivacySettingsResponse struct {
	IsPrivate bool 
}