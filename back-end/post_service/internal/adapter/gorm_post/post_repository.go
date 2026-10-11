package gorm_post

import (
	"regexp"
	"sort"
	"strings"

	"gorm.io/gorm"

	"hoshiBmaTchi/back-end/post_service/internal/core/port"

	"context"
	"hoshiBmaTchi/back-end/post_service/internal/core/domain"
)

type postRepository struct {
	db *gorm.DB
}

func NewGormPostRepo(db *gorm.DB) port.PostRepositoryPort {
	return &postRepository{
		db: db,
	}
}

func (pr *postRepository) CreatePost(c context.Context, post *domain.Post) error {
	result := pr.db.WithContext(c).Create(post)
	return result.Error
}

func (pr *postRepository) GetPostByID(c context.Context, postID string) (*domain.Post, error) {
	var post domain.Post
	result := pr.db.WithContext(c).First(&post, "id = ?", postID)
	return &post, result.Error
}

func (pr *postRepository) UpdatePost(c context.Context, post *domain.Post) error {
	result := pr.db.WithContext(c).Save(post)
	return result.Error
}

func (pr *postRepository) CreatePostMedia(c context.Context, media *domain.PostMedia) error {
	result := pr.db.WithContext(c).Create(media)
	return result.Error
}

func (pr *postRepository) DeletePostMediaByPostID(c context.Context, postID string) error {
	result := pr.db.WithContext(c).Where("post_id = ?", postID).Delete(&domain.PostMedia{})
	return result.Error
}

func (pr *postRepository) DeletePost(c context.Context, postID string) error {
	result := pr.db.WithContext(c).Delete(&domain.Post{}, "id = ?", postID)

	if result.Error != nil {
		return result.Error
	}
	result2 := pr.DeletePostMediaByPostID(c, postID)
	if result2 != nil {
		return result2
	}
	result3 := pr.DeletePostCollaboratorsByPostID(c, postID)
	if result3 != nil {
		return result.Error
	}
	return nil
}

func (pr *postRepository) CreatePostCollaborator(c context.Context, collaborator *domain.PostCollaborator) error {
	result := pr.db.WithContext(c).Create(collaborator)
	return result.Error
}

func (pr *postRepository) DeletePostCollaboratorsByPostID(c context.Context, postID string) error {
	result := pr.db.WithContext(c).Where("post_id = ?", postID).Delete(&domain.PostCollaborator{})
	return result.Error
}

func (pr *postRepository) GetAllPosts(c context.Context) ([]domain.Post, error) {
	var posts []domain.Post
	result := pr.db.WithContext(c).Where("is_reel = ?", false).Find(&posts)
	return posts, result.Error
}

func (pr *postRepository) GetAllPostsByUserID(c context.Context, userID string) ([]domain.Post, error) {
	var posts []domain.Post
	result := pr.db.WithContext(c).Where("user_id = ? AND is_reel = ?", userID, false).Find(&posts)
	return posts, result.Error
}

func (pr *postRepository) GetAllMediaByPostID(c context.Context, postID string) ([]domain.PostMedia, error) {
	var media []domain.PostMedia
	result := pr.db.WithContext(c).Where("post_id = ?", postID).Find(&media)
	return media, result.Error
}

func (pr *postRepository) GetAllCollaboratorsByPostID(c context.Context, postID string) ([]domain.PostCollaborator, error) {
	var collaborators []domain.PostCollaborator
	result := pr.db.WithContext(c).Where("post_id = ?", postID).Find(&collaborators)
	return collaborators, result.Error
}

func (pr *postRepository) GetAllReels(c context.Context) ([]domain.Post, error) {
	var posts []domain.Post
	result := pr.db.WithContext(c).Where("is_reel = ?", true).Find(&posts)
	return posts, result.Error
}

func (pr *postRepository) GetAllReelsByUserID(c context.Context, userID string) ([]domain.Post, error) {
	var posts []domain.Post
	result := pr.db.WithContext(c).Where("user_id = ? AND is_reel = ?", userID, true).Find(&posts)
	return posts, result.Error
}

