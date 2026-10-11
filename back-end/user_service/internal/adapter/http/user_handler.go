package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hoshiBmaTchi/back-end/user_service/internal/adapter/gorm_user"
	"hoshiBmaTchi/back-end/user_service/internal/adapter/grpcclient"
	"hoshiBmaTchi/back-end/user_service/internal/core/domain"
	"hoshiBmaTchi/back-end/user_service/internal/core/port"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type UserHandler struct {
	userService  port.UserServicePort
	otpService   port.OTPServicePort
	oauthService port.OAuthServicePort
	spamService  port.SpamServicePort
	photoClient  port.PhotoClientPort
	socialClient *grpcclient.SocialClient
}

func NewUserHandler(
	userService port.UserServicePort,
	otpService port.OTPServicePort,
	oauthService port.OAuthServicePort,
	spamService port.SpamServicePort,
	photoClient port.PhotoClientPort,
	socialClient *grpcclient.SocialClient,
) *UserHandler {
	return &UserHandler{
		userService:  userService,
		otpService:   otpService,
		oauthService: oauthService,
		spamService:  spamService,
		photoClient:  photoClient,
		socialClient: socialClient,
	}
}

func (uh *UserHandler) LoginHandler(c *gin.Context) {
	var req gorm_user.LoginRequest
	var err = c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err = uh.spamService.Validate(req.TurnstileToken)
	if err != nil {
		c.JSON(400, gin.H{"error": "spam verification failed"})
		return
	}

	user, err2 := uh.userService.Login(c.Request.Context(), req.Email, req.Password)
	if err2 != nil {
		c.JSON(401, gin.H{"error": err2.Error()})
		return
	}

	userResponse := gorm_user.LoginResponse{
		Email:        user.Email,
		Username:     user.Username,
		Name:         user.Name,
		TwoFAEnabled: user.TwoFAEnabled,
	}

	token, err := generateJWT(user.Email, user.ID)
	if err != nil {
		log.Printf("LoginHandler: failed to generate token: %v", err)
		c.JSON(500, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "user": userResponse, "token": token})
}

