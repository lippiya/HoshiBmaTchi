package main

import (
	"log"
	"net/http"
	"os"
	"story_service/internal/adapter/gorm_story"
	"story_service/internal/adapter/grpcclient"
	httpHandler "story_service/internal/adapter/http"
	"story_service/internal/core/service"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	db := gorm_story.GetDB()

	photoAddr := os.Getenv("PHOTO_SERVICE_GRPC_ADDR")
	if photoAddr == "" {
		photoAddr = "photo-service:50053"
	}
	photoClient, err := grpcclient.NewPhotoClient(photoAddr)
	if err != nil {
		log.Fatalf("Failed to create photo gRPC client: %v", err)
	}

	repo := gorm_story.NewStoryRepository(db)
	storyService := service.NewStoryService(repo)
	handler := httpHandler.NewStoryHandler(storyService, photoClient)

	go cleanupExpiredStories(storyService)

	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api/v1/stories").Subrouter()
	handler.RegisterRoutes(apiRouter)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://localhost", "https://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	handlerWithCORS := c.Handler(router)

	log.Println("Story service starting on :8084")
	if err := http.ListenAndServe(":8084", handlerWithCORS); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func cleanupExpiredStories(storyService *service.StoryService) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		if err := storyService.CleanupExpiredStories(); err != nil {
			log.Printf("Failed to cleanup expired stories: %v", err)
		} else {
			log.Println("Expired stories cleaned up successfully")
		}
	}
}
