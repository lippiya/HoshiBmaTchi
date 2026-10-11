package main

import (
	"hoshiBmaTchi/back-end/message_service/internal/adapter/gorm_message"
	"hoshiBmaTchi/back-end/message_service/internal/adapter/grpcclient"
	"hoshiBmaTchi/back-end/message_service/internal/adapter/http"
	"hoshiBmaTchi/back-end/message_service/internal/core/service"
	"os"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//@title        Message Service API
//@version      1.0
//@description  Message Service to handle conversations and messaging
//@host         localhost:8080
//@BasePath     /api/v1

func main() {
	db := gorm_message.GetDatabase()
	err := db.AutoMigrate(&gorm_message.ConversationModel{}, &gorm_message.MessageModel{}, &gorm_message.ConversationParticipantModel{})
	if err != nil {
		panic("Failed to migrate database: " + err.Error())
	}

	messageRepo := gorm_message.NewGormMessageRepo(db)
	messageService := service.NewMessageService(messageRepo)

	photoAddr := os.Getenv("PHOTO_SERVICE_GRPC_ADDR")
	if photoAddr == "" {
		photoAddr = "photo-service:50052"
	}
	photoClient, err := grpcclient.NewPhotoClient(photoAddr)
	if err != nil {
		panic("Failed to create photo gRPC client: " + err.Error())
	}

	messageHandler := http.NewMessageHandler(messageService, photoClient)

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "INFO"
	}

	if logLevel == "DEBUG" {
		gin.SetMode(gin.DebugMode)
	} else if logLevel == "RELEASE" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.TestMode)
	}

	router := gin.Default()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	api := router.Group("/api/v1/")
	messages := api.Group("/messages")

	messages.GET("/ws", messageHandler.WebSocketHandler)

	messages.Use(http.AuthMiddleware())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	messages.GET("/conversations", messageHandler.GetConversationsHandler)
	messages.GET("/conversations/search", messageHandler.SearchConversationsHandler)
	messages.GET("/conversations/:conversationID", messageHandler.GetConversationHandler)
	messages.POST("/conversations", messageHandler.CreateConversationHandler)
	messages.DELETE("/conversations/:conversationID", messageHandler.DeleteConversationHandler)

	messages.POST("/conversations/:conversationID/messages", messageHandler.SendMessageHandler)
	messages.DELETE("/conversations/:conversationID/messages/:messageID", messageHandler.UnsendMessageHandler)
	messages.PUT("/conversations/:conversationID/seen", messageHandler.MarkAsSeenHandler)
	messages.GET("/conversations/:conversationID/search", messageHandler.SearchMessagesHandler)

	messages.POST("/conversations/:conversationID/participants", messageHandler.AddParticipantsHandler)
	messages.DELETE("/conversations/:conversationID/participants/:participantID", messageHandler.RemoveParticipantHandler)

	err2 := router.Run(":8080")
	if err2 != nil {
		panic("Failed to start server: " + err2.Error())
	}
}
