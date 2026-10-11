package http

import (
	"encoding/json"
	"fmt"
	"hoshiBmaTchi/back-end/social_service/internal/core/domain"
	"hoshiBmaTchi/back-end/social_service/internal/core/port"
	"hoshiBmaTchi/back-end/social_service/internal/core/service"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type SocialHandler struct {
	commentService *service.CommentService
	followService  port.FollowServicePort
	socialService  *service.SocialService
}

func NewSocialHandler(commentService *service.CommentService, followService port.FollowServicePort, socialService *service.SocialService) *SocialHandler {
	return &SocialHandler{
		commentService: commentService,
		followService:  followService,
		socialService:  socialService,
	}
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
			if uid, ok := claims["user_id"].(string); ok {
				c.Set("user_id", uid)
			}
		}

		c.Next()
	}
}

func (sh *SocialHandler) CreateCommentHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	postID := c.PostForm("post_id")
	parentCommentID := c.PostForm("parent_comment_id")
	text := c.PostForm("text")
	gifUrl := c.PostForm("gif_url")

	if postID == "" && parentCommentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "post_id or parent_comment_id is required"})
		return
	}

	if text == "" && gifUrl == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text or gif_url is required"})
		return
	}

	var parentIDPtr *string
	if parentCommentID != "" {
		parentIDPtr = &parentCommentID
	}

	comment, err := sh.commentService.CreateComment(domain.CreateCommentRequest{
		PostID:          postID,
		ParentCommentID: parentIDPtr,
		UserID:          userID.(string),
		Text:            text,
		GifUrl:          gifUrl,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comment": comment,
	})
}

func (sh *SocialHandler) GetCommentsHandler(c *gin.Context) {
	postID := c.Param("post_id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "post_id is required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	sort := c.DefaultQuery("sort", "created")

	comments, err := sh.commentService.GetComments(postID, page, limit, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": comments,
	})
}

func (sh *SocialHandler) GetCommentRepliesHandler(c *gin.Context) {
	commentID := c.Param("comment_id")
	if commentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comment_id is required"})
		return
	}

	replies, err := sh.commentService.GetCommentReplies(commentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"replies": replies,
	})
}

func (sh *SocialHandler) DeleteCommentHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	commentID := c.Param("comment_id")
	if commentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comment_id is required"})
		return
	}

	err := sh.commentService.DeleteComment(commentID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "comment deleted successfully",
	})
}

func (sh *SocialHandler) LikeCommentHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	commentID := c.Param("comment_id")
	if commentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comment_id is required"})
		return
	}

	liked, err := sh.commentService.LikeComment(commentID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	message := "comment liked"
	if !liked {
		message = "comment unliked"
	}

	c.JSON(http.StatusOK, gin.H{
		"liked":   liked,
		"message": message,
	})
}

func (sh *SocialHandler) CheckUserCommentedHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	postID := c.Param("post_id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "post_id is required"})
		return
	}

	commented, err := sh.commentService.CheckUserCommented(postID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"commented": commented,
	})
}

func (sh *SocialHandler) GetCommentsCountHandler(c *gin.Context) {
	postID := c.Param("post_id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "post_id is required"})
		return
	}

	count, err := sh.commentService.GetCommentsCount(postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": count,
	})
}

func (sh *SocialHandler) GetFollowingHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	following, total, err := sh.followService.GetFollowingList(c.Request.Context(), userID.(string), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get following list"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"following": following,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

func (sh *SocialHandler) GetCloseFriendsHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	closeFriends, err := sh.socialService.GetCloseFriends(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get close friends"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"close_friends": closeFriends})
}

func (sh *SocialHandler) AddCloseFriendHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	friendID := c.Param("user_id")
	if friendID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	err := sh.socialService.AddCloseFriend(c.Request.Context(), userID.(string), friendID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add close friend"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "close friend added successfully"})
}

func (sh *SocialHandler) RemoveCloseFriendHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	friendID := c.Param("user_id")
	if friendID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	err := sh.socialService.RemoveCloseFriend(c.Request.Context(), userID.(string), friendID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove close friend"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "close friend removed successfully"})
}

func (sh *SocialHandler) GetBlockedUsersHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	blockedUsers, err := sh.socialService.GetBlockedUsers(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get blocked users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"blocked_users": blockedUsers})
}

func (sh *SocialHandler) GetHiddenStoryViewersHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	hiddenViewers, err := sh.socialService.GetHiddenStoryViewers(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get hidden story viewers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"hidden_viewers": hiddenViewers})
}

