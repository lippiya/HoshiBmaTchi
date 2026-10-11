package main

import (
	_ "hoshiBmaTchi/back-end/user_service/docs"
	"hoshiBmaTchi/back-end/user_service/internal/adapter/gorm_user"
	"hoshiBmaTchi/back-end/user_service/internal/adapter/grpcclient"
	"hoshiBmaTchi/back-end/user_service/internal/adapter/http"
	"hoshiBmaTchi/back-end/user_service/internal/adapter/redis"
	"hoshiBmaTchi/back-end/user_service/internal/core/domain"
	"hoshiBmaTchi/back-end/user_service/internal/core/service"
	"log"
	"os"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//@title        User Service API
//@version      1.0
//@description  User Service to handle user registration, authentication, and profile management
//@host         localhost:8080
//@BasePath     /api/v1

func main() {

	db := gorm_user.GetDatabase()

	err := db.AutoMigrate(&domain.User{}, &gorm_user.NotificationSettings{}, &gorm_user.PrivacySettings{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	userRepo := gorm_user.NewGormUserRepo(db)
	redisClient := redis.NewRedisClient()
	userService := service.NewUserService(userRepo, redisClient)

	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://guest:guest@localhost:5672/"
	}

	var rabbitmqChannel *amqp.Channel
	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		log.Printf("Failed to connect to RabbitMQ: %v. OTP emails will not be sent.", err)
	} else {

		rabbitmqChannel, err = conn.Channel()
		if err != nil {
			log.Printf("Failed to open RabbitMQ channel: %v. OTP emails will not be sent.", err)
		} else {

			_, err = rabbitmqChannel.QueueDeclare(
				"queue",
				true,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				log.Printf("Failed to declare queue: %v", err)
			} else {
				log.Println("RabbitMQ queue declared successfully")
			}
		}
	}

	var emailGRPCClient *grpcclient.Client
	emailAddr := os.Getenv("EMAIL_SERVICE_GRPC_ADDR")
	if emailAddr == "" {
		emailAddr = "email-service:50051"
	}
	emailGRPCClient, err = grpcclient.New(emailAddr)
	if err != nil {
		log.Printf("Failed to create email gRPC client: %v. Will use RabbitMQ fallback.", err)
		emailGRPCClient = nil
	}

	photoAddr := os.Getenv("PHOTO_SERVICE_GRPC_ADDR")
	if photoAddr == "" {
		photoAddr = "photo-service:50053"
	}
	photoClient, err := grpcclient.NewPhotoClient(photoAddr)
	if err != nil {
		log.Fatalf("Failed to create photo gRPC client: %v", err)
	}

	socialAddr := os.Getenv("SOCIAL_SERVICE_GRPC_ADDR")
	if socialAddr == "" {
		socialAddr = "social-service:50052"
	}
	socialClient, err := grpcclient.NewSocialClient(socialAddr)
	if err != nil {
		log.Printf("Failed to create social gRPC client: %v. Follow features may not work.", err)
	}

	otpService := service.NewOTPService(redisClient, emailGRPCClient, rabbitmqChannel)
	oauthService := service.NewOAuthService()
	spamService := service.NewSpamService()

	userHandler := http.NewUserHandler(userService, otpService, oauthService, spamService, photoClient, socialClient)

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
	user := api.Group("/users")
	user.POST("/register", userHandler.RegisterHandler)
	user.POST("/login", userHandler.LoginHandler)
	user.POST("/verify-otp", userHandler.VerifyOTPHandler)
	user.POST("/send-otp", userHandler.ResendOTPHandler)
	user.GET("/send-otp", userHandler.ResendOTPQueryHandler)
	user.POST("/complete-registration", userHandler.CompleteRegistrationHandler)
	user.GET("/", http.AuthMiddleware(), userHandler.HomeHandler)
	user.GET("/get-current-user", http.AuthMiddleware(), userHandler.GetCurrentUserHandler)
	user.POST("/upload-picture", http.AuthMiddleware(), userHandler.UploadPictureHandler)
	user.GET("/photo", userHandler.GetPhotoHandler)
	user.GET("/logout", http.AuthMiddleware(), userHandler.LogoutHandler)
	user.POST("/reset-password", userHandler.ResetPasswordHandler)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	user.PUT("/edit-profile", http.AuthMiddleware(), userHandler.EditProfileHandler)
	user.GET("/get-user-by-id/:id", userHandler.GetUserByIDHandler)
	user.POST("/add-post-count/:id", userHandler.AddPostCountHandler)
	user.POST("/add-follower/:id", userHandler.AddFollowerCountHandler)
	user.POST("/add-following/:id", userHandler.AddFollowingCountHandler)
	user.POST("/decrease-follower/:id", userHandler.DecreaseFollowerCountHandler)
	user.POST("/decrease-following/:id", userHandler.DecreaseFollowingCountHandler)
	user.GET("/search", http.AuthMiddleware(), userHandler.SearchUsersHandler)

	user.POST("/follow", http.AuthMiddleware(), userHandler.FollowHandler)
	user.POST("/unfollow", http.AuthMiddleware(), userHandler.UnfollowHandler)
	user.GET("/follow-status/:userId", http.AuthMiddleware(), userHandler.GetFollowStatusHandler)
	user.GET("/get-following/:userId", http.AuthMiddleware(), userHandler.GetFollowingHandler)
	user.GET("/get-suggested-users", http.AuthMiddleware(), userHandler.GetSuggestedUsersHandler)
	user.GET("/get-notification-settings", http.AuthMiddleware(), userHandler.GetNotificationSettingsHandler)
	user.PUT("/update-notification-settings", http.AuthMiddleware(), userHandler.UpdateNotificationSettingsHandler)
	user.GET("/get-privacy-settings", http.AuthMiddleware(), userHandler.GetPrivacySettingsHandler)
	user.PUT("/update-privacy-settings", http.AuthMiddleware(), userHandler.UpdatePrivacySettingsHandler)

	err2 := router.Run(":8080")
	if err2 != nil {
		log.Fatalf("Failed to run server: %v", err2)
	}

}