func (uh *UserHandler) ResetPasswordHandler(c *gin.Context) {
	var req gorm_user.ResetPasswordRequest
	var err = c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if req.User == "" {
		c.JSON(400, gin.H{"error": "email or username is required"})
		return
	}
	if req.NewPassword == "" {
		c.JSON(400, gin.H{"error": "new password is required"})
		return
	}
	if req.ConfPassword == "" {
		c.JSON(400, gin.H{"error": "confirm password is required"})
		return
	}

	err2 := uh.userService.ResetPassword(c.Request.Context(), req.User, req.NewPassword, req.ConfPassword)
	if err2 != nil {
		c.JSON(400, gin.H{"error": err2.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
}

func (uh *UserHandler) RegisterHandler(c *gin.Context) {
	var req gorm_user.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	err := uh.spamService.Validate(req.TurnstileToken)
	if err != nil {
		c.JSON(400, gin.H{"error": "spam verification failed"})
		return
	}

	dob, err := time.Parse("2006-01-02", req.DoB)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid date of birth format"})
		return
	}

	user := &domain.User{
		Username:             req.Username,
		Name:                 req.Name,
		Email:                req.Email,
		Password:             req.Password,
		Gender:               req.Gender,
		DoB:                  dob,
		ProfilePicture:       req.ProfilePicture,
		NewsletterSubscribed: req.NewsletterSubscribed,
		TwoFAEnabled:         req.TwoFAEnabled,
	}

	err2 := uh.userService.Register(c.Request.Context(), user, req.ConfPassword)
	if err2 != nil {
		c.JSON(400, gin.H{"error": err2.Error()})
		return
	}

	_, err = uh.otpService.SendOTP(c, req.Email)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent to email"})
}

type VerifyOTPRequest struct {
	Email  string `json:"email"`
	Code   string `json:"code"`
	UserID string `json:"user_id"`
}

func (uh *UserHandler) EditProfileHandler(c *gin.Context) {
	var req gorm_user.EditProfileRequest
	var err error
	var user *domain.User
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user, err = uh.userService.GetByEmail(c.Request.Context(), email.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	err = uh.userService.EditProfile(c.Request.Context(), req.ProfilePicture, user.Email, req.Gender, req.Bio, req.Username, req.Name)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

func (uh *UserHandler) VerifyOTPHandler(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	valid, err := uh.otpService.VerifyOTP(req.Email, req.Code)
	if err != nil || !valid {
		c.JSON(400, gin.H{"error": "invalid or expired OTP"})
		return
	}
	token, err := generateJWT(req.Email, req.UserID)
	if err != nil {
		log.Printf("VerifyOTPHandler: failed to generate token: %v", err)
		c.JSON(500, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Registration successful", "token": token})
}

func generateJWT(email string, userID string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "secret"
	}

	claims := jwt.MapClaims{
		"sub":     email,
		"user_id": userID,
		"exp":     time.Now().Add(168 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid auth header"})
			return
		}

		tokenString := strings.TrimPrefix(auth, "Bearer ")
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "secret"
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if sub, ok := claims["sub"].(string); ok {
				c.Set("email", sub)
			}
		}

		c.Next()
	}
}

func (uh *UserHandler) HomeHandler(c *gin.Context) {
	email, _ := c.Get("email")
	c.JSON(http.StatusOK, gin.H{"message": "welcome", "email": email})
}

func (uh *UserHandler) GetCurrentUserHandler(c *gin.Context) {

	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := uh.userService.GetByEmail(c.Request.Context(), email.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	userResponse := gorm_user.GetUserRequest{
		ID:             user.ID,
		Email:          user.Email,
		Username:       user.Username,
		Name:           user.Name,
		TwoFAEnabled:   user.TwoFAEnabled,
		Gender:         user.Gender,
		DoB:            user.DoB.Format("2006-01-02"),
		ProfilePicture: user.ProfilePicture,
		PostsCount:     user.PostsCount,
		FollowersCount: user.FollowersCount,
		FollowingCount: user.FollowingCount,
		Bio:            user.Bio,
	}

	c.JSON(http.StatusOK, gin.H{"user": userResponse})
}

type ResendOTPRequest struct {
	Email string `json:"email"`
}

func (uh *UserHandler) ResendOTPHandler(c *gin.Context) {
	var req ResendOTPRequest

	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("ResendOTPHandler: failed to read request body: %v", err)
	} else {
		log.Printf("ResendOTPHandler: raw body=%s", string(raw))
		c.Request.Body = io.NopCloser(bytes.NewBuffer(raw))
	}

	parsed := false

	if err := json.Unmarshal(raw, &req); err == nil && req.Email != "" {
		parsed = true
	}

	if !parsed {
		if vals, err := url.ParseQuery(string(raw)); err == nil {
			if vals.Get("email") != "" {
				req.Email = vals.Get("email")
				parsed = true
			}
		}
	}

	if !parsed {
		s := strings.ReplaceAll(string(raw), "\\", "")
		s = strings.Trim(s, "\"')")
		if err := json.Unmarshal([]byte(s), &req); err == nil && req.Email != "" {
			parsed = true
		} else {
			if idx := strings.Index(s, "email:"); idx != -1 {
				start := idx + len("email:")
				rest := s[start:]
				rest = strings.Trim(rest, " {}\"')")
				if j := strings.IndexAny(rest, ",}"); j != -1 {
					rest = rest[:j]
				}
				req.Email = strings.TrimSpace(rest)
				if req.Email != "" {
					parsed = true
				}
			}
		}
	}

	if !parsed || req.Email == "" {
		log.Printf("ResendOTPHandler: unable to parse body: %s", string(raw))
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	log.Printf("ResendOTPHandler: parsed email=%s", req.Email)

	msg, err := uh.otpService.SendOTP(c.Request.Context(), req.Email)
	if err != nil {
		log.Printf("ResendOTPHandler: SendOTP error: %v", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": msg})
}

func (uh *UserHandler) ResendOTPQueryHandler(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(400, gin.H{"error": "email query param required"})
		return
	}

	msg, err := uh.otpService.SendOTP(c.Request.Context(), email)
	if err != nil {
		log.Printf("ResendOTPHandler: SendOTP error: %v", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": msg})
}

func (uh *UserHandler) GoogleLoginHandler(c *gin.Context) {
	url := uh.oauthService.GenerateAuthURL("random_state")
	c.Redirect(http.StatusFound, url)
}

func (uh *UserHandler) GoogleCallbackHandler(c *gin.Context) {
	code := c.Query("code")

	oauthData, err := uh.oauthService.GetGoogleUserInfo(code)
	if err != nil {
		c.Redirect(http.StatusFound, "/register?error=oauth_failed")
		return
	}

	user, err := uh.userService.OAuthRegister(c.Request.Context(), *oauthData)
	if err != nil {
		c.Redirect(http.StatusFound, "/register?error="+err.Error())
		return
	}

	if !user.ProfileCompleted {
		c.Redirect(http.StatusFound, "/complete-profile?email="+user.Email)
		return
	}

	c.Redirect(http.StatusFound, "/login?success=true")
}

func (uh *UserHandler) CompleteRegistrationHandler(c *gin.Context) {
	var req domain.CompleteRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	user, err := uh.userService.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(400, gin.H{"error": "user not found"})
		return
	}

	if req.ProfilePicture != "" {
		user.ProfilePicture = req.ProfilePicture
	}
	user.NewsletterSubscribed = req.NewsletterSubscribed
	user.TwoFAEnabled = req.TwoFAEnabled
	user.IsVerified = false

	err = uh.userService.UpdateUser(c.Request.Context(), user)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to update user"})
		return
	}

	if err := uh.otpService.SendWelcomeEmail(c.Request.Context(), req.Email, user.Name); err != nil {
		log.Printf("Failed to send welcome email: %v", err)

	}

	c.JSON(http.StatusOK, gin.H{"message": "Registration completed successfully"})
}

func (uh *UserHandler) UploadPictureHandler(c *gin.Context) {
	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := uh.userService.GetByEmail(c.Request.Context(), email.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}

	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file size exceeds 5MB"})
		return
	}

	contentType := file.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/jpg" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only JPEG and PNG images are allowed"})
		return
	}

	filePath, err := uh.photoClient.UploadPhoto(file, user.Email)
	if err != nil {
		log.Printf("Failed to upload photo: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload photo"})
		return
	}

	user.ProfilePicture = filePath
	err = uh.userService.UpdateUser(c.Request.Context(), user)
	if err != nil {
		log.Printf("Failed to update user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile picture"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "profile picture uploaded successfully",
		"url":     filePath,
	})
}

func (uh *UserHandler) GetPhotoHandler(c *gin.Context) {
	photoPath := c.Query("path")
	if photoPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "photo path is required"})
		return
	}

	photoURL, err := uh.photoClient.GetPhotoURL(photoPath)
	if err != nil {
		log.Printf("Failed to get photo URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get photo"})
		return
	}

	resp, err := http.Get(photoURL)
	if err != nil {
		log.Printf("Failed to fetch photo from MinIO: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch photo"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("MinIO returned status %d for photo %s", resp.StatusCode, photoPath)
		c.JSON(resp.StatusCode, gin.H{"error": "photo not found"})
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=3600")

	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		log.Printf("Failed to stream photo: %v", err)
	}
}