func (sh *SocialHandler) HideStoryViewerHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	hiddenID := c.Param("user_id")
	if hiddenID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	err := sh.socialService.HideStoryViewer(c.Request.Context(), userID.(string), hiddenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hide story viewer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "story viewer hidden successfully"})
}

func (sh *SocialHandler) UnhideStoryViewerHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	hiddenID := c.Param("user_id")
	if hiddenID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	err := sh.socialService.UnhideStoryViewer(c.Request.Context(), userID.(string), hiddenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unhide story viewer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "story viewer unhidden successfully"})
}

func (sh *SocialHandler) BlockUserHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	blockedID := c.Param("user_id")
	if blockedID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	err := sh.socialService.BlockUser(c.Request.Context(), userID.(string), blockedID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to block user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user blocked successfully"})
}

func (sh *SocialHandler) UnblockUserHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	blockedID := c.Param("user_id")
	if blockedID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	err := sh.socialService.UnblockUser(c.Request.Context(), userID.(string), blockedID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unblock user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user unblocked successfully"})
}

func (sh *SocialHandler) CheckBlockStatusHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	targetID := c.Param("user_id")
	if targetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	isBlocked, err := sh.socialService.CheckBlockStatus(c.Request.Context(), userID.(string), targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check block status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_blocked": isBlocked})
}

func SetupSocialRoutes(router *gin.Engine, commentService *service.CommentService, followService port.FollowServicePort, socialService *service.SocialService, notificationService *service.NotificationService) {
	handler := NewSocialHandler(commentService, followService, socialService)
	notificationHandler := NewNotificationHandler(notificationService)

	social := router.Group("/api/v1/social")
	social.Use(AuthMiddleware())
	{
		social.POST("/create-comment", handler.CreateCommentHandler)
		social.GET("/get-comments/:post_id", handler.GetCommentsHandler)
		social.GET("/get-comment-replies/:comment_id", handler.GetCommentRepliesHandler)
		social.DELETE("/delete-comment/:comment_id", handler.DeleteCommentHandler)
		social.POST("/like-comment/:comment_id", handler.LikeCommentHandler)
		social.GET("/check-user-commented/:post_id", handler.CheckUserCommentedHandler)
		social.GET("/get-comments-count/:post_id", handler.GetCommentsCountHandler)
		social.GET("/get-following", handler.GetFollowingHandler)
		social.GET("/get-close-friends", handler.GetCloseFriendsHandler)
		social.POST("/add-close-friend/:user_id", handler.AddCloseFriendHandler)
		social.DELETE("/remove-close-friend/:user_id", handler.RemoveCloseFriendHandler)
		social.GET("/get-blocked-users", handler.GetBlockedUsersHandler)
		social.GET("/get-hidden-story-viewers", handler.GetHiddenStoryViewersHandler)
		social.POST("/hide-story-viewer/:user_id", handler.HideStoryViewerHandler)
		social.DELETE("/unhide-story-viewer/:user_id", handler.UnhideStoryViewerHandler)
		social.POST("/block-user/:user_id", handler.BlockUserHandler)
		social.DELETE("/unblock-user/:user_id", handler.UnblockUserHandler)
		social.GET("/check-block-status/:user_id", handler.CheckBlockStatusHandler)

		social.GET("/get-notifications", notificationHandler.GetNotificationsHandler)
		social.GET("/notifications/unread-count", notificationHandler.GetUnreadCountHandler)
		social.PUT("/mark-notification-read/:id", notificationHandler.MarkAsReadHandler)
		social.PUT("/mark-all-notifications-read", notificationHandler.MarkAllAsReadHandler)
		social.DELETE("/delete-notification/:id", notificationHandler.DeleteNotificationHandler)
		social.POST("/create-notification", notificationHandler.CreateNotificationHandler)
	}
}

func RegisterSocialRoutes(router *gin.Engine, commentService *service.CommentService, followService port.FollowServicePort, socialService *service.SocialService, notificationService *service.NotificationService) {
	SetupSocialRoutes(router, commentService, followService, socialService, notificationService)
}

type CommentResponse struct {
	ID              string          `json:"ID"`
	PostID          string          `json:"PostID"`
	ParentCommentID *string         `json:"ParentCommentID"`
	UserID          string          `json:"UserID"`
	Text            string          `json:"Text"`
	GifUrl          string          `json:"GifUrl"`
	LikesCount      int             `json:"LikesCount"`
	RepliesCount    int             `json:"RepliesCount"`
	CreatedAt       string          `json:"CreatedAt"`
	User            json.RawMessage `json:"user,omitempty"`
}
