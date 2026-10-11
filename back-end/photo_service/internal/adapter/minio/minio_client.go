package minio

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"hoshiBmaTchi/back-end/photo_service/internal/core/domain"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	MaxFileSize        = 100 * 1024 * 1024 
	PresignedURLExpiry = 15 * time.Minute
	MaxPresignedExpiry = 7 * 24 * time.Hour
)

var (
	AllowedContentTypes = map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"video/mp4":  true,
		"video/webm": true,
		"video/quicktime": true,
	}

	AllowedExtensions = map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".mp4":  true,
		".webm": true,
		".mov":  true,
	}
)

type MinioClient struct {
	client         *minio.Client
	bucketName     string
	endpoint       string
	publicEndpoint string
	maxFileSize    int64
	useSSL         bool
}

type MinioConfig struct {
	Endpoint       string
	AccessKey      string
	SecretKey      string
	BucketName     string
	PublicEndpoint string
	UseSSL         bool
	MaxFileSize    int64
}

func NewMinioClient(endpoint, accessKey, secretKey, bucketName, publicEndpoint string, useSSL bool) (*MinioClient, error) {
	config := MinioConfig{
		Endpoint:       endpoint,
		AccessKey:      accessKey,
		SecretKey:      secretKey,
		BucketName:     bucketName,
		PublicEndpoint: publicEndpoint,
		UseSSL:         useSSL,
		MaxFileSize:    MaxFileSize,
	}
	return NewMinioClientWithConfig(config)
}

func NewMinioClientWithConfig(config MinioConfig) (*MinioClient, error) {

	if config.Endpoint == "" || config.AccessKey == "" || config.SecretKey == "" || config.BucketName == "" {
		return nil, fmt.Errorf("invalid MinIO configuration: missing required fields")
	}

	if len(config.AccessKey) < 3 || len(config.SecretKey) < 8 {
		return nil, fmt.Errorf("invalid credentials: access key must be at least 3 characters and secret key at least 8 characters")
	}

	minioClient, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx := context.Background()

	exists, err := minioClient.BucketExists(ctx, config.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {

		err = minioClient.MakeBucket(ctx, config.BucketName, minio.MakeBucketOptions{
			Region:        "us-east-1",
			ObjectLocking: false,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}

		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Deny",
					"Principal": {"AWS": ["*"]},
					"Action": ["s3:*"],
					"Resource": ["arn:aws:s3:::%s/*"]
				}
			]
		}`, config.BucketName)

		err = minioClient.SetBucketPolicy(ctx, config.BucketName, policy)
		if err != nil {

			fmt.Printf("Warning: failed to set bucket policy: %v\n", err)
		}

		fmt.Printf("Bucket %s created successfully with secure configuration\n", config.BucketName)
	}

	if config.MaxFileSize <= 0 {
		config.MaxFileSize = MaxFileSize
	}

	return &MinioClient{
		client:         minioClient,
		bucketName:     config.BucketName,
		endpoint:       config.Endpoint,
		publicEndpoint: config.PublicEndpoint,
		maxFileSize:    config.MaxFileSize,
		useSSL:         config.UseSSL,
	}, nil
}

func (m *MinioClient) ValidateFile(upload domain.PhotoUpload) error {

	if int64(len(upload.FileData)) > m.maxFileSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", m.maxFileSize)
	}

	if len(upload.FileData) == 0 {
		return fmt.Errorf("file is empty")
	}

	if !AllowedContentTypes[upload.ContentType] {
		return fmt.Errorf("content type %s is not allowed", upload.ContentType)
	}

	ext := strings.ToLower(filepath.Ext(upload.FileName))
	if !AllowedExtensions[ext] {
		return fmt.Errorf("file extension %s is not allowed", ext)
	}

	if strings.Contains(upload.FileName, "..") || strings.Contains(upload.FileName, "/") || strings.Contains(upload.FileName, "\\") {
		return fmt.Errorf("invalid filename: path traversal detected")
	}

	return nil
}

func (m *MinioClient) GetPhoto(filePath string) (string, error) {

	if strings.Contains(filePath, "..") {
		return "", fmt.Errorf("invalid file path: path traversal detected")
	}

	presignedURL, err := m.client.PresignedGetObject(context.Background(), m.bucketName, filePath, PresignedURLExpiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	if m.publicEndpoint != "" {
		urlStr := presignedURL.String()
		protocol := "http"
		if m.useSSL {
			protocol = "https"
		}
		urlStr = fmt.Sprintf("%s://%s%s?%s", protocol, m.publicEndpoint, presignedURL.Path, presignedURL.RawQuery)
		return urlStr, nil
	}
	return presignedURL.String(), nil
}

func (m *MinioClient) Upload(upload domain.PhotoUpload) (string, string, error) {

	if err := m.ValidateFile(upload); err != nil {
		return "", "", fmt.Errorf("file validation failed: %w", err)
	}

	ctx := context.Background()
	ext := strings.ToLower(filepath.Ext(upload.FileName))

	uniqueFileName := fmt.Sprintf("%s/%s%s", upload.UserID, uuid.New().String(), ext)

	reader := bytes.NewReader(upload.FileData)

	_, err := m.client.PutObject(ctx, m.bucketName, uniqueFileName, reader, int64(len(upload.FileData)), minio.PutObjectOptions{
		ContentType: upload.ContentType,
		UserMetadata: map[string]string{
			"uploaded-by": upload.UserID,
			"upload-time": time.Now().UTC().Format(time.RFC3339),
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to upload to minio: %w", err)
	}

	presignedURL, err := m.client.PresignedGetObject(ctx, m.bucketName, uniqueFileName, PresignedURLExpiry, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	urlStr := presignedURL.String()
	if m.publicEndpoint != "" {
		protocol := "http"
		if m.useSSL {
			protocol = "https"
		}
		urlStr = fmt.Sprintf("%s://%s%s?%s", protocol, m.publicEndpoint, presignedURL.Path, presignedURL.RawQuery)
	}

	return urlStr, uniqueFileName, nil
}

func (m *MinioClient) Delete(filePath string) error {

	if strings.Contains(filePath, "..") {
		return fmt.Errorf("invalid file path: path traversal detected")
	}

	ctx := context.Background()
	err := m.client.RemoveObject(ctx, m.bucketName, filePath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete from minio: %w", err)
	}

	return nil
}

func (m *MinioClient) HealthCheck(ctx context.Context) error {
	_, err := m.client.BucketExists(ctx, m.bucketName)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	return nil
}
