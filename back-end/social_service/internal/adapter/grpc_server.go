package adapter

import (
	"context"
	"hoshiBmaTchi/back-end/proto"
	"hoshiBmaTchi/back-end/social_service/internal/core/domain"
	"hoshiBmaTchi/back-end/social_service/internal/core/port"
)

type SocialGRPCServer struct {
	proto.UnimplementedSocialServiceServer
	followService  port.FollowServicePort
	commentService port.CommentServicePort
}

func NewSocialGRPCServer(followService port.FollowServicePort, commentService port.CommentServicePort) *SocialGRPCServer {
	return &SocialGRPCServer{
		followService:  followService,
		commentService: commentService,
	}
}

func (s *SocialGRPCServer) Follow(ctx context.Context, req *proto.FollowRequest) (*proto.FollowResponse, error) {
	err := s.followService.FollowUser(ctx, req.FollowerId, req.FollowingId)
	if err != nil {
		return &proto.FollowResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	followersCount, _ := s.followService.GetFollowersCount(ctx, req.FollowingId)
	followingCount, _ := s.followService.GetFollowingCount(ctx, req.FollowerId)

	return &proto.FollowResponse{
		Success:        true,
		Message:        "Successfully followed user",
		FollowersCount: int32(followersCount),
		FollowingCount: int32(followingCount),
	}, nil
}

func (s *SocialGRPCServer) Unfollow(ctx context.Context, req *proto.UnfollowRequest) (*proto.UnfollowResponse, error) {
	err := s.followService.UnfollowUser(ctx, req.FollowerId, req.FollowingId)
	if err != nil {
		return &proto.UnfollowResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	followersCount, _ := s.followService.GetFollowersCount(ctx, req.FollowingId)
	followingCount, _ := s.followService.GetFollowingCount(ctx, req.FollowerId)

	return &proto.UnfollowResponse{
		Success:        true,
		Message:        "Successfully unfollowed user",
		FollowersCount: int32(followersCount),
		FollowingCount: int32(followingCount),
	}, nil
}

func (s *SocialGRPCServer) IsFollowing(ctx context.Context, req *proto.IsFollowingRequest) (*proto.IsFollowingResponse, error) {
	isFollowing, err := s.followService.IsFollowing(ctx, req.FollowerId, req.FollowingId)
	if err != nil {
		return &proto.IsFollowingResponse{
			IsFollowing: false,
		}, nil
	}

	return &proto.IsFollowingResponse{
		IsFollowing: isFollowing,
	}, nil
}

func (s *SocialGRPCServer) GetFollowersCount(ctx context.Context, req *proto.GetFollowersCountRequest) (*proto.GetFollowersCountResponse, error) {
	count, err := s.followService.GetFollowersCount(ctx, req.UserId)
	if err != nil {
		return &proto.GetFollowersCountResponse{Count: 0}, nil
	}

	return &proto.GetFollowersCountResponse{
		Count: int32(count),
	}, nil
}

func (s *SocialGRPCServer) GetFollowingCount(ctx context.Context, req *proto.GetFollowingCountRequest) (*proto.GetFollowingCountResponse, error) {
	count, err := s.followService.GetFollowingCount(ctx, req.UserId)
	if err != nil {
		return &proto.GetFollowingCountResponse{Count: 0}, nil
	}

	return &proto.GetFollowingCountResponse{
		Count: int32(count),
	}, nil
}

func (s *SocialGRPCServer) GetFollowing(ctx context.Context, req *proto.GetFollowingRequest) (*proto.GetFollowingResponse, error) {
	following, err := s.followService.GetFollowing(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	var users []*proto.User
	for _, u := range following {
		users = append(users, &proto.User{
			Id:             u.ID,
			Username:       u.Username,
			ProfilePicture: u.ProfilePicture,
		})
	}

	return &proto.GetFollowingResponse{Users: users}, nil
}

func (s *SocialGRPCServer) CreateComment(ctx context.Context, req *proto.CreateCommentRequest) (*proto.CreateCommentResponse, error) {
	var parentCommentID *string
	if req.ParentCommentId != "" {
		parentCommentID = &req.ParentCommentId
	}

	comment, err := s.commentService.CreateComment(domain.CreateCommentRequest{
		PostID:          req.PostId,
		ParentCommentID: parentCommentID,
		UserID:          req.UserId,
		Text:            req.Text,
		GifUrl:          req.GifUrl,
	})

	if err != nil {
		return nil, err
	}

	parentID := ""
	if comment.ParentCommentID != nil {
		parentID = *comment.ParentCommentID
	}

	return &proto.CreateCommentResponse{
		Id:              comment.ID,
		PostId:          comment.PostID,
		ParentCommentId: parentID,
		UserId:          comment.UserID,
		Text:            comment.Text,
		GifUrl:          comment.GifUrl,
		LikesCount:      int32(comment.LikesCount),
		RepliesCount:    int32(comment.RepliesCount),
		CreatedAt:       comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *SocialGRPCServer) GetComments(ctx context.Context, req *proto.GetCommentsRequest) (*proto.GetCommentsResponse, error) {
	comments, err := s.commentService.GetComments(req.PostId, int(req.Page), int(req.Limit), req.Sort)
	if err != nil {
		return nil, err
	}

	protoComments := make([]*proto.Comment, len(comments))
	for i, c := range comments {
		parentID := ""
		if c.ParentCommentID != nil {
			parentID = *c.ParentCommentID
		}

		protoComments[i] = &proto.Comment{
			Id:              c.ID,
			PostId:          c.PostID,
			ParentCommentId: parentID,
			UserId:          c.UserID,
			Text:            c.Text,
			GifUrl:          c.GifUrl,
			LikesCount:      int32(c.LikesCount),
			RepliesCount:    int32(c.RepliesCount),
			CreatedAt:       c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return &proto.GetCommentsResponse{
		Comments: protoComments,
	}, nil
}

func (s *SocialGRPCServer) GetCommentReplies(ctx context.Context, req *proto.GetCommentRepliesRequest) (*proto.GetCommentRepliesResponse, error) {
	replies, err := s.commentService.GetCommentReplies(req.CommentId)
	if err != nil {
		return nil, err
	}

	protoReplies := make([]*proto.Comment, len(replies))
	for i, r := range replies {
		parentID := ""
		if r.ParentCommentID != nil {
			parentID = *r.ParentCommentID
		}

		protoReplies[i] = &proto.Comment{
			Id:              r.ID,
			PostId:          r.PostID,
			ParentCommentId: parentID,
			UserId:          r.UserID,
			Text:            r.Text,
			GifUrl:          r.GifUrl,
			LikesCount:      int32(r.LikesCount),
			RepliesCount:    int32(r.RepliesCount),
			CreatedAt:       r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return &proto.GetCommentRepliesResponse{
		Replies: protoReplies,
	}, nil
}

func (s *SocialGRPCServer) DeleteComment(ctx context.Context, req *proto.DeleteCommentRequest) (*proto.DeleteCommentResponse, error) {
	err := s.commentService.DeleteComment(req.CommentId, req.UserId)
	if err != nil {
		return &proto.DeleteCommentResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &proto.DeleteCommentResponse{
		Success: true,
		Message: "Comment deleted successfully",
	}, nil
}

func (s *SocialGRPCServer) LikeComment(ctx context.Context, req *proto.LikeCommentRequest) (*proto.LikeCommentResponse, error) {
	liked, err := s.commentService.LikeComment(req.CommentId, req.UserId)
	if err != nil {
		return nil, err
	}

	message := "Comment liked"
	if !liked {
		message = "Comment unliked"
	}

	return &proto.LikeCommentResponse{
		Liked:   liked,
		Message: message,
	}, nil
}

func (s *SocialGRPCServer) CheckUserCommented(ctx context.Context, req *proto.CheckUserCommentedRequest) (*proto.CheckUserCommentedResponse, error) {
	commented, err := s.commentService.CheckUserCommented(req.PostId, req.UserId)
	if err != nil {
		return &proto.CheckUserCommentedResponse{Commented: false}, nil
	}

	return &proto.CheckUserCommentedResponse{
		Commented: commented,
	}, nil
}

func (s *SocialGRPCServer) GetCommentsCount(ctx context.Context, req *proto.GetCommentsCountRequest) (*proto.GetCommentsCountResponse, error) {
	count, err := s.commentService.GetCommentsCount(req.PostId)
	if err != nil {
		return &proto.GetCommentsCountResponse{Count: 0}, nil
	}

	return &proto.GetCommentsCountResponse{
		Count: int32(count),
	}, nil
}
