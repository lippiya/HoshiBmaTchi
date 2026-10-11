package http

import (
	"encoding/json"
	"fmt"
	"hoshiBmaTchi/back-end/post_service/internal/core/domain"
	"log"
	"net/http"
	"os"

	"strings"

	"hoshiBmaTchi/back-end/post_service/internal/core/port"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type PostHandler struct {
	PostService port.PostServicePort
	photoClient port.PhotoClientPort
}

func NewPostHandler(postService port.PostServicePort, photoClient port.PhotoClientPort) *PostHandler {
	return &PostHandler{
		PostService: postService,
		photoClient: photoClient,
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

// @Summary Create a new post with media
// @Description Create a new post with multiple media files and optional collaborators
// @Tags Posts
// @Accept multipart/form-data
// @Produce json
// @Param user_id formData string true "User ID"
// @Param caption formData string false "Post caption"
// @Param is_comment_off formData string false "Disable comments"
// @Param is_hide_likes formData string false "Hide likes count"
// @Param media formData file true "Media files"
// @Param collaborators formData []string false "Collaborator user IDs"
// @Security BearerAuth
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/create-post [post]
func (ph *PostHandler) CreatePostHandler(c *gin.Context) {
	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err := c.Request.ParseMultipartForm(100 << 20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse form"})
		return
	}

	userID := c.PostForm("user_id")
	caption := c.PostForm("caption")
	isCommentOff := c.PostForm("is_comment_off") == "true"
	isHideLikes := c.PostForm("is_hide_likes") == "true"

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	form := c.Request.MultipartForm
	files := form.File["media"]

	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one media file is required"})
		return
	}

	if len(files) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "maximum 10 files allowed"})
		return
	}

	var mediaList []domain.PostMedia
	for i, fileHeader := range files {
		if fileHeader.Size > 10*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("file %s exceeds 10MB", fileHeader.Filename),
			})
			return
		}

		contentType := fileHeader.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") && !strings.HasPrefix(contentType, "video/") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("file %s is not an image or video", fileHeader.Filename),
			})
			return
		}

		filePath, err := ph.photoClient.UploadPhoto(fileHeader, email.(string))
		if err != nil {
			log.Printf("Failed to upload photo: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("failed to upload file %s", fileHeader.Filename),
			})
			return
		}

		mediaType := "image"
		if strings.HasPrefix(contentType, "video/") {
			mediaType = "video"
		}

		mediaList = append(mediaList, domain.PostMedia{
			FilePath:  filePath,
			MediaType: mediaType,
			OrderNum:  i,
		})
	}

	isReel := len(files) == 1 && strings.HasPrefix(files[0].Header.Get("Content-Type"), "video/")

	post := &domain.Post{
		UserID:       userID,
		Caption:      caption,
		IsCommentOff: isCommentOff,
		IsHideLikes:  isHideLikes,
		IsReel:       isReel,
	}

	var collaboratorList []domain.PostCollaborator
	collaboratorIDs := c.PostFormArray("collaborators")
	if len(collaboratorIDs) > 0 {
		for _, colID := range collaboratorIDs {
			if colID != "" {
				collaboratorList = append(collaboratorList, domain.PostCollaborator{
					UserID: colID,
				})
			}
		}
	}

	var collabPtr *[]domain.PostCollaborator
	if len(collaboratorList) > 0 {
		collabPtr = &collaboratorList
	}

	err = ph.PostService.CreatePost(c.Request.Context(), post, &mediaList, collabPtr)
	if err != nil {
		log.Printf("Failed to create post: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "post created successfully",
		"post_id": post.ID,
	})
}

type MediaResponse struct {
	FilePath  string
	MediaType string
	OrderNum  int
	FileName  string
}

