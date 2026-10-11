package port

import (
	"context"
	"hoshiBmaTchi/back-end/post_service/internal/core/domain"
)

type HashtagCount struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

type PostRepositoryPort interface {
	CreatePost(c context.Context, post *domain.Post) error
	GetPostByID(c context.Context, postID string) (*domain.Post, error)
	UpdatePost(c context.Context, post *domain.Post) error
	DeletePost(c context.Context, postID string) error
	CreatePostMedia(c context.Context, media *domain.PostMedia) error
	DeletePostMediaByPostID(c context.Context, postID string) error
	CreatePostCollaborator(c context.Context, collaborator *domain.PostCollaborator) error
	GetAllPosts(c context.Context) ([]domain.Post, error)
	GetAllPostsByUserID(c context.Context, userID string) ([]domain.Post, error)
	GetAllMediaByPostID(c context.Context, postID string) ([]domain.PostMedia, error)
	GetAllCollaboratorsByPostID(c context.Context, postID string) ([]domain.PostCollaborator, error)
	GetAllReels(c context.Context) ([]domain.Post, error)
	GetAllReelsByUserID(c context.Context, userID string) ([]domain.Post, error)
	SearchHashtags(c context.Context, query string) ([]HashtagCount, error)
	GetPostsByHashtag(c context.Context, hashtag string) ([]domain.Post, error)

	CreateCollection(c context.Context, collection *domain.Collection) error
	GetCollectionsByUserID(c context.Context, userID string) ([]domain.Collection, error)
	GetCollectionByID(c context.Context, collectionID string) (*domain.Collection, error)
	UpdateCollection(c context.Context, collection *domain.Collection) error
	DeleteCollection(c context.Context, collectionID string) error

	AddPostToCollection(c context.Context, savedPost *domain.SavedPost) error
	RemovePostFromCollection(c context.Context, collectionID string, postID string) error
	GetPostsByCollectionID(c context.Context, collectionID string) ([]domain.Post, error)
	GetRecentPostsByCollectionID(c context.Context, collectionID string, limit int) ([]domain.Post, error)
}

type PostServicePort interface {
	CreatePost(c context.Context, post *domain.Post, postMedia *[]domain.PostMedia, collaborator *[]domain.PostCollaborator) error
	GetPostByID(c context.Context, postID string) (*domain.Post, error)
	UpdatePost(c context.Context, post *domain.Post) error
	DeletePost(c context.Context, postID string) error
	CreatePostMedia(c context.Context, media *domain.PostMedia) error
	CreatePostCollaborator(c context.Context, collaborator *domain.PostCollaborator) error
	GetAllPosts(c context.Context) ([]domain.Post, error)
	GetAllPostsByUserID(c context.Context, userID string) ([]domain.Post, error)
	GetAllMediaByPostID(c context.Context, postID string) ([]domain.PostMedia, error)
	GetAllCollaboratorsByPostID(c context.Context, postID string) ([]domain.PostCollaborator, error)
	GetAllReels(c context.Context) ([]domain.Post, error)
	GetAllReelsByUserID(c context.Context, userID string) ([]domain.Post, error)
	SearchHashtags(c context.Context, query string) ([]HashtagCount, error)
	GetPostsByHashtag(c context.Context, hashtag string) ([]domain.Post, error)

	CreateCollection(c context.Context, collection *domain.Collection) error
	GetCollectionsByUserID(c context.Context, userID string) ([]domain.Collection, error)
	GetCollectionByID(c context.Context, collectionID string) (*domain.Collection, error)
	UpdateCollection(c context.Context, collection *domain.Collection) error
	DeleteCollection(c context.Context, collectionID string) error

	AddPostToCollection(c context.Context, savedPost *domain.SavedPost) error
	RemovePostFromCollection(c context.Context, collectionID string, postID string) error
	GetPostsByCollectionID(c context.Context, collectionID string) ([]domain.Post, error)
	GetRecentPostsByCollectionID(c context.Context, collectionID string, limit int) ([]domain.Post, error)
}

type PhotoClientPort interface {
	UploadPhoto(file interface{}, userID string) (string, error)
	DeletePhoto(filePath string) error
	GetPhotoURL(filePath string) (string, error)
}
