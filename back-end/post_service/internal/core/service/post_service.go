package service

import (
	"context"
	"hoshiBmaTchi/back-end/post_service/internal/core/domain"
	"hoshiBmaTchi/back-end/post_service/internal/core/port"
	"time"

	"github.com/google/uuid"
)

type PostServicePort struct {
	PostRepository port.PostRepositoryPort
}

func NewPostService(postRepo port.PostRepositoryPort) port.PostServicePort {
	return &PostServicePort{
		PostRepository: postRepo,
	}
}

func (ps *PostServicePort) GetAllPostsByUserID(c context.Context, userID string) ([]domain.Post, error) {
	return ps.PostRepository.GetAllPostsByUserID(c, userID)
}

func (ps *PostServicePort) CreatePost(c context.Context, post *domain.Post, postMedia *[]domain.PostMedia, collaborators *[]domain.PostCollaborator) error {
	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()
	post.ID = uuid.New().String()
	if postMedia != nil {
		for i := range *postMedia {
			(*postMedia)[i].ID = uuid.New().String()
			(*postMedia)[i].OrderNum = i + 1
			(*postMedia)[i].PostID = post.ID
			(*postMedia)[i].CreatedAt = time.Now()
			(*postMedia)[i].UpdatedAt = time.Now()
		}
		post.Media = *postMedia
	}
	if collaborators != nil {
		for i := range *collaborators {
			(*collaborators)[i].ID = uuid.New().String()
			(*collaborators)[i].PostID = post.ID
			(*collaborators)[i].CreatedAt = time.Now()
			(*collaborators)[i].UpdatedAt = time.Now()
		}
		post.Collaborators = *collaborators
	}

	err := ps.PostRepository.CreatePost(c, post)
	return err
}

func (ps *PostServicePort) GetPostByID(c context.Context, postID string) (*domain.Post, error) {
	return ps.PostRepository.GetPostByID(c, postID)
}

func (ps *PostServicePort) UpdatePost(c context.Context, post *domain.Post) error {
	post.UpdatedAt = time.Now()
	err := ps.PostRepository.UpdatePost(c, post)
	return err
}

func (ps *PostServicePort) DeletePost(c context.Context, postID string) error {
	err := ps.PostRepository.DeletePost(c, postID)
	return err
}

func (ps *PostServicePort) CreatePostMedia(c context.Context, media *domain.PostMedia) error {
	return ps.PostRepository.CreatePostMedia(c, media)
}

func (ps *PostServicePort) CreatePostCollaborator(c context.Context, collaborator *domain.PostCollaborator) error {
	return ps.PostRepository.CreatePostCollaborator(c, collaborator)
}

func (ps *PostServicePort) GetAllPosts(c context.Context) ([]domain.Post, error) {
	return ps.PostRepository.GetAllPosts(c)
}

func (ps *PostServicePort) GetAllMediaByPostID(c context.Context, postID string) ([]domain.PostMedia, error) {
	return ps.PostRepository.GetAllMediaByPostID(c, postID)
}

func (ps *PostServicePort) GetAllCollaboratorsByPostID(c context.Context, postID string) ([]domain.PostCollaborator, error) {
	return ps.PostRepository.GetAllCollaboratorsByPostID(c, postID)
}

func (ps *PostServicePort) GetAllReels(c context.Context) ([]domain.Post, error) {
	return ps.PostRepository.GetAllReels(c)
}

func (ps *PostServicePort) GetAllReelsByUserID(c context.Context, userID string) ([]domain.Post, error) {
	return ps.PostRepository.GetAllReelsByUserID(c, userID)
}

func (ps *PostServicePort) CreateCollection(c context.Context, collection *domain.Collection) error {
	collection.ID = uuid.New().String()
	collection.CreatedAt = time.Now()
	collection.UpdatedAt = time.Now()
	return ps.PostRepository.CreateCollection(c, collection)
}

func (ps *PostServicePort) GetCollectionsByUserID(c context.Context, userID string) ([]domain.Collection, error) {
	return ps.PostRepository.GetCollectionsByUserID(c, userID)
}

func (ps *PostServicePort) GetCollectionByID(c context.Context, collectionID string) (*domain.Collection, error) {
	return ps.PostRepository.GetCollectionByID(c, collectionID)
}

func (ps *PostServicePort) UpdateCollection(c context.Context, collection *domain.Collection) error {
	collection.UpdatedAt = time.Now()
	return ps.PostRepository.UpdateCollection(c, collection)
}

func (ps *PostServicePort) DeleteCollection(c context.Context, collectionID string) error {
	return ps.PostRepository.DeleteCollection(c, collectionID)
}

func (ps *PostServicePort) AddPostToCollection(c context.Context, savedPost *domain.SavedPost) error {
	savedPost.ID = uuid.New().String()
	savedPost.CreatedAt = time.Now()

	collection, err := ps.PostRepository.GetCollectionByID(c, savedPost.CollectionID)
	if err != nil {
		return err
	}

	err = ps.PostRepository.AddPostToCollection(c, savedPost)
	if err != nil {
		return err
	}

	collection.UpdatedAt = time.Now()
	return ps.PostRepository.UpdateCollection(c, collection)
}

func (ps *PostServicePort) RemovePostFromCollection(c context.Context, collectionID string, postID string) error {
	return ps.PostRepository.RemovePostFromCollection(c, collectionID, postID)
}

func (ps *PostServicePort) GetPostsByCollectionID(c context.Context, collectionID string) ([]domain.Post, error) {
	return ps.PostRepository.GetPostsByCollectionID(c, collectionID)
}

func (ps *PostServicePort) GetRecentPostsByCollectionID(c context.Context, collectionID string, limit int) ([]domain.Post, error) {
	return ps.PostRepository.GetRecentPostsByCollectionID(c, collectionID, limit)
}

func (ps *PostServicePort) SearchHashtags(c context.Context, query string) ([]port.HashtagCount, error) {
	return ps.PostRepository.SearchHashtags(c, query)
}

func (ps *PostServicePort) GetPostsByHashtag(c context.Context, hashtag string) ([]domain.Post, error) {
	return ps.PostRepository.GetPostsByHashtag(c, hashtag)
}
