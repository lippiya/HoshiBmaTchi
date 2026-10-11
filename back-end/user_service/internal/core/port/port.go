package port

import (
	"context"

	
	"hoshiBmaTchi/back-end/user_service/internal/core/domain"
)

type UserServicePort interface {
	Register(c context.Context, user *domain.User, confPassword string) error
	OAuthRegister(c context.Context, authData domain.OAuthUserData) (*domain.User, error)
	Login(c context.Context, userstr string, password string) (*domain.User, error)
	PasswordValidation(password string, confirmPassword string) (bool, error)
	CompleteProfile(c context.Context, user *domain.User) error
	UpdateUser(c context.Context, user *domain.User) error
	GetByEmail(c context.Context, email string) (*domain.User, error)
	GetByUsername(c context.Context, id string) (*domain.User, error)
	LogOut(c context.Context, user *domain.User) error
	ResetPassword(c context.Context, userstr string, newPassword string, confPassword string) error
	EditProfile(c context.Context, pfp string, email string, gender string, bio string, username string, name string) error
	GetByID(c context.Context, ID string) (*domain.User, error)
	SearchUsers(c context.Context, query string) ([]domain.User, error)
	GetAllUsers(c context.Context) ([]domain.User, error)

	FollowUser(c context.Context, followerID, followingID string) (*domain.User, error)
	UnfollowUser(c context.Context, followerID, followingID string) (*domain.User, error)
	IsFollowing(c context.Context, followerID, followingID string) (bool, error)
	GetFollowersCount(c context.Context, userID string) (int, error)
	GetFollowingCount(c context.Context, userID string) (int, error)
	GetNotificationSettings(c context.Context, userID string) (*domain.NotificationSettingsResponse, error)
	UpdateNotificationSettings(c context.Context, userID string, settings *domain.NotificationSettings) error
	GetPrivacySettings(c context.Context, userID string) (*domain.PrivacySettingsResponse, error)
	UpdatePrivacySettings(c context.Context, userID string, settings *domain.PrivacySettings) error
}

type UserRepositoryPort interface {
	CreateUser(c context.Context, user *domain.User) error
	UpdateUser(c context.Context, user *domain.User) error
	DeleteUser(c context.Context, ID string) error
	GetByEmail(c context.Context, email string) (*domain.User, error)
	GetByUsername(c context.Context, id string) (*domain.User, error)
	GetByOAuthID(c context.Context, provider, oauthID string) (*domain.User, error)
	GetByID(c context.Context, ID string) (*domain.User, error)
	SearchUsers(c context.Context, query string) ([]domain.User, error)
	GetAllUsers(c context.Context) ([]domain.User, error)
	GetNotificationSettings(c context.Context, userID string) (*domain.NotificationSettings, error)
	UpdateNotificationSettings(c context.Context, userID string, settings *domain.NotificationSettings) error
	GetPrivacySettings(c context.Context, userID string) (*domain.PrivacySettings, error)
	UpdatePrivacySettings(c context.Context, userID string, settings *domain.PrivacySettings) error
}

type OTPServicePort interface {
	SendOTP(c context.Context, email string) (string, error)
	VerifyOTP(email string, code string) (bool, error)
	SendWelcomeEmail(c context.Context, recipient string, name string) error
}

type SpamServicePort interface {
	Validate(token string) error
}

type OAuthServicePort interface {
	GetGoogleUserInfo(code string) (*domain.OAuthUserData, error)
	GenerateAuthURL(state string) string
}

type PhotoClientPort interface {
	UploadPhoto(file interface{}, userID string) (string, error)
	DeletePhoto(filePath string) error
	GetPhotoURL(filePath string) (string, error)
}