func (uh *UserHandler) LogoutHandler(c *gin.Context) {
	email, exists := c.Get("email")
	if !exists {
		log.Println("LogoutHandler: email not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	} else {
		log.Printf("LogoutHandler: logging out user with email %s", email.(string))
	}

	user, err := uh.userService.GetByEmail(c.Request.Context(), email.(string))
	if err != nil {
		log.Printf("User not found during logout: %v", err)
		c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
		return
	}
	err = uh.userService.LogOut(c.Request.Context(), user)
	if err != nil {
		log.Printf("Failed to invalidate user cache: %v", err)

	}
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

func (uh *UserHandler) GetUserByIDHandler(c *gin.Context) {
	userID := c.Param("id")

	user, err := uh.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (uh *UserHandler) AddFollowingCountHandler(c *gin.Context) {
	userID := c.Param("id")
	user, err := uh.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	user.FollowingCount += 1
	err = uh.userService.UpdateUser(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update following count"})
		return
	}
}

func (uh *UserHandler) DecreaseFollowingCountHandler(c *gin.Context) {
	userID := c.Param("id")
	user, err := uh.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	user.FollowingCount -= 1
	err = uh.userService.UpdateUser(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update following count"})
		return
	}
}

func (uh *UserHandler) AddFollowerCountHandler(c *gin.Context) {
	userID := c.Param("id")
	user, err := uh.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	user.FollowersCount += 1
	err = uh.userService.UpdateUser(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update follower count"})
		return
	}
}

func (uh *UserHandler) DecreaseFollowerCountHandler(c *gin.Context) {
	userID := c.Param("id")
	user, err := uh.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	user.FollowersCount -= 1
	err = uh.userService.UpdateUser(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update follower count"})
		return
	}
}

func (uh *UserHandler) AddPostCountHandler(c *gin.Context) {
	userID := c.Param("id")
	user, err := uh.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	user.PostsCount += 1
	err = uh.userService.UpdateUser(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update post count"})
		return
	}
}

func (uh *UserHandler) SearchUsersHandler(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter is required"})
		return
	}

	users, err := uh.userService.SearchUsers(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search users"})
		return
	}

	var userResponses []gorm_user.GetUserRequest
	for _, user := range users {
		userResponse := gorm_user.GetUserRequest{
			ID:             user.ID,
			Email:          user.Email,
			Username:       user.Username,
			Name:           user.Name,
			TwoFAEnabled:   user.TwoFAEnabled,
			Gender:         user.Gender,
			DoB:            user.DoB.Format("2006-01-02"),
			ProfilePicture: user.ProfilePicture,
			PostsCount:     user.PostsCount,
			FollowersCount: user.FollowersCount,
			FollowingCount: user.FollowingCount,
			Bio:            user.Bio,
		}
		userResponses = append(userResponses, userResponse)
	}

	c.JSON(http.StatusOK, gin.H{"users": userResponses})
}

func (uh *UserHandler) FollowHandler(c *gin.Context) {
	var req gorm_user.FollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	currentUser, err := uh.userService.GetByEmail(c.Request.Context(), email.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "current user not found"})
		return
	}

	targetUser, err := uh.userService.GetByID(c.Request.Context(), req.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user to follow not found"})
		return
	}

	if currentUser.ID == targetUser.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot follow yourself"})
		return
	}

	if uh.socialClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "social service not available"})
		return
	}

	resp, err := uh.socialClient.Follow(c.Request.Context(), currentUser.ID, req.UserID)
	if err != nil {
		log.Printf("Failed to follow user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to follow user"})
		return
	}

	if !resp.Success {
		log.Printf("Follow failed: %s", resp.Message)

		freshTargetUser, _ := uh.userService.GetByID(c.Request.Context(), req.UserID)
		targetFollowersCount := int(resp.FollowersCount)
		targetFollowingCount := int(resp.FollowingCount)

		if freshTargetUser != nil {
			targetFollowersCount = freshTargetUser.FollowersCount
			targetFollowingCount = freshTargetUser.FollowingCount
		}

		response := gorm_user.FollowResponse{
			Success:          false,
			Message:          resp.Message,
			IsFollowing:      true,
			FollowersCount:   targetFollowersCount,
			FollowingCount:   targetFollowingCount,
			MyFollowingCount: currentUser.FollowingCount,
		}
		c.JSON(http.StatusOK, response)
		return
	}

	currentUser.FollowingCount += 1
	err = uh.userService.UpdateUser(c.Request.Context(), currentUser)
	if err != nil {
		log.Printf("Failed to update current user following count: %v", err)
	}

	targetUser.FollowersCount += 1
	err = uh.userService.UpdateUser(c.Request.Context(), targetUser)
	if err != nil {
		log.Printf("Failed to update target user followers count: %v", err)
	}

	updatedUser, err := uh.userService.GetByID(c.Request.Context(), currentUser.ID)
	if err == nil && updatedUser != nil {
		log.Printf("Refetched user. New FollowingCount: %d", updatedUser.FollowingCount)
		currentUser = updatedUser
	} else {
		log.Printf("Failed to refetch user, using local value")
	}

	updatedTargetUser, err := uh.userService.GetByID(c.Request.Context(), req.UserID)
	if err == nil && updatedTargetUser != nil {
		log.Printf("Refetched target user. New FollowersCount: %d", updatedTargetUser.FollowersCount)
		targetUser = updatedTargetUser
	}

	response := gorm_user.FollowResponse{
		Success:          true,
		Message:          "Successfully followed user",
		IsFollowing:      true,
		FollowersCount:   targetUser.FollowersCount,
		FollowingCount:   targetUser.FollowingCount,
		MyFollowingCount: currentUser.FollowingCount,
	}

	c.JSON(http.StatusOK, response)
}

