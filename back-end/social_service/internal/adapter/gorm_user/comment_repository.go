package gorm_user

import (
	"hoshiBmaTchi/back-end/social_service/internal/core/domain"
	"time"

	"gorm.io/gorm"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) CreateComment(req domain.CreateCommentRequest) (*domain.Comment, error) {
	comment := CommentModel{
		PostID:          req.PostID,
		ParentCommentID: req.ParentCommentID,
		UserID:          req.UserID,
		Text:            req.Text,
		GifUrl:          req.GifUrl,
		LikesCount:      0,
		RepliesCount:    0,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := r.db.Create(&comment).Error; err != nil {
		return nil, err
	}

	if req.ParentCommentID != nil {
		r.db.Model(&CommentModel{}).
			Where("id = ?", *req.ParentCommentID).
			UpdateColumn("replies_count", gorm.Expr("replies_count + ?", 1))
	}

	return &domain.Comment{
		ID:              comment.ID,
		PostID:          comment.PostID,
		ParentCommentID: comment.ParentCommentID,
		UserID:          comment.UserID,
		Text:            comment.Text,
		GifUrl:          comment.GifUrl,
		LikesCount:      comment.LikesCount,
		RepliesCount:    comment.RepliesCount,
		CreatedAt:       comment.CreatedAt,
		UpdatedAt:       comment.UpdatedAt,
	}, nil
}

func (r *CommentRepository) GetCommentsByPostID(postID string, page, limit int, sort string) ([]domain.Comment, error) {
	var comments []CommentModel
	query := r.db.Where("post_id = ? AND parent_comment_id IS NULL AND deleted_at IS NULL", postID)

	if sort == "likes" {
		query = query.Order("likes_count DESC, created_at DESC")
	} else {
		query = query.Order("created_at DESC")
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&comments).Error; err != nil {
		return nil, err
	}

	result := make([]domain.Comment, len(comments))
	for i, c := range comments {
		result[i] = domain.Comment{
			ID:              c.ID,
			PostID:          c.PostID,
			ParentCommentID: c.ParentCommentID,
			UserID:          c.UserID,
			Text:            c.Text,
			GifUrl:          c.GifUrl,
			LikesCount:      c.LikesCount,
			RepliesCount:    c.RepliesCount,
			CreatedAt:       c.CreatedAt,
			UpdatedAt:       c.UpdatedAt,
		}
	}

	return result, nil
}

func (r *CommentRepository) GetCommentReplies(commentID string) ([]domain.Comment, error) {
	var comments []CommentModel
	if err := r.db.Where("parent_comment_id = ? AND deleted_at IS NULL", commentID).
		Order("likes_count DESC, created_at DESC").
		Find(&comments).Error; err != nil {
		return nil, err
	}

	result := make([]domain.Comment, len(comments))
	for i, c := range comments {
		result[i] = domain.Comment{
			ID:              c.ID,
			PostID:          c.PostID,
			ParentCommentID: c.ParentCommentID,
			UserID:          c.UserID,
			Text:            c.Text,
			GifUrl:          c.GifUrl,
			LikesCount:      c.LikesCount,
			RepliesCount:    c.RepliesCount,
			CreatedAt:       c.CreatedAt,
			UpdatedAt:       c.UpdatedAt,
		}
	}

	return result, nil
}

func (r *CommentRepository) DeleteComment(commentID, userID string) error {
	var comment CommentModel
	if err := r.db.Where("id = ? AND user_id = ?", commentID, userID).First(&comment).Error; err != nil {
		return err
	}

	now := time.Now()
	comment.DeletedAt = &now

	if err := r.db.Save(&comment).Error; err != nil {
		return err
	}

	if comment.ParentCommentID != nil {
		r.db.Model(&CommentModel{}).
			Where("id = ?", *comment.ParentCommentID).
			UpdateColumn("replies_count", gorm.Expr("GREATEST(replies_count - 1, 0)"))
	}

	return nil
}

func (r *CommentRepository) LikeComment(commentID, userID string) (bool, error) {
	var existingLike CommentLikeModel
	err := r.db.Where("comment_id = ? AND user_id = ?", commentID, userID).First(&existingLike).Error

	if err == nil {
		if err := r.db.Delete(&existingLike).Error; err != nil {
			return false, err
		}
		r.db.Model(&CommentModel{}).
			Where("id = ?", commentID).
			UpdateColumn("likes_count", gorm.Expr("GREATEST(likes_count - 1, 0)"))
		return false, nil
	}

	like := CommentLikeModel{
		CommentID: commentID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	if err := r.db.Create(&like).Error; err != nil {
		return false, err
	}

	r.db.Model(&CommentModel{}).
		Where("id = ?", commentID).
		UpdateColumn("likes_count", gorm.Expr("likes_count + ?", 1))

	return true, nil
}

func (r *CommentRepository) CheckUserCommented(postID, userID string) (bool, error) {
	var count int64
	err := r.db.Model(&CommentModel{}).
		Where("post_id = ? AND user_id = ? AND deleted_at IS NULL", postID, userID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *CommentRepository) GetCommentsCount(postID string) (int, error) {
	var count int64
	err := r.db.Model(&CommentModel{}).
		Where("post_id = ? AND deleted_at IS NULL", postID).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return int(count), nil
}