// @Summary Upload media files
// @Description Upload multiple media files without creating a post
// @Tags Posts
// @Accept multipart/form-data
// @Produce json
// @Param media formData file true "Media files"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/upload-media [post]
func (ph *PostHandler) UploadMediaHandler(c *gin.Context) {
	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err := c.Request.ParseMultipartForm(50 << 20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse form"})
		return
	}

	form := c.Request.MultipartForm
	files := form.File["media"]

	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no files uploaded"})
		return
	}

	if len(files) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "maximum 10 files allowed"})
		return
	}

	var uploadedMedia []MediaResponse

	for i, fileHeader := range files {
		if fileHeader.Size > 10*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("file %s exceeds 10MB", fileHeader.Filename),
			})
			return
		}

		contentType := fileHeader.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") && !strings.HasPrefix(contentType, "video/") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("file %s is not an image or video", fileHeader.Filename),
			})
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("failed to open file %s", fileHeader.Filename),
			})
			return
		}
		defer file.Close()

		filePath, err := ph.photoClient.UploadPhoto(file, email.(string))
		if err != nil {
			log.Printf("Failed to upload photo: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("failed to upload file %s", fileHeader.Filename),
			})
			return
		}

		mediaType := "image"
		if strings.HasPrefix(contentType, "video/") {
			mediaType = "video"
		}

		uploadedMedia = append(uploadedMedia, MediaResponse{
			FilePath:  filePath,
			MediaType: mediaType,
			OrderNum:  i,
			FileName:  fileHeader.Filename,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "media uploaded successfully",
		"media":   uploadedMedia,
	})
}

// @Summary Create post with predefined file paths
// @Description Create a post using existing file paths without uploading new media
// @Tags Posts
// @Accept json
// @Produce json
// @Param request body object true "Post creation request"
// @Security BearerAuth
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/create-with-paths [post]
func (ph *PostHandler) CreatePostWithPathsHandler(c *gin.Context) {
	_, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		UserID       string `json:"user_id" binding:"required"`
		Caption      string `json:"caption"`
		IsCommentOff bool   `json:"is_comment_off"`
		Media        []struct {
			FilePath  string `json:"file_path" binding:"required"`
			MediaType string `json:"media_type" binding:"required"`
			OrderNum  int    `json:"order_num"`
			Duration  int    `json:"duration"`
			Width     int    `json:"width"`
			Height    int    `json:"height"`
		} `json:"media" binding:"required,min=1"`
		Collaborators []string `json:"collaborators"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var mediaList []domain.PostMedia
	for _, m := range req.Media {
		mediaList = append(mediaList, domain.PostMedia{
			FilePath:  m.FilePath,
			MediaType: m.MediaType,
			OrderNum:  m.OrderNum,
			Duration:  m.Duration,
			Width:     m.Width,
			Height:    m.Height,
		})
	}

	post := &domain.Post{
		UserID:       req.UserID,
		Caption:      req.Caption,
		IsCommentOff: req.IsCommentOff,
	}

	var collaboratorList []domain.PostCollaborator
	if len(req.Collaborators) > 0 {
		for _, colID := range req.Collaborators {
			if colID != "" {
				collaboratorList = append(collaboratorList, domain.PostCollaborator{
					UserID: colID,
				})
			}
		}
	}

	var collabPtr *[]domain.PostCollaborator
	if len(collaboratorList) > 0 {
		collabPtr = &collaboratorList
	}

	err := ph.PostService.CreatePost(c.Request.Context(), post, &mediaList, collabPtr)
	if err != nil {
		log.Printf("Failed to create post: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "post created successfully",
		"post_id": post.ID,
	})
}

// @Summary Delete media files
// @Description Delete media files from storage
// @Tags Posts
// @Accept json
// @Produce json
// @Param request body object true "File paths to delete"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 206 {object} map[string]interface{}
// @Router /posts/delete-media [delete]
func (ph *PostHandler) DeleteMediaHandler(c *gin.Context) {
	_, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		FilePaths []string `json:"file_paths" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var failedDeletes []string
	for _, filePath := range req.FilePaths {
		if err := ph.photoClient.DeletePhoto(filePath); err != nil {
			log.Printf("Failed to delete photo %s: %v", filePath, err)
			failedDeletes = append(failedDeletes, filePath)
		}
	}

	if len(failedDeletes) > 0 {
		c.JSON(http.StatusPartialContent, gin.H{
			"message": "some files could not be deleted",
			"failed":  failedDeletes,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "all media deleted successfully",
	})
}

// @Summary Delete a post
// @Description Delete a post by ID
// @Tags Posts
// @Produce json
// @Param postID path string true "Post ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/delete-post/{postID} [delete]
func (ph *PostHandler) DeletePostHandler(c *gin.Context) {
	postID := c.Param("postID")
	if postID != "" {
		err := ph.PostService.DeletePost(c.Request.Context(), postID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "post deleted successfully",
	})
}

// @Summary Get all posts
// @Description Retrieve all posts
// @Tags Posts
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/get-all-posts [get]
func (ph *PostHandler) GetAllPostsHandler(c *gin.Context) {
	_, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		log.Println("Unauthorized access attempt to GetAllPostsHandler")
		return
	}

	posts, err := ph.PostService.GetAllPosts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
	})
}

