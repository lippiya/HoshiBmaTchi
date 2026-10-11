package main

import (
	_ "hoshiBmaTchi/back-end/post_service/docs"
	"hoshiBmaTchi/back-end/post_service/internal/adapter/gorm_post"
	"hoshiBmaTchi/back-end/post_service/internal/adapter/grpcclient"
	"hoshiBmaTchi/back-end/post_service/internal/adapter/http"
	"hoshiBmaTchi/back-end/post_service/internal/core/service"
	"os"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title        Post Service API
// @version      1.0
// @description  Post Service to handle posts, media uploads, and collections
// @host         localhost:8080
// @BasePath     /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token

func main() {
	db := gorm_post.GetDatabase()
	err := db.AutoMigrate(&gorm_post.PostModel{}, &gorm_post.PostMediaModel{}, &gorm_post.PostCollaboratorModel{}, &gorm_post.CollectionModel{}, &gorm_post.SavedPostModel{})
	if err != nil {
		panic("Failed to migrate database: " + err.Error())
	}

	postRepo := gorm_post.NewGormPostRepo(db)
	postService := service.NewPostService(postRepo)
	photoAddr := os.Getenv("PHOTO_SERVICE_GRPC_ADDR")
	if photoAddr == "" {
		photoAddr = "photo-service:50052"
	}
	photoClient, err := grpcclient.NewPhotoClient(photoAddr)
	if err != nil {
		panic("Failed to create photo gRPC client: " + err.Error())
	}

	postHandler := http.NewPostHandler(postService, photoClient)

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
	post := api.Group("/posts")
	post.POST("/create-post", http.AuthMiddleware(), postHandler.CreatePostHandler)
	post.POST("/upload-media", http.AuthMiddleware(), postHandler.UploadMediaHandler)
	post.POST("/create-with-paths", http.AuthMiddleware(), postHandler.CreatePostWithPathsHandler)
	post.DELETE("/delete-media", http.AuthMiddleware(), postHandler.DeleteMediaHandler)
	post.GET("/get-all-posts", http.AuthMiddleware(), postHandler.GetAllPostsHandler)
	post.GET("/get-post-media-by-id/:postID", http.AuthMiddleware(), postHandler.GetAllMediaByPostIDHandler)
	post.GET("/get-post-collaborators-by-id/:postID", http.AuthMiddleware(), postHandler.GetAllCollaboratorsByIDHandler)
	post.POST("/like-post/:postID", http.AuthMiddleware(), postHandler.LikePostHandler)
	post.GET("/check-like/:postID", http.AuthMiddleware(), postHandler.CheckLikeHandler)
	post.POST("/save-post/:postID", http.AuthMiddleware(), postHandler.SavePostHandler)
	post.GET("/check-save/:postID", http.AuthMiddleware(), postHandler.CheckSaveHandler)
	post.GET("/search-hashtags", http.AuthMiddleware(), postHandler.SearchHashtagsHandler)
	post.GET("/get-posts-by-hashtag", http.AuthMiddleware(), postHandler.GetPostsByHashtagHandler)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	post.GET("/get-post-by-id/:postID", http.AuthMiddleware(), postHandler.GetPostByIDHandler)
	post.GET("/get-posts-by-user/:userID", http.AuthMiddleware(), postHandler.GetAllPostsByUserIDHandler)
	post.GET("/get-reels-by-user/:userID", http.AuthMiddleware(), postHandler.GetAllReelsByUserIDHandler)
	post.DELETE("/delete-post/:postID", http.AuthMiddleware(), postHandler.DeletePostHandler)

	collection := api.Group("/collections")
	collection.POST("/create", http.AuthMiddleware(), postHandler.CreateCollectionHandler)
	collection.GET("/user/:userID", http.AuthMiddleware(), postHandler.GetCollectionsByUserIDHandler)
	collection.PUT("/update/:collectionID", http.AuthMiddleware(), postHandler.UpdateCollectionHandler)
	collection.DELETE("/delete/:collectionID", http.AuthMiddleware(), postHandler.DeleteCollectionHandler)
	collection.POST("/add-post", http.AuthMiddleware(), postHandler.AddPostToCollectionHandler)
	collection.DELETE("/remove-post/:collectionID/:postID", http.AuthMiddleware(), postHandler.RemovePostFromCollectionHandler)
	collection.GET("/posts/:collectionID", http.AuthMiddleware(), postHandler.GetPostsByCollectionIDHandler)

	err2 := router.Run(":8080")
	if err2 != nil {
		panic("Failed to start server: " + err2.Error())
	}

}
