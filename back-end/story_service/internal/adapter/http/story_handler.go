package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"story_service/internal/adapter/grpcclient"
	"story_service/internal/core/domain"
	"story_service/internal/core/port"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type StoryHandler struct {
	service     port.StoryService
	photoClient *grpcclient.PhotoClient
}

func NewStoryHandler(service port.StoryService, photoClient *grpcclient.PhotoClient) *StoryHandler {
	return &StoryHandler{
		service:     service,
		photoClient: photoClient,
	}
}

func (h *StoryHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/create", h.CreateStory).Methods("POST")
	router.HandleFunc("/get/{id}", h.GetStoryByID).Methods("GET")
	router.HandleFunc("/user/{userId}", h.GetStoriesByUserID).Methods("GET")
	router.HandleFunc("/get-all-stories", h.GetAllActiveStories).Methods("GET")
	router.HandleFunc("/archived", h.GetArchivedStories).Methods("GET")
	router.HandleFunc("/delete/{id}", h.DeleteStory).Methods("DELETE")

	router.HandleFunc("/like-story/{id}", h.LikeStory).Methods("POST")
	router.HandleFunc("/check-like/{id}", h.CheckStoryLike).Methods("GET")

	router.HandleFunc("/reply/{id}", h.ReplyToStory).Methods("POST")
	router.HandleFunc("/replies/{id}", h.GetStoryReplies).Methods("GET")

	router.HandleFunc("/view/{id}", h.ViewStory).Methods("POST")
	router.HandleFunc("/check-view/{id}", h.CheckStoryView).Methods("GET")
	router.HandleFunc("/views/{id}", h.GetStoryViews).Methods("GET")

	router.HandleFunc("/share/{id}", h.ShareStory).Methods("POST")
}

func getUserIDFromToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	tokenString := parts[1]
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
		return ""
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if userID, ok := claims["user_id"].(string); ok {
			return userID
		}
	}

	return ""
}
func (h *StoryHandler) CreateStory(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromToken(r)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("media")
	if err != nil {
		http.Error(w, "media file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filePath, err := h.photoClient.UploadPhoto(header, userID)
	if err != nil {

		fmt.Printf("Upload Failed: %v\n", err)
		http.Error(w, "failed to upload media to storage", http.StatusInternalServerError)
		return
	}

	durationStr := r.FormValue("duration")
	duration := 5
	if d, err := strconv.Atoi(durationStr); err == nil {
		duration = d
	}

	mediaType := "image"
	if strings.HasPrefix(header.Header.Get("Content-Type"), "video/") {
		mediaType = "video"
	}

	story, err := h.service.CreateStory(userID, filePath, mediaType, duration)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"story":   story,
		"message": "Story created successfully",
	})
}
func (h *StoryHandler) GetStoryByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	story, err := h.service.GetStoryByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"story": story,
	})
}

func (h *StoryHandler) GetStoriesByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	stories, err := h.service.GetStoriesByUserID(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stories": stories,
	})
}

func (h *StoryHandler) GetAllActiveStories(w http.ResponseWriter, r *http.Request) {
	stories, err := h.service.GetAllActiveStories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	storiesByUser := make(map[string][]*domain.Story)
	userIDs := make([]string, 0)
	for _, story := range stories {
		if _, exists := storiesByUser[story.UserID]; !exists {
			userIDs = append(userIDs, story.UserID)
		}
		storiesByUser[story.UserID] = append(storiesByUser[story.UserID], story)
	}

	userInfoMap := make(map[string]map[string]interface{})
	for _, userID := range userIDs {
		userInfo := fetchUserInfo(userID)
		if userInfo != nil {
			userInfoMap[userID] = userInfo
		}
	}

	var storyGroups []map[string]interface{}
	for userID, userStories := range storiesByUser {
		group := map[string]interface{}{
			"user_id": userID,
			"stories": userStories,
		}
		if userInfo, ok := userInfoMap[userID]; ok {
			group["user"] = userInfo
		}
		storyGroups = append(storyGroups, group)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"story_groups": storyGroups,
	})
}

func (h *StoryHandler) GetArchivedStories(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromToken(r)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	stories, err := h.service.GetArchivedStories(userID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stories": stories,
	})
}

func fetchUserInfo(userID string) map[string]interface{} {
	userServiceURL := os.Getenv("USER_SERVICE_URL")
	if userServiceURL == "" {
		userServiceURL = "http://user-service:8080"
	}

	resp, err := http.Get(fmt.Sprintf("%s/api/v1/users/get/%s", userServiceURL, userID))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil
	}

	if user, ok := result["user"].(map[string]interface{}); ok {
		return user
	}

	return nil
}

func (h *StoryHandler) DeleteStory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	userID := getUserIDFromToken(r)

	if err := h.service.DeleteStory(id, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Story deleted successfully",
	})
}

func (h *StoryHandler) LikeStory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	storyID := vars["id"]
	userID := getUserIDFromToken(r)

	if err := h.service.LikeStory(storyID, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Success",
	})
}

func (h *StoryHandler) CheckStoryLike(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	storyID := vars["id"]
	userID := getUserIDFromToken(r)

	liked, err := h.service.IsStoryLiked(storyID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"liked": liked,
	})
}

func (h *StoryHandler) ReplyToStory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	storyID := vars["id"]
	userID := getUserIDFromToken(r)

	var req struct {
		Reply string `json:"reply"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.ReplyToStory(storyID, userID, req.Reply); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Reply sent successfully",
	})
}

func (h *StoryHandler) GetStoryReplies(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	storyID := vars["id"]

	replies, err := h.service.GetStoryReplies(storyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"replies": replies,
	})
}

func (h *StoryHandler) ViewStory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	storyID := vars["id"]
	userID := getUserIDFromToken(r)

	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.service.ViewStory(storyID, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Story viewed",
	})
}

func (h *StoryHandler) CheckStoryView(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	storyID := vars["id"]
	userID := getUserIDFromToken(r)

	viewed, err := h.service.HasViewedStory(storyID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"viewed": viewed,
	})
}

func (h *StoryHandler) GetStoryViews(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	storyID := vars["id"]

	views, err := h.service.GetStoryViews(storyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"views": views,
	})
}

func (h *StoryHandler) ShareStory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	storyID := vars["id"]
	senderID := getUserIDFromToken(r)

	var req struct {
		RecipientID string `json:"recipient_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.ShareStory(storyID, senderID, req.RecipientID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Story shared successfully",
	})
}
