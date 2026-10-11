package gorm_user

import (
	"context"
	"hoshiBmaTchi/back-end/user_service/internal/core/domain"
	"hoshiBmaTchi/back-end/user_service/internal/core/port"
	"strings"

	"gorm.io/gorm"
)


type userRepository struct {
	db *gorm.DB
}

func NewGormUserRepo(db *gorm.DB) port.UserRepositoryPort {
	return &userRepository{
		db: db,
	}
}

func (ur *userRepository) CreateUser(c context.Context, user *domain.User) error {
	result := ur.db.WithContext(c).Create(user)
	return result.Error

}

func (ur *userRepository) DeleteUser(c context.Context, ID string) error {
	result := ur.db.WithContext(c).Delete(&domain.User{}, "ID = ?", ID)
	return result.Error
}

func (ur *userRepository) GetByEmail(c context.Context, email string) (*domain.User, error) {
	var user domain.User
	result := ur.db.WithContext(c).Where("email = ?", email).First(&user)
	return &user, result.Error
}

func (ur *userRepository) GetByUsername(c context.Context, username string) (*domain.User, error) {
	var user domain.User
	result := ur.db.WithContext(c).Where("username = ?", username).First(&user)
	return &user, result.Error
}

func (ur *userRepository) UpdateUser(c context.Context, user *domain.User) error {
	result := ur.db.WithContext(c).Save(user)
	return result.Error
}

func (ur *userRepository) GetByOAuthID(ctx context.Context, provider, oauthID string) (*domain.User, error) {
	var user domain.User
	result := ur.db.WithContext(ctx).Where("oauth_provider = ? AND oauth_id = ?", provider, oauthID).First(&user)
	return &user, result.Error
}

func (ur *userRepository) GetByID(ctx context.Context, ID string) (*domain.User, error) {
	var user domain.User
	result := ur.db.WithContext(ctx).Where("id = ?", ID).First(&user)
	return &user, result.Error
}

func (ur *userRepository) SearchUsers(ctx context.Context, query string) ([]domain.User, error) {
	var allUsers []domain.User
	result := ur.db.WithContext(ctx).Find(&allUsers)
	if result.Error != nil {
		return nil, result.Error
	}

	var matchedUsers []domain.User
	for _, user := range allUsers {
		if jaroWinklerDistance(query, user.Username) >= 0.7 || jaroWinklerDistance(query, user.Email) >= 0.7 {
			matchedUsers = append(matchedUsers, user)
		}
	}

	return matchedUsers, nil
}

func (ur *userRepository) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	var users []domain.User
	result := ur.db.WithContext(ctx).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func jaroWinklerDistance(s1, s2 string) float64 {
	if s1 == "" || s2 == "" {
		return 0
	}

	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	matchWindow := len(s1)/2 - 1
	if matchWindow < 0 {
		matchWindow = 0
	}

	s1Matches := make([]bool, len(s1))
	s2Matches := make([]bool, len(s2))

	matches := 0
	for i := 0; i < len(s1); i++ {
		start := i - matchWindow
		if start < 0 {
			start = 0
		}
		end := i + matchWindow + 1
		if end > len(s2) {
			end = len(s2)
		}

		for j := start; j < end; j++ {
			if !s2Matches[j] && s1[i] == s2[j] {
				s1Matches[i] = true
				s2Matches[j] = true
				matches++
				break
			}
		}
	}

	if matches == 0 {
		return 0
	}

	transpositions := 0
	k := 0
	for i := 0; i < len(s1); i++ {
		if s1Matches[i] {
			for !s2Matches[k] {
				k++
			}
			if s1[i] != s2[k] {
				transpositions++
			}
			k++
		}
	}

	jaro := float64(matches)/float64(len(s1)) +
		float64(matches)/float64(len(s2)) +
		float64(matches-transpositions/2)/float64(matches)
	jaro /= 3

	prefixLength := 0
	for i := 0; i < len(s1) && i < len(s2) && i < 4 && s1[i] == s2[i]; i++ {
		prefixLength++
	}

	jaroWinkler := jaro + 0.1*float64(prefixLength)*(1-jaro)

	return jaroWinkler
}



func (ur *userRepository) GetNotificationSettings(ctx context.Context, userID string) (*domain.NotificationSettings, error) {
	var settings domain.NotificationSettings
	result := ur.db.WithContext(ctx).Where("user_id = ?", userID).First(&settings)
	if result.Error == gorm.ErrRecordNotFound {
		settings = domain.NotificationSettings{
			UserID:            userID,
			PushNotification:  true,
			EmailNotification: true,
		}
		ur.db.WithContext(ctx).Create(&settings)
		return &settings, nil
	}
	return &settings, result.Error
}

func (ur *userRepository) UpdateNotificationSettings(ctx context.Context, userID string, settings *domain.NotificationSettings) error {
	settings.UserID = userID
	return ur.db.WithContext(ctx).Save(settings).Error
}

func (ur *userRepository) GetPrivacySettings(ctx context.Context, userID string) (*domain.PrivacySettings, error) {
	var settings domain.PrivacySettings
	result := ur.db.WithContext(ctx).Where("user_id = ?", userID).First(&settings)
	if result.Error == gorm.ErrRecordNotFound {
		settings = domain.PrivacySettings{
			UserID:    userID,
			IsPrivate: false,
		}
		ur.db.WithContext(ctx).Create(&settings)
		return &settings, nil
	}
	return &settings, result.Error
}

func (ur *userRepository) UpdatePrivacySettings(ctx context.Context, userID string, settings *domain.PrivacySettings) error {
	settings.UserID = userID
	return ur.db.WithContext(ctx).Save(settings).Error
}