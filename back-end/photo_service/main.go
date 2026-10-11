package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	grpcadapter "hoshiBmaTchi/back-end/photo_service/internal/adapter"
	httpadapter "hoshiBmaTchi/back-end/photo_service/internal/adapter/http"
	minioadapter "hoshiBmaTchi/back-end/photo_service/internal/adapter/minio"
	"hoshiBmaTchi/back-end/photo_service/internal/core/service"
	pb "hoshiBmaTchi/back-end/proto"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {

	requiredEnvVars := []string{"MINIO_ACCESS_KEY", "MINIO_SECRET_KEY"}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("Required environment variable %s is not set", envVar)
		}
	}

	minioEndpoint := getEnv("MINIO_ENDPOINT", "minio:9000")
	minioPublicEndpoint := getEnv("MINIO_PUBLIC_ENDPOINT", "localhost:9000")
	minioAccessKeyID := getEnv("MINIO_ACCESS_KEY", "")
	minioSecretKey := getEnv("MINIO_SECRET_KEY", "")
	minioBucket := getEnv("MINIO_BUCKET", "photos")
	grpcPort := getEnv("GRPC_PORT", "50053")

	useSSL := getEnvBool("MINIO_USE_SSL", false)

	maxFileSize := getEnvInt64("MAX_FILE_SIZE", 10*1024*1024)

	minioConfig := minioadapter.MinioConfig{
		Endpoint:       minioEndpoint,
		AccessKey:      minioAccessKeyID,
		SecretKey:      minioSecretKey,
		BucketName:     minioBucket,
		PublicEndpoint: minioPublicEndpoint,
		UseSSL:         useSSL,
		MaxFileSize:    maxFileSize,
	}

	minioClient, err := minioadapter.NewMinioClientWithConfig(minioConfig)
	if err != nil {
		log.Fatalf("Failed to initialize MinIO client: %v", err)
	}

	photoService := service.NewPhotoService(minioClient)

	go func() {
		grpcServer := grpc.NewServer()
		photoServer := grpcadapter.NewPhotoServer(photoService)
		pb.RegisterPhotoServiceServer(grpcServer, photoServer)

		listener, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
		if err != nil {
			log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
		}

		log.Printf("Photo service gRPC server listening on port %s", grpcPort)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	router := gin.Default()

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

	httpadapter.SetupPhotoHTTPRoutes(router, photoService)

	httpPort := getEnv("HTTP_PORT", "8086")
	log.Printf("Photo service HTTP server listening on port %s", httpPort)
	if err := router.Run(fmt.Sprintf(":%s", httpPort)); err != nil {
		log.Fatalf("Failed to serve HTTP: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		log.Printf("Warning: invalid boolean value for %s, using default: %v", key, defaultValue)
		return defaultValue
	}
	return boolValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log.Printf("Warning: invalid integer value for %s, using default: %d", key, defaultValue)
		return defaultValue
	}
	return intValue
}
