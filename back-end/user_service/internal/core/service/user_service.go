package service

import (
	"context"
	"encoding/json"
	"errors"
	"hoshiBmaTchi/back-end/user_service/internal/adapter/redis"
	"hoshiBmaTchi/back-end/user_service/internal/core/domain"
	"hoshiBmaTchi/back-end/user_service/internal/core/port"
	"log"

	"regexp"
	"time"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var regex = regexp.MustCompile(`^[a-zA-Z ]+$`)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.com$`)
var today = time.Now()
var validDate = today.AddDate(-13, 0, 0)

type UserService struct {
	userRepo    port.UserRepositoryPort
	redisClient *redis.RedisClient
}

func NewUserService(userRepo port.UserRepositoryPort, redisClient *redis.RedisClient) port.UserServicePort {
	return &UserService{
		userRepo:    userRepo,
		redisClient: redisClient,
	}

}

func (us *UserService) Register(c context.Context, user *domain.User, confirmPassword string) error {

	_, errorEmail := us.userRepo.GetByEmail(c, user.Email)
	_, errorUsername := us.userRepo.GetByUsername(c, user.Username)
	_, errorPassword := us.PasswordValidation(user.Password, confirmPassword)

	if len(user.Username) < 4 {
		return errors.New("username must be at least 4 characters")
	}

	if errorUsername == nil {
		return errors.New("username is already taken")
	}

	if !emailRegex.MatchString(user.Email) {
		return errors.New("email must be in format [emailname]@[domain].com")
	}

	if !regex.MatchString(user.Name) {
		return errors.New("name must NOT contain numbers and symbols")
	}

	if errorEmail == nil {
		return errors.New("email is already registered")
	}

	if errorPassword != nil {
		return errorPassword
	}

	if user.Gender != "Male" && user.Gender != "Female" {
		return errors.New("gender must be either Male or Female")
	}

	if user.DoB.After(validDate) {
		return errors.New("you must be at least 13 years old to register")
	}

	pwByte := []byte(user.Password)
	password, err := bcrypt.GenerateFromPassword(pwByte, 12)
	if err != nil {
		return errors.New("failed to hash password")
	}
	user.Password = string(password)
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	return us.userRepo.CreateUser(c, user)
}

func (us *UserService) EditProfile(c context.Context, pfp string, email string, gender string, bio string, username string, name string) error {
	var user *domain.User
	var err error
	cachedData, cacheErr := us.redisClient.GetUserCache(email)
	if cacheErr == nil && cachedData != "" {
		user = &domain.User{}
		if err := json.Unmarshal([]byte(cachedData), user); err == nil {
			log.Printf("Cache hit for user: %s", email)
		} else {
			log.Printf("Failed to unmarshal cached user: %v", err)
			user = nil
		}
	}

	if user == nil {
		log.Printf("Cache miss for user: %s", email)
		user, err = us.userRepo.GetByEmail(c, email)
		if err != nil {

			return errors.New("user not found with provided email or username")

		}
	}

	if username != "" && username != user.Username {
		if len(username) < 4 {
			return errors.New("username must be at least 4 characters")
		}
		existingUser, err := us.userRepo.GetByUsername(c, username)
		if err == nil && existingUser != nil && existingUser.ID != user.ID {
			return errors.New("username is already taken")
		}
		user.Username = username
	}

	if name != "" {
		if !regex.MatchString(name) {
			return errors.New("name must NOT contain numbers and symbols")
		}
		user.Name = name
	}

	if gender != "" {
		user.Gender = gender
	}
	if bio != "" {
		user.Bio = bio

	}
	if pfp != "" {
		user.ProfilePicture = pfp
	}
	user.UpdatedAt = time.Now()
	err = us.userRepo.UpdateUser(c, user)
	if err != nil {
		return err
	}
	userJSON, _ := json.Marshal(user)
	us.redisClient.SetUserCache(user.Email, string(userJSON))
	us.redisClient.SetUserCache(user.Username, string(userJSON))
	return nil
}
func (us *UserService) ResetPassword(c context.Context, userstr string, newPassword string, confirmPassword string) error {
	_, errorPassword := us.PasswordValidation(newPassword, confirmPassword)
	if errorPassword != nil {
		return errorPassword
	}
	var user *domain.User
	var err error

	cachedData, cacheErr := us.redisClient.GetUserCache(userstr)
	if cacheErr == nil && cachedData != "" {
		user = &domain.User{}
		if err := json.Unmarshal([]byte(cachedData), user); err == nil {
			log.Printf("Cache hit for user: %s", userstr)
		} else {
			log.Printf("Failed to unmarshal cached user: %v", err)
			user = nil
		}
	}

	if user == nil {
		log.Printf("Cache miss for user: %s", userstr)
		user, err = us.userRepo.GetByEmail(c, userstr)
		if err != nil {
			user, err = us.userRepo.GetByUsername(c, userstr)
			if err != nil {
				return errors.New("user not found with provided email or username")
			}
		}
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(newPassword))
	if err == nil {
		return errors.New("new password must not be the same as old password")
	}
	pwByte := []byte(newPassword)
	password, err := bcrypt.GenerateFromPassword(pwByte, 12)
	if err != nil {
		return err
	}
	user.Password = string(password)
	user.UpdatedAt = time.Now()
	err = us.userRepo.UpdateUser(c, user)
	if err != nil {
		return err
	}
	userJSON, _ := json.Marshal(user)
	us.redisClient.SetUserCache(user.Email, string(userJSON))
	us.redisClient.SetUserCache(user.Username, string(userJSON))
	return nil
}

func (us *UserService) Login(c context.Context, userstr string, password string) (*domain.User, error) {
	var user *domain.User
	var err error

	cachedData, cacheErr := us.redisClient.GetUserCache(userstr)
	if cacheErr == nil && cachedData != "" {
		user = &domain.User{}
		if err := json.Unmarshal([]byte(cachedData), user); err == nil {
			log.Printf("Cache hit for user: %s", userstr)
		} else {
			log.Printf("Failed to unmarshal cached user: %v", err)
			user = nil
		}
	}

	if user == nil {

		user, err = us.userRepo.GetByEmail(c, userstr)
		if err != nil {
			user, err = us.userRepo.GetByUsername(c, userstr)
			if err != nil {
				return nil, errors.New("invalid email/username or password")
			}
		}

		userJSON, _ := json.Marshal(user)
		us.redisClient.SetUserCache(user.Email, string(userJSON))
		us.redisClient.SetUserCache(user.Username, string(userJSON))
	}

	if user.IsBanned {
		return nil, errors.New("your account has been banned")
	}

	if user.IsDeactivated {
		return nil, errors.New("your account has been deactivated")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid email/username or password")
	}

	return user, nil
}

func (us *UserService) PasswordValidation(password string, confirmPassword string) (bool, error) {
	valid := false
	if len(password) < 8 {
		return valid, errors.New("password must be at least 8 characters")
	}
	var (
		hasUpper  = false
		hasSymbol = false
		hasNumber = false
	)
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsSymbol(char):
			hasSymbol = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char):
			hasSymbol = true
		}

	}
	if !hasUpper {
		return valid, errors.New("password must contain at least 1 uppercase letter")
	}
	if !hasSymbol {
		return valid, errors.New("password must contain at least 1 symbol")
	}
	if !hasNumber {
		return valid, errors.New("password must contain at least 1 number")
	}
	if password != confirmPassword {
		return valid, errors.New("passwords do not match")
	}
	return true, nil
}

func (us *UserService) OAuthRegister(c context.Context, authData domain.OAuthUserData) (*domain.User, error) {
	existingUser, _ := us.userRepo.GetByOAuthID(c, "google", authData.ID)
	if existingUser != nil {
		return existingUser, nil
	}

	existingUser, _ = us.userRepo.GetByEmail(c, authData.Email)
	if existingUser != nil {
		return nil, errors.New("email already registered with password login")
	}

	user := &domain.User{
		ID:               uuid.New().String(),
		Email:            authData.Email,
		Name:             authData.Name,
		ProfilePicture:   authData.Picture,
		OAuthProvider:    "google",
		OAuthID:          authData.ID,
		IsVerified:       authData.VerifiedEmail,
		ProfileCompleted: false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if user.IsVerified {
		user.VerifiedAt = time.Now()
	}

	return user, nil
}

func (us *UserService) CompleteProfile(c context.Context, user *domain.User) error {
	if user.Username == "" || len(user.Username) < 4 {
		return errors.New("username must be at least 4 characters")
	}

	_, err := us.userRepo.GetByUsername(c, user.Username)
	if err == nil {
		return errors.New("username is already taken")
	}

	if user.Gender == "" {
		return errors.New("gender is required")
	}

	if user.Gender != "Male" && user.Gender != "Female" {
		return errors.New("gender must be either Male or Female")
	}

	if user.DoB.IsZero() {
		return errors.New("date of birth is required")
	}

	if user.DoB.After(validDate) {
		return errors.New("you must be at least 13 years old to register")
	}

	user.ProfileCompleted = true
	user.UpdatedAt = time.Now()
	user.FollowersCount = 0
	user.FollowingCount = 0
	user.PostsCount = 0

	return us.userRepo.CreateUser(c, user)
}

func (us *UserService) UpdateUser(c context.Context, user *domain.User) error {
	us.redisClient.InvalidateUserCache(user.Email, user.Username)
	return us.userRepo.UpdateUser(c, user)
}

func (us *UserService) GetByEmail(c context.Context, email string) (*domain.User, error) {
	cachedData, err := us.redisClient.GetUserCache(email)
	if err == nil && cachedData != "" {
		user := &domain.User{}
		if err := json.Unmarshal([]byte(cachedData), user); err == nil {

			return user, nil
		}
	}

	user, err := us.userRepo.GetByEmail(c, email)
	if err == nil {

		userJSON, _ := json.Marshal(user)
		us.redisClient.SetUserCache(email, string(userJSON))
	}
	return user, err
}

func (us *UserService) GetByUsername(c context.Context, username string) (*domain.User, error) {

	cachedData, err := us.redisClient.GetUserCache(username)
	if err == nil && cachedData != "" {
		user := &domain.User{}
		if err := json.Unmarshal([]byte(cachedData), user); err == nil {
			return user, nil
		}
	}

	user, err := us.userRepo.GetByUsername(c, username)
	if err == nil {
		userJSON, _ := json.Marshal(user)
		us.redisClient.SetUserCache(username, string(userJSON))
	}
	return user, err
}

func (us *UserService) LogOut(c context.Context, user *domain.User) error {
	err := us.redisClient.InvalidateUserCache(user.Email, user.Username)

	return err
}

func (us *UserService) GetByID(c context.Context, ID string) (*domain.User, error) {

	user, err := us.userRepo.GetByID(c, ID)
	if err == nil {
		userJSON, _ := json.Marshal(user)
		us.redisClient.SetUserCache(ID, string(userJSON))
	}
	return user, err
}

func (us *UserService) SearchUsers(c context.Context, query string) ([]domain.User, error) {
	return us.userRepo.SearchUsers(c, query)
}

func (us *UserService) GetAllUsers(c context.Context) ([]domain.User, error) {
	return us.userRepo.GetAllUsers(c)
}

func (us *UserService) FollowUser(c context.Context, followerID, followingID string) (*domain.User, error) {
	follower, err := us.userRepo.GetByID(c, followerID)
	if err != nil {
		return nil, errors.New("follower not found")
	}

	following, err := us.userRepo.GetByID(c, followingID)
	if err != nil {
		return nil, errors.New("user to follow not found")
	}

	if followerID == followingID {
		return nil, errors.New("cannot follow yourself")
	}

	follower.FollowingCount++
	following.FollowersCount++

	if err := us.userRepo.UpdateUser(c, follower); err != nil {
		return nil, err
	}

	if err := us.userRepo.UpdateUser(c, following); err != nil {
		return nil, err
	}

	us.redisClient.InvalidateUserCache(follower.Email, follower.Username)
	us.redisClient.InvalidateUserCache(following.Email, following.Username)

	return following, nil
}

func (us *UserService) UnfollowUser(c context.Context, followerID, followingID string) (*domain.User, error) {
	follower, err := us.userRepo.GetByID(c, followerID)
	if err != nil {
		return nil, errors.New("follower not found")
	}

	following, err := us.userRepo.GetByID(c, followingID)
	if err != nil {
		return nil, errors.New("user to unfollow not found")
	}

	if followerID == followingID {
		return nil, errors.New("cannot unfollow yourself")
	}

	if follower.FollowingCount > 0 {
		follower.FollowingCount--
	}

	if following.FollowersCount > 0 {
		following.FollowersCount--
	}

	if err := us.userRepo.UpdateUser(c, follower); err != nil {
		return nil, err
	}

	if err := us.userRepo.UpdateUser(c, following); err != nil {
		return nil, err
	}

	us.redisClient.InvalidateUserCache(follower.Email, follower.Username)
	us.redisClient.InvalidateUserCache(following.Email, following.Username)

	return following, nil
}

func (us *UserService) IsFollowing(c context.Context, followerID, followingID string) (bool, error) {
	return false, nil
}

func (us *UserService) GetFollowersCount(c context.Context, userID string) (int, error) {
	user, err := us.userRepo.GetByID(c, userID)
	if err != nil {
		return 0, err
	}
	return user.FollowersCount, nil
}

func (us *UserService) GetFollowingCount(c context.Context, userID string) (int, error) {
	user, err := us.userRepo.GetByID(c, userID)
	if err != nil {
		return 0, err
	}
	return user.FollowingCount, nil
}

func (us *UserService) GetNotificationSettings(c context.Context, userID string) (*domain.NotificationSettingsResponse, error) {
	settings, err := us.userRepo.GetNotificationSettings(c, userID)
	if err != nil {
		return nil, err
	}
	return &domain.NotificationSettingsResponse{
		PushNotification:  settings.PushNotification,
		EmailNotification: settings.EmailNotification,
	}, nil
}

func (us *UserService) UpdateNotificationSettings(c context.Context, userID string, settings *domain.NotificationSettings) error {
	return us.userRepo.UpdateNotificationSettings(c, userID, settings)
}

func (us *UserService) GetPrivacySettings(c context.Context, userID string) (*domain.PrivacySettingsResponse, error) {
	settings, err := us.userRepo.GetPrivacySettings(c, userID)
	if err != nil {
		return nil, err
	}
	return &domain.PrivacySettingsResponse{
		IsPrivate: settings.IsPrivate,
	}, nil
}

func (us *UserService) UpdatePrivacySettings(c context.Context, userID string, settings *domain.PrivacySettings) error {
	return us.userRepo.UpdatePrivacySettings(c, userID, settings)
}
