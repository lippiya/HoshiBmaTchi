package http

import (
	"hoshiBmaTchi/back-end/social_service/internal/core/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notificationService *service.NotificationService
}

func NewNotificationHandler(notificationService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// GetNotificationsHandler godoc
// @Summary Get user notifications
// @Description Get paginated list of notifications for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /social/get-notifications [get]
func (nh *NotificationHandler) GetNotificationsHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	notifications, total, err := nh.notificationService.GetNotifications(c.Request.Context(), userID.(string), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get notifications"})
		return
	}

	unreadCount, _ := nh.notificationService.GetUnreadCount(c.Request.Context(), userID.(string))

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"total":         total,
		"page":          page,
		"limit":         limit,
		"unread_count":  unreadCount,
	})
}

// GetUnreadCountHandler godoc
// @Summary Get unread notification count
// @Description Get the count of unread notifications for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]int64
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /social/notifications/unread-count [get]
func (nh *NotificationHandler) GetUnreadCountHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	count, err := nh.notificationService.GetUnreadCount(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get unread count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

// MarkAsReadHandler godoc
// @Summary Mark notification as read
// @Description Mark a specific notification as read
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /social/mark-notification-read/{id} [put]
func (nh *NotificationHandler) MarkAsReadHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notification id is required"})
		return
	}

	err := nh.notificationService.MarkAsRead(c.Request.Context(), notificationID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification marked as read"})
}

// MarkAllAsReadHandler godoc
// @Summary Mark all notifications as read
// @Description Mark all notifications as read for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /social/mark-all-notifications-read [put]
func (nh *NotificationHandler) MarkAllAsReadHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err := nh.notificationService.MarkAllAsRead(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark all notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "all notifications marked as read"})
}

// DeleteNotificationHandler godoc
// @Summary Delete a notification
// @Description Delete a specific notification
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /social/delete-notification/{id} [delete]
func (nh *NotificationHandler) DeleteNotificationHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notification id is required"})
		return
	}

	err := nh.notificationService.DeleteNotification(c.Request.Context(), notificationID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification deleted"})
}

// CreateNotificationHandler godoc
// @Summary Create a notification (internal use)
// @Description Create a new notification - typically used by other services
// @Tags notifications
// @Accept json
// @Produce json
// @Param notification body CreateNotificationRequest true "Notification data"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /social/create-notification [post]
func (nh *NotificationHandler) CreateNotificationHandler(c *gin.Context) {
	var req CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.RecipientID == "" || req.ActorID == "" || req.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "recipient_id, actor_id, and type are required"})
		return
	}

	validTypes := map[string]bool{"like": true, "comment": true, "follow": true, "mention": true}
	if !validTypes[req.Type] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification type"})
		return
	}

	err := nh.notificationService.CreateNotification(
		c.Request.Context(),
		req.RecipientID,
		req.ActorID,
		req.Type,
		req.PostID,
		req.CommentText,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification created"})
}

type CreateNotificationRequest struct {
	RecipientID string  `json:"recipient_id"`
	ActorID     string  `json:"actor_id"`
	Type        string  `json:"type"`
	PostID      *string `json:"post_id,omitempty"`
	CommentText *string `json:"comment_text,omitempty"`
}
