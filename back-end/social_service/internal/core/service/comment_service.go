package service

import (
	"hoshiBmaTchi/back-end/social_service/internal/core/domain"
	"hoshiBmaTchi/back-end/social_service/internal/core/port"
)

type CommentService struct {
	repo port.CommentRepositoryPort
}

func NewCommentService(repo port.CommentRepositoryPort) *CommentService {
	return &CommentService{repo: repo}
}

func (s *CommentService) CreateComment(req domain.CreateCommentRequest) (*domain.Comment, error) {
	return s.repo.CreateComment(req)
}

func (s *CommentService) GetComments(postID string, page, limit int, sort string) ([]domain.Comment, error) {
	return s.repo.GetCommentsByPostID(postID, page, limit, sort)
}

func (s *CommentService) GetCommentReplies(commentID string) ([]domain.Comment, error) {
	return s.repo.GetCommentReplies(commentID)
}

func (s *CommentService) DeleteComment(commentID, userID string) error {
	return s.repo.DeleteComment(commentID, userID)
}

func (s *CommentService) LikeComment(commentID, userID string) (bool, error) {
	return s.repo.LikeComment(commentID, userID)
}

func (s *CommentService) CheckUserCommented(postID, userID string) (bool, error) {
	return s.repo.CheckUserCommented(postID, userID)
}

func (s *CommentService) GetCommentsCount(postID string) (int, error) {
	return s.repo.GetCommentsCount(postID)
}
