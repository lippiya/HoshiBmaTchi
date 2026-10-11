package http

import (
	"io"
	"net/http"

	"hoshiBmaTchi/back-end/photo_service/internal/core/domain"
	"hoshiBmaTchi/back-end/photo_service/internal/core/service"

	"github.com/gin-gonic/gin"
)

type PhotoHTTPHandler struct {
	photoService *service.PhotoService
}

func NewPhotoHTTPHandler(photoService *service.PhotoService) *PhotoHTTPHandler {
	return &PhotoHTTPHandler{
		photoService: photoService,
	}
}

func (h *PhotoHTTPHandler) UploadPhoto(c *gin.Context) {
	file, header, err := c.Request.FormFile("media")
	if err != nil {
		file, header, err = c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to get file from request"})
			return
		}
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(fileBytes)
	}

	filename := header.Filename

	upload := domain.PhotoUpload{
		FileData:    fileBytes,
		FileName:    filename,
		ContentType: contentType,
	}

	photoResult, err := h.photoService.UploadPhoto(upload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"photo_url": photoResult.URL})
}

func (h *PhotoHTTPHandler) DeletePhoto(c *gin.Context) {
	photoURL := c.Query("photo_url")
	if photoURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "photo_url is required"})
		return
	}

	err := h.photoService.DeletePhoto(photoURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "photo deleted successfully"})
}

func SetupPhotoHTTPRoutes(router *gin.Engine, photoService *service.PhotoService) {
	handler := NewPhotoHTTPHandler(photoService)

	photos := router.Group("/api/v1/photos")
	{
		photos.POST("/upload", handler.UploadPhoto)
		photos.DELETE("/delete", handler.DeletePhoto)
	}
}
