package http

import (
	"encoding/json"
	"fmt"
	"hoshiBmaTchi/back-end/message_service/internal/core/domain"
	"hoshiBmaTchi/back-end/message_service/internal/core/port"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
)

type MessageHandler struct {
	MessageService port.MessageServicePort
	photoClient    port.PhotoClientPort
	wsConnections  map[string]map[*websocket.Conn]bool
	wsMutex        sync.RWMutex
	upgrader       websocket.Upgrader
}

func NewMessageHandler(messageService port.MessageServicePort, photoClient port.PhotoClientPort) *MessageHandler {
	return &MessageHandler{
		MessageService: messageService,
		photoClient:    photoClient,
		wsConnections:  make(map[string]map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
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

func (mh *MessageHandler) GetConversationsHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversations, err := mh.MessageService.GetConversationsByUserID(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, conversations)
}

func (mh *MessageHandler) SearchConversationsHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	query := c.Query("q")
	if query == "" {
		mh.GetConversationsHandler(c)
		return
	}

	conversations, err := mh.MessageService.SearchConversations(c.Request.Context(), userID.(string), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, conversations)
}

func (mh *MessageHandler) GetConversationHandler(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationID := c.Param("conversationID")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id is required"})
		return
	}

	conversation, err := mh.MessageService.GetConversationByID(c.Request.Context(), conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	messages, err := mh.MessageService.GetMessagesByConversationID(c.Request.Context(), conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	participants, err := mh.MessageService.GetParticipantsByConversationID(c.Request.Context(), conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conversation": conversation,
		"messages":     messages,
		"participants": participants,
	})
}

func (mh *MessageHandler) CreateConversationHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		ParticipantIDs []string `json:"participant_ids"`
		IsGroup        bool     `json:"is_group"`
		GroupName      string   `json:"group_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.ParticipantIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "participant_ids is required"})
		return
	}

	if !contains(req.ParticipantIDs, userID.(string)) {
		req.ParticipantIDs = append(req.ParticipantIDs, userID.(string))
	}

	if !req.IsGroup && len(req.ParticipantIDs) == 2 {
		otherUserID := ""
		for _, id := range req.ParticipantIDs {
			if id != userID.(string) {
				otherUserID = id
				break
			}
		}

		conversation, err := mh.MessageService.GetOrCreateDirectConversation(c.Request.Context(), userID.(string), otherUserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, conversation)
		return
	}

	conversation := &domain.Conversation{
		IsGroup:   req.IsGroup,
		GroupName: req.GroupName,
	}

	if err := mh.MessageService.CreateConversation(c.Request.Context(), conversation, req.ParticipantIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, conversation)
}

func (mh *MessageHandler) DeleteConversationHandler(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationID := c.Param("conversationID")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id is required"})
		return
	}

	if err := mh.MessageService.DeleteConversation(c.Request.Context(), conversationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "conversation deleted"})
}

func (mh *MessageHandler) SendMessageHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationID := c.Param("conversationID")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id is required"})
		return
	}

	err := c.Request.ParseMultipartForm(10 << 20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse form"})
		return
	}

	content := c.PostForm("content")
	mediaType := c.PostForm("media_type")

	message := &domain.Message{
		ConversationID: conversationID,
		SenderID:       userID.(string),
		Content:        content,
		MediaType:      mediaType,
	}

	file, fileHeader, err := c.Request.FormFile("media")
	if err == nil {
		defer file.Close()

		email, _ := c.Get("email")
		filePath, err := mh.photoClient.UploadPhoto(fileHeader, email.(string))
		if err != nil {
			log.Printf("Failed to upload media: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload media"})
			return
		}
		message.MediaURL = filePath
	}

	if err := mh.MessageService.SendMessage(c.Request.Context(), message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mh.broadcastToConversation(conversationID, message)

	c.JSON(http.StatusCreated, message)
}

func (mh *MessageHandler) UnsendMessageHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationID := c.Param("conversationID")
	messageID := c.Param("messageID")

	if conversationID == "" || messageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id and message_id are required"})
		return
	}

	if err := mh.MessageService.UnsendMessage(c.Request.Context(), conversationID, messageID, userID.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mh.broadcastToConversation(conversationID, map[string]string{
		"type":       "delete",
		"message_id": messageID,
	})

	c.JSON(http.StatusOK, gin.H{"message": "message unsent"})
}

func (mh *MessageHandler) MarkAsSeenHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationID := c.Param("conversationID")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id is required"})
		return
	}

	if err := mh.MessageService.MarkAsSeen(c.Request.Context(), conversationID, userID.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "marked as seen"})
}

func (mh *MessageHandler) AddParticipantsHandler(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationID := c.Param("conversationID")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id is required"})
		return
	}

	var req struct {
		ParticipantIDs []string `json:"participant_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := mh.MessageService.AddParticipants(c.Request.Context(), conversationID, req.ParticipantIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "participants added"})
}

func (mh *MessageHandler) RemoveParticipantHandler(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationID := c.Param("conversationID")
	participantID := c.Param("participantID")

	if conversationID == "" || participantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id and participant_id are required"})
		return
	}

	if err := mh.MessageService.RemoveParticipant(c.Request.Context(), conversationID, participantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "participant removed"})
}

func (mh *MessageHandler) SearchMessagesHandler(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conversationID := c.Param("conversationID")
	query := c.Query("q")

	if conversationID == "" || query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id and query are required"})
		return
	}

	messages, err := mh.MessageService.SearchMessagesInConversation(c.Request.Context(), conversationID, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

func (mh *MessageHandler) WebSocketHandler(c *gin.Context) {
	conversationID := c.Query("conversation_id")
	tokenString := c.Query("token")

	if conversationID == "" || tokenString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id and token are required"})
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "secret"
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := mh.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	mh.wsMutex.Lock()
	if mh.wsConnections[conversationID] == nil {
		mh.wsConnections[conversationID] = make(map[*websocket.Conn]bool)
	}
	mh.wsConnections[conversationID][conn] = true
	mh.wsMutex.Unlock()

	defer func() {
		mh.wsMutex.Lock()
		delete(mh.wsConnections[conversationID], conn)
		if len(mh.wsConnections[conversationID]) == 0 {
			delete(mh.wsConnections, conversationID)
		}
		mh.wsMutex.Unlock()
		conn.Close()
	}()

	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-done:
				return
			}
		}
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			close(done)
			break
		}
	}
}

func (mh *MessageHandler) broadcastToConversation(conversationID string, data interface{}) {
	mh.wsMutex.RLock()
	defer mh.wsMutex.RUnlock()

	connections, exists := mh.wsConnections[conversationID]
	if !exists {
		return
	}

	messageJSON, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	for conn := range connections {
		if err := conn.WriteMessage(websocket.TextMessage, messageJSON); err != nil {
			log.Printf("Failed to send message: %v", err)
			conn.Close()
			delete(connections, conn)
		}
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