func (uh *UserHandler) UnfollowHandler(c *gin.Context) {
	var req gorm_user.FollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	currentUser, err := uh.userService.GetByEmail(c.Request.Context(), email.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "current user not found"})
		return
	}

	targetUser, err := uh.userService.GetByID(c.Request.Context(), req.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user to unfollow not found"})
		return
	}

	if currentUser.ID == targetUser.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot unfollow yourself"})
		return
	}

	if uh.socialClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "social service not available"})
		return
	}

	resp, err := uh.socialClient.Unfollow(c.Request.Context(), currentUser.ID, req.UserID)
	if err != nil {
		log.Printf("Failed to unfollow user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unfollow user"})
		return
	}

	if !resp.Success {
		log.Printf("Unfollow failed: %s", resp.Message)

		freshTargetUser, _ := uh.userService.GetByID(c.Request.Context(), req.UserID)
		targetFollowersCount := int(resp.FollowersCount)
		targetFollowingCount := int(resp.FollowingCount)

		if freshTargetUser != nil {
			targetFollowersCount = freshTargetUser.FollowersCount
			targetFollowingCount = freshTargetUser.FollowingCount
		}

		response := gorm_user.FollowResponse{
			Success:          false,
			Message:          resp.Message,
			IsFollowing:      true,
			FollowersCount:   targetFollowersCount,
			FollowingCount:   targetFollowingCount,
			MyFollowingCount: currentUser.FollowingCount,
		}
		c.JSON(http.StatusOK, response)
		return
	}

	if currentUser.FollowingCount > 0 {
		currentUser.FollowingCount -= 1
	}
	err = uh.userService.UpdateUser(c.Request.Context(), currentUser)
	if err != nil {
		log.Printf("Failed to update current user following count: %v", err)
	}

	if targetUser.FollowersCount > 0 {
		targetUser.FollowersCount -= 1
	}
	err = uh.userService.UpdateUser(c.Request.Context(), targetUser)
	if err != nil {
		log.Printf("Failed to update target user followers count: %v", err)
	}

	updatedUser, err := uh.userService.GetByID(c.Request.Context(), currentUser.ID)
	if err == nil && updatedUser != nil {
		log.Printf("Refetched user. New FollowingCount: %d", updatedUser.FollowingCount)
		currentUser = updatedUser
	} else {
		log.Printf("Failed to refetch user, using local value")
	}

	updatedTargetUser, err := uh.userService.GetByID(c.Request.Context(), req.UserID)
	if err == nil && updatedTargetUser != nil {
		targetUser = updatedTargetUser
	}

	response := gorm_user.FollowResponse{
		Success:          true,
		Message:          "Successfully unfollowed user",
		IsFollowing:      false,
		FollowersCount:   targetUser.FollowersCount,
		FollowingCount:   targetUser.FollowingCount,
		MyFollowingCount: currentUser.FollowingCount,
	}

	c.JSON(http.StatusOK, response)
}

