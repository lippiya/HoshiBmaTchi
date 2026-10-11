package gorm_user

import "time"

type UserModel struct {
	ID             string    `gorm:"primaryKey;type:uuid"`
	Name           string    `gorm:"type:varchar(100);not null"`
	Username       string    `gorm:"type:varchar(100);not null;uniqueIndex"`
	Email          string    `gorm:"type:varchar(100);not null;uniqueIndex"`
	Password       string    `gorm:"type:varchar(100);not null"`
	Gender         string    `gorm:"type:varchar(10);not null"`
	ProfilePicture string    `gorm:"type:text"`
	DoB            time.Time `gorm:"type:date"`
	Bio            string    `gorm:"type:text"`
	FollowersCount int       `gorm:"type:int;default:0"`
	FollowingCount int       `gorm:"type:int;default:0"`
	PostsCount     int       `gorm:"type:int;default:0"`

	NewsletterSubscribed bool       `gorm:"type:boolean;default:false"`
	TwoFAEnabled         bool       `gorm:"type:boolean;default:false"`
	TwoFASecret          string     `gorm:"type:varchar(100)"`
	IsVerified           bool       `gorm:"type:boolean;default:false"`
	VerifiedAt           *time.Time `gorm:"type:timestamp"`

	OAuthProvider    string `gorm:"type:varchar(50);index"`
	OAuthID          string `gorm:"type:varchar(255);index"`
	ProfileCompleted bool   `gorm:"type:boolean;default:true"`

	IsBanned      bool `gorm:"type:boolean;default:false"`
	IsDeactivated bool `gorm:"type:boolean;default:false"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (UserModel) TableName() string {
	return "user_models"
}

type LoginRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	TurnstileToken string `json:"turnstile_token"`
}

type LoginResponse struct {
	Email        string `json:"email"`
	Username     string `json:"username"`
	Name         string `json:"name"`
	TwoFAEnabled bool   `json:"two_fa_enabled"`
}

type RegisterRequest struct {
	Username             string `json:"username"`
	Name                 string `json:"name"`
	Email                string `json:"email"`
	Password             string `json:"password"`
	ConfPassword         string `json:"confpassword"`
	Gender               string `json:"gender"`
	DoB                  string `json:"dob"`
	NewsletterSubscribed bool   `json:"newsletter_subscribed"`
	TwoFAEnabled         bool   `json:"two_fa_enabled"`
	TurnstileToken       string `json:"turnstile_token"`
	ProfilePicture       string `json:"profile_picture"`
}

type GetUserRequest struct {
	ID             string `json:"id"`
	Email          string `json:"email"`
	Username       string `json:"username"`
	Name           string `json:"name"`
	TwoFAEnabled   bool   `json:"two_fa_enabled"`
	Gender         string `json:"gender"`
	DoB            string `json:"dob"`
	ProfilePicture string `json:"profile_picture"`
	PostsCount     int    `json:"posts_count"`
	FollowersCount int    `json:"followers_count"`
	FollowingCount int    `json:"following_count"`
	Bio            string `json:"bio"`
}

type ResetPasswordRequest struct {
	User         string `json:"user"`
	NewPassword  string `json:"new_password"`
	ConfPassword string `json:"conf_password"`
}

type EditProfileRequest struct {
	Username       string `json:"username"`
	Name           string `json:"name"`
	Bio            string `json:"bio"`
	ProfilePicture string `json:"profile_picture"`
	Gender         string `json:"gender"`
}

type ProfilePictureRequest struct {
	UserID string `json:"user_id"`
}

type GetUserByIDRequest struct {
	ID string `json:"id"`
}

type GetUserByIDResponse struct {
	ID             string `json:"id"`
	Email          string `json:"email"`
	Username       string `json:"username"`
	Name           string `json:"name"`
	TwoFAEnabled   bool   `json:"two_fa_enabled"`
	Gender         string `json:"gender"`
	DoB            string `json:"dob"`
	ProfilePicture string `json:"profile_picture"`
	Bio            string `json:"bio"`
}

type FollowRequest struct {
	UserID string `json:"user_id"`
}

type FollowResponse struct {
	Success          bool   `json:"success"`
	Message          string `json:"message"`
	IsFollowing      bool   `json:"is_following"`
	FollowersCount   int    `json:"followers_count"`
	FollowingCount   int    `json:"following_count"`
	MyFollowingCount int    `json:"my_following_count"`
}

type GetFollowStatusResponse struct {
	IsFollowing    bool `json:"is_following"`
	FollowersCount int  `json:"followers_count"`
	FollowingCount int  `json:"following_count"`
}

type NotificationSettings struct {
	UserID            string    `gorm:"primaryKey;type:uuid"`
	PushNotification  bool      `gorm:"type:boolean;default:true"`
	EmailNotification bool      `gorm:"type:boolean;default:true"`
	CreatedAt         time.Time `gorm:"type:timestamp;not null;default:now()"`
	UpdatedAt         time.Time `gorm:"type:timestamp;not null;default:now()"`
}

func (NotificationSettings) TableName() string {
	return "notification_settings"
}

type PrivacySettings struct {
	UserID    string    `gorm:"primaryKey;type:uuid"`
	IsPrivate bool      `gorm:"type:boolean;default:false"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:now()"`
	UpdatedAt time.Time `gorm:"type:timestamp;not null;default:now()"`
}

func (PrivacySettings) TableName() string {
	return "privacy_settings"
}