// @Summary Get all posts by user
// @Description Retrieve all posts created by a specific user
// @Tags Posts
// @Produce json
// @Param userID path string true "User ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/get-posts-by-user/{userID} [get]
func (ph *PostHandler) GetAllPostsByUserIDHandler(c *gin.Context) {
	_, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		log.Println("Unauthorized access attempt to GetAllPostsHandler")
		return
	}

	userID := c.Param("userID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	posts, err := ph.PostService.GetAllPostsByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
	})
}

// @Summary Check if post is liked by user
// @Description Check if current user has liked a post
// @Tags Posts
// @Produce json
// @Param postID path string true "Post ID"
// @Security BearerAuth
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /posts/check-like/{postID} [get]
func (ph *PostHandler) CheckLikeHandler(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "post ID is required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	post, err := ph.PostService.GetPostByID(c.Request.Context(), postID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"liked": false})
		return
	}

	userIDStr, _ := userID.(string)
	hasLiked := false
	var likedUsers []string

	if post.LikedByUsers != "" {
		err := json.Unmarshal([]byte(post.LikedByUsers), &likedUsers)
		if err == nil {
			for _, id := range likedUsers {
				if id == userIDStr {
					hasLiked = true
					break
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"liked": hasLiked})
}

// @Summary Check if post is saved by user
// @Description Check if current user has saved a post
// @Tags Posts
// @Produce json
// @Param postID path string true "Post ID"
// @Security BearerAuth
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /posts/check-save/{postID} [get]
func (ph *PostHandler) CheckSaveHandler(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "post ID is required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userIDStr, _ := userID.(string)

	collections, err := ph.PostService.GetCollectionsByUserID(c.Request.Context(), userIDStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"saved": false})
		return
	}

	var defaultCollection *domain.Collection
	for i := range collections {
		if collections[i].Name == "All Saved Posts" {
			defaultCollection = &collections[i]
			break
		}
	}

	if defaultCollection == nil {
		c.JSON(http.StatusOK, gin.H{"saved": false})
		return
	}

	posts, err := ph.PostService.GetPostsByCollectionID(c.Request.Context(), defaultCollection.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"saved": false})
		return
	}

	hasSaved := false
	for _, p := range posts {
		if p.ID == postID {
			hasSaved = true
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{"saved": hasSaved})
}

// @Summary Get all media by post ID
// @Description Retrieve all media files associated with a post
// @Tags Posts
// @Produce json
// @Param postID path string true "Post ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/get-post-media-by-id/{postID} [get]
func (ph *PostHandler) GetAllMediaByPostIDHandler(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "postID is required"})
		return
	}

	media, err := ph.PostService.GetAllMediaByPostID(c.Request.Context(), postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"media": media,
	})
}

// @Summary Get all collaborators by post ID
// @Description Retrieve all collaborators for a post
// @Tags Posts
// @Produce json
// @Param postID path string true "Post ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/get-post-collaborators-by-id/{postID} [get]
func (ph *PostHandler) GetAllCollaboratorsByIDHandler(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "postID is required"})
		return
	}

	collaborators, err := ph.PostService.GetAllCollaboratorsByPostID(c.Request.Context(), postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"collaborators": collaborators,
	})
}

// @Summary Get post by ID
// @Description Retrieve a specific post by ID
// @Tags Posts
// @Produce json
// @Param postID path string true "Post ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/get-post-by-id/{postID} [get]
func (ph *PostHandler) GetPostByIDHandler(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "post ID is required"})
		return
	}
	post, err := ph.PostService.GetPostByID(c.Request.Context(), postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"posts": []domain.Post{*post},
	})
}

