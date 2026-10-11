package main

import (
	"hoshiBmaTchi/back-end/proto"
	_ "hoshiBmaTchi/back-end/social_service/docs"
	"hoshiBmaTchi/back-end/social_service/internal/adapter"
	"hoshiBmaTchi/back-end/social_service/internal/adapter/gorm_user"
	httpHandler "hoshiBmaTchi/back-end/social_service/internal/adapter/http"
	"hoshiBmaTchi/back-end/social_service/internal/core/service"
	"log"
	"net"
	"os"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
)

//@title        Social Service API
//@version      1.0
//@description  Social Service to handle follows, comments, and likes
//@host         localhost:8085
//@BasePath     /api/v1

func main() {
	db := gorm_user.GetDatabase()

	err := db.AutoMigrate(&gorm_user.FollowModel{}, &gorm_user.CommentModel{}, &gorm_user.CommentLikeModel{}, &gorm_user.CloseFriendModel{}, &gorm_user.BlockedUserModel{}, &gorm_user.HiddenStoryViewerModel{}, &gorm_user.NotificationModel{})
	if err != nil {
		log.Fatalf("Failed to migrate models: %v", err)
	}

	followRepo := gorm_user.NewFollowRepository(db)
	followService := service.NewFollowService(followRepo)

	commentRepo := gorm_user.NewCommentRepository(db)
	commentService := service.NewCommentService(commentRepo)

	socialRepo := gorm_user.NewSocialRepository(db)
	socialService := service.NewSocialService(socialRepo)

	notificationRepo := gorm_user.NewNotificationRepository(db)
	notificationService := service.NewNotificationService(notificationRepo)

	go func() {
		grpcServer := grpc.NewServer()
		socialGRPCServer := adapter.NewSocialGRPCServer(followService, commentService)

		proto.RegisterSocialServiceServer(grpcServer, socialGRPCServer)

		listener, err := net.Listen("tcp", ":50052")
		if err != nil {
			log.Fatalf("Failed to listen on port 50052: %v", err)
		}

		log.Println("Social service gRPC server running on port 50052")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

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

	httpHandler.RegisterSocialRoutes(router, commentService, followService, socialService, notificationService)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Println("Social service HTTP server running on port 8085")
	if err := router.Run(":8085"); err != nil {
		log.Fatalf("Failed to serve HTTP: %v", err)
	}
}