func (uh *UserHandler) GetFollowStatusHandler(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}

	email, exists := c.Get("email")
	isFollowing := false

	if exists {
		currentUser, err := uh.userService.GetByEmail(c.Request.Context(), email.(string))
		if err == nil && uh.socialClient != nil {

			isFollowing, _ = uh.socialClient.IsFollowing(c.Request.Context(), currentUser.ID, userID)
		}
	}

	targetUser, err := uh.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	response := gorm_user.GetFollowStatusResponse{
		IsFollowing:    isFollowing,
		FollowersCount: targetUser.FollowersCount,
		FollowingCount: targetUser.FollowingCount,
	}

	c.JSON(http.StatusOK, response)
}

func (uh *UserHandler) GetSuggestedUsersHandler(c *gin.Context) {
	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	currentUser, err := uh.userService.GetByEmail(c.Request.Context(), email.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get current user"})
		return
	}

	allUsers, err := uh.userService.GetAllUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get users"})
		return
	}

	var suggestedUsers []domain.User
	for _, user := range allUsers {
		if user.ID != currentUser.ID {
			suggestedUsers = append(suggestedUsers, user)
		}
	}

	if len(suggestedUsers) > 10 {
		suggestedUsers = suggestedUsers[:10]
	}

	c.JSON(http.StatusOK, gin.H{"users": suggestedUsers})
}