// @Summary Like or unlike a post
// @Description Toggle like status for a post
// @Tags Posts
// @Produce json
// @Param postID path string true "Post ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/like-post/{postID} [post]
func (ph *PostHandler) LikePostHandler(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "post ID is required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	post, err := ph.PostService.GetPostByID(c.Request.Context(), postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userIDStr, _ := userID.(string)
	hasLiked := false
	var likedUsers []string

	if post.LikedByUsers != "" {
		err := json.Unmarshal([]byte(post.LikedByUsers), &likedUsers)
		if err != nil {
			likedUsers = []string{}
		}
	}

	for _, id := range likedUsers {
		if id == userIDStr {
			hasLiked = true
			break
		}
	}

	if hasLiked {
		var newLikedUsers []string
		for _, id := range likedUsers {
			if id != userIDStr {
				newLikedUsers = append(newLikedUsers, id)
			}
		}
		likedUsers = newLikedUsers
		post.LikesCount = post.LikesCount - 1
	} else {
		likedUsers = append(likedUsers, userIDStr)
		post.LikesCount = post.LikesCount + 1
	}

	likedUsersJSON, _ := json.Marshal(likedUsers)
	post.LikedByUsers = string(likedUsersJSON)

	err = ph.PostService.UpdatePost(c.Request.Context(), post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	status := "unliked"
	if !hasLiked {
		status = "liked"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "post " + status + " successfully",
		"liked":      !hasLiked,
		"likesCount": post.LikesCount,
	})
}

// @Summary Save or unsave a post
// @Description Toggle save status for a post
// @Tags Posts
// @Produce json
// @Param postID path string true "Post ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/save-post/{postID} [post]
func (ph *PostHandler) SavePostHandler(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "post ID is required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userIDStr, _ := userID.(string)

	collections, err := ph.PostService.GetCollectionsByUserID(c.Request.Context(), userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var defaultCollection *domain.Collection
	for i := range collections {
		if collections[i].Name == "All Saved Posts" {
			defaultCollection = &collections[i]
			break
		}
	}

	if defaultCollection == nil {
		newCollection := &domain.Collection{
			UserID: userIDStr,
			Name:   "All Saved Posts",
		}
		err = ph.PostService.CreateCollection(c.Request.Context(), newCollection)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create default collection"})
			return
		}
		defaultCollection = newCollection
	}

	posts, err := ph.PostService.GetPostsByCollectionID(c.Request.Context(), defaultCollection.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	hasSaved := false
	for _, p := range posts {
		if p.ID == postID {
			hasSaved = true
			break
		}
	}

	if hasSaved {
		err = ph.PostService.RemovePostFromCollection(c.Request.Context(), defaultCollection.ID, postID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "post unsaved successfully",
			"saved":   false,
		})
	} else {
		savedPost := &domain.SavedPost{
			CollectionID: defaultCollection.ID,
			PostID:       postID,
		}
		err = ph.PostService.AddPostToCollection(c.Request.Context(), savedPost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "post saved successfully",
			"saved":   true,
		})
	}
}

// @Summary Get all reels by user
// @Description Retrieve all reels created by a specific user
// @Tags Posts
// @Produce json
// @Param userID path string true "User ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/get-reels-by-user/{userID} [get]
func (ph *PostHandler) GetAllReelsByUserIDHandler(c *gin.Context) {
	_, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		log.Println("Unauthorized access attempt to GetAllReelsByUserIDHandler")
		return
	}

	userID := c.Param("userID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	posts, err := ph.PostService.GetAllReelsByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
	})
}

// @Summary Create a collection
// @Description Create a new collection for saving posts
// @Tags Collections
// @Accept json
// @Produce json
// @Param request body object true "Collection name"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /collections/create [post]
func (ph *PostHandler) CreateCollectionHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := &domain.Collection{
		UserID: userID.(string),
		Name:   req.Name,
	}

	err := ph.PostService.CreateCollection(c.Request.Context(), collection)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "collection created successfully",
		"collection": collection,
	})
}

// @Summary Get collections by user
// @Description Retrieve all collections for a user with recent posts
// @Tags Collections
// @Produce json
// @Param userID path string true "User ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /collections/user/{userID} [get]
func (ph *PostHandler) GetCollectionsByUserIDHandler(c *gin.Context) {
	_, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := c.Param("userID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	collections, err := ph.PostService.GetCollectionsByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type CollectionWithPosts struct {
		Collection  domain.Collection `json:"collection"`
		RecentPosts []domain.Post     `json:"recent_posts"`
	}

	var result []CollectionWithPosts
	for _, collection := range collections {
		posts, err := ph.PostService.GetRecentPostsByCollectionID(c.Request.Context(), collection.ID, 4)
		if err != nil {
			posts = []domain.Post{}
		}
		result = append(result, CollectionWithPosts{
			Collection:  collection,
			RecentPosts: posts,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"collections": result,
	})
}

// @Summary Update a collection
// @Description Update collection name
// @Tags Collections
// @Accept json
// @Produce json
// @Param collectionID path string true "Collection ID"
// @Param request body object true "New collection name"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /collections/update/{collectionID} [put]
func (ph *PostHandler) UpdateCollectionHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	collectionID := c.Param("collectionID")
	if collectionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "collection_id is required"})
		return
	}

	collection, err := ph.PostService.GetCollectionByID(c.Request.Context(), collectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
		return
	}

	if collection.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to update this collection"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection.Name = req.Name
	err = ph.PostService.UpdateCollection(c.Request.Context(), collection)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "collection updated successfully",
		"collection": collection,
	})
}