func (pr *postRepository) CreateCollection(c context.Context, collection *domain.Collection) error {
	result := pr.db.WithContext(c).Create(collection)
	return result.Error
}

func (pr *postRepository) GetCollectionsByUserID(c context.Context, userID string) ([]domain.Collection, error) {
	var collections []domain.Collection
	result := pr.db.WithContext(c).Where("user_id = ?", userID).Order("updated_at DESC").Find(&collections)
	return collections, result.Error
}

func (pr *postRepository) GetCollectionByID(c context.Context, collectionID string) (*domain.Collection, error) {
	var collection domain.Collection
	result := pr.db.WithContext(c).First(&collection, "id = ?", collectionID)
	return &collection, result.Error
}

func (pr *postRepository) UpdateCollection(c context.Context, collection *domain.Collection) error {
	result := pr.db.WithContext(c).Save(collection)
	return result.Error
}

func (pr *postRepository) DeleteCollection(c context.Context, collectionID string) error {
	tx := pr.db.WithContext(c).Begin()

	if err := tx.Where("collection_id = ?", collectionID).Delete(&domain.SavedPost{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&domain.Collection{}, "id = ?", collectionID).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (pr *postRepository) AddPostToCollection(c context.Context, savedPost *domain.SavedPost) error {
	result := pr.db.WithContext(c).Create(savedPost)
	return result.Error
}

func (pr *postRepository) RemovePostFromCollection(c context.Context, collectionID string, postID string) error {
	result := pr.db.WithContext(c).Where("collection_id = ? AND post_id = ?", collectionID, postID).Delete(&domain.SavedPost{})
	return result.Error
}

func (pr *postRepository) GetPostsByCollectionID(c context.Context, collectionID string) ([]domain.Post, error) {
	var posts []domain.Post
	result := pr.db.WithContext(c).
		Joins("JOIN saved_post_models ON saved_post_models.post_id = post_models.id").
		Where("saved_post_models.collection_id = ?", collectionID).
		Order("saved_post_models.created_at DESC").
		Find(&posts)
	return posts, result.Error
}

func (pr *postRepository) GetRecentPostsByCollectionID(c context.Context, collectionID string, limit int) ([]domain.Post, error) {
	var posts []domain.Post
	result := pr.db.WithContext(c).
		Joins("JOIN saved_post_models ON saved_post_models.post_id = post_models.id").
		Where("saved_post_models.collection_id = ?", collectionID).
		Order("saved_post_models.created_at DESC").
		Limit(limit).
		Find(&posts)
	return posts, result.Error
}

func (pr *postRepository) SearchHashtags(c context.Context, query string) ([]port.HashtagCount, error) {
	var posts []domain.Post
	result := pr.db.WithContext(c).Where("caption ILIKE ?", "%#%").Find(&posts)
	if result.Error != nil {
		return nil, result.Error
	}

	hashtagRegex := regexp.MustCompile(`#[a-zA-Z0-9_\-]+`)
	hashtagCounts := make(map[string]int)

	for _, post := range posts {
		matches := hashtagRegex.FindAllString(post.Caption, -1)
		for _, match := range matches {
			tag := strings.ToLower(match)
			hashtagCounts[tag]++
		}
	}

	query = strings.ToLower(query)
	var results []port.HashtagCount
	for tag, count := range hashtagCounts {
		if query == "" || strings.Contains(tag, query) {
			results = append(results, port.HashtagCount{
				Tag:   tag,
				Count: count,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Count > results[j].Count
	})

	if len(results) > 20 {
		results = results[:20]
	}

	return results, nil
}

func (pr *postRepository) GetPostsByHashtag(c context.Context, hashtag string) ([]domain.Post, error) {
	var posts []domain.Post
	searchPattern := "%" + hashtag + "%"
	result := pr.db.WithContext(c).Where("caption ILIKE ?", searchPattern).Find(&posts)
	return posts, result.Error
}