func (uh *UserHandler) GetFollowingHandler(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}

	following, err := uh.socialClient.GetFollowing(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get following list"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"following": following})
}

func (uh *UserHandler) GetNotificationSettingsHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		email, _ := c.Get("email")
		user, err := uh.userService.GetByEmail(c.Request.Context(), email.(string))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID = user.ID
	}

	settings, err := uh.userService.GetNotificationSettings(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get notification settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

type UpdateNotificationSettingsRequest struct {
	PushNotification  bool `json:"push_notification"`
	EmailNotification bool `json:"email_notification"`
}

func (uh *UserHandler) UpdateNotificationSettingsHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		email, _ := c.Get("email")
		user, err := uh.userService.GetByEmail(c.Request.Context(), email.(string))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID = user.ID
	}

	var req UpdateNotificationSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings := &domain.NotificationSettings{
		PushNotification:  req.PushNotification,
		EmailNotification: req.EmailNotification,
	}

	err := uh.userService.UpdateNotificationSettings(c.Request.Context(), userID.(string), settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update notification settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification settings updated successfully"})
}

func (uh *UserHandler) GetPrivacySettingsHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		email, _ := c.Get("email")
		user, err := uh.userService.GetByEmail(c.Request.Context(), email.(string))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID = user.ID
	}

	settings, err := uh.userService.GetPrivacySettings(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get privacy settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

type UpdatePrivacySettingsRequest struct {
	IsPrivate bool `json:"is_private"`
}

func (uh *UserHandler) UpdatePrivacySettingsHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		email, _ := c.Get("email")
		user, err := uh.userService.GetByEmail(c.Request.Context(), email.(string))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID = user.ID
	}

	var req UpdatePrivacySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings := &domain.PrivacySettings{
		IsPrivate: req.IsPrivate,
	}

	err := uh.userService.UpdatePrivacySettings(c.Request.Context(), userID.(string), settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "privacy settings updated successfully"})
}