// @Summary Delete a collection
// @Description Delete a collection
// @Tags Collections
// @Produce json
// @Param collectionID path string true "Collection ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /collections/delete/{collectionID} [delete]
func (ph *PostHandler) DeleteCollectionHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	collectionID := c.Param("collectionID")
	if collectionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "collection_id is required"})
		return
	}

	collection, err := ph.PostService.GetCollectionByID(c.Request.Context(), collectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
		return
	}

	if collection.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to delete this collection"})
		return
	}

	err = ph.PostService.DeleteCollection(c.Request.Context(), collectionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "collection deleted successfully",
	})
}

// @Summary Add post to collection
// @Description Add a post to a collection
// @Tags Collections
// @Accept json
// @Produce json
// @Param request body object true "Collection and post IDs"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /collections/add-post [post]
func (ph *PostHandler) AddPostToCollectionHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		CollectionID string `json:"collection_id" binding:"required"`
		PostID       string `json:"post_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection, err := ph.PostService.GetCollectionByID(c.Request.Context(), req.CollectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
		return
	}

	if collection.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to add to this collection"})
		return
	}

	savedPost := &domain.SavedPost{
		CollectionID: req.CollectionID,
		PostID:       req.PostID,
	}

	err = ph.PostService.AddPostToCollection(c.Request.Context(), savedPost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "post added to collection successfully",
	})
}

// @Summary Remove post from collection
// @Description Remove a post from a collection
// @Tags Collections
// @Produce json
// @Param collectionID path string true "Collection ID"
// @Param postID path string true "Post ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /collections/remove-post/{collectionID}/{postID} [delete]
func (ph *PostHandler) RemovePostFromCollectionHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	collectionID := c.Param("collectionID")
	postID := c.Param("postID")

	if collectionID == "" || postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "collection_id and post_id are required"})
		return
	}

	collection, err := ph.PostService.GetCollectionByID(c.Request.Context(), collectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
		return
	}

	if collection.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to modify this collection"})
		return
	}

	err = ph.PostService.RemovePostFromCollection(c.Request.Context(), collectionID, postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "post removed from collection successfully",
	})
}

// @Summary Get posts by collection ID
// @Description Retrieve all posts in a collection
// @Tags Collections
// @Produce json
// @Param collectionID path string true "Collection ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /collections/posts/{collectionID} [get]
func (ph *PostHandler) GetPostsByCollectionIDHandler(c *gin.Context) {
	_, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	collectionID := c.Param("collectionID")
	if collectionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "collection_id is required"})
		return
	}

	posts, err := ph.PostService.GetPostsByCollectionID(c.Request.Context(), collectionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
	})
}

// @Summary Search hashtags
// @Description Search for hashtags
// @Tags Posts
// @Produce json
// @Param q query string true "Search query"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/search-hashtags [get]
func (ph *PostHandler) SearchHashtagsHandler(c *gin.Context) {
	_, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	query := c.Query("q")
	if query != "" && !strings.HasPrefix(query, "#") {
		query = "#" + query
	}

	hashtags, err := ph.PostService.SearchHashtags(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hashtags": hashtags,
	})
}

// @Summary Get posts by hashtag
// @Description Retrieve all posts with a specific hashtag
// @Tags Posts
// @Produce json
// @Param hashtag query string true "Hashtag to search for"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/get-posts-by-hashtag [get]
func (ph *PostHandler) GetPostsByHashtagHandler(c *gin.Context) {
	_, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	hashtag := c.Query("hashtag")
	if hashtag == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "hashtag query parameter is required"})
		return
	}

	if !strings.HasPrefix(hashtag, "#") {
		hashtag = "#" + hashtag
	}

	posts, err := ph.PostService.GetPostsByHashtag(c.Request.Context(), hashtag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
	})
}
