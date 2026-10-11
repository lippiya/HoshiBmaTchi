package grpcclient

import (
	"context"
	"fmt"

	"hoshiBmaTchi/back-end/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SocialClient struct {
	client proto.SocialServiceClient
	conn   *grpc.ClientConn
}

func NewSocialClient(grpcAddr string) (*SocialClient, error) {
	conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to social service: %w", err)
	}

	client := proto.NewSocialServiceClient(conn)

	return &SocialClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *SocialClient) Close() error {
	return c.conn.Close()
}

func (c *SocialClient) Follow(ctx context.Context, followerID, followingID string) (*proto.FollowResponse, error) {
	req := &proto.FollowRequest{
		FollowerId:  followerID,
		FollowingId: followingID,
	}

	return c.client.Follow(ctx, req)
}

func (c *SocialClient) Unfollow(ctx context.Context, followerID, followingID string) (*proto.UnfollowResponse, error) {
	req := &proto.UnfollowRequest{
		FollowerId:  followerID,
		FollowingId: followingID,
	}

	return c.client.Unfollow(ctx, req)
}

func (c *SocialClient) IsFollowing(ctx context.Context, followerID, followingID string) (bool, error) {
	req := &proto.IsFollowingRequest{
		FollowerId:  followerID,
		FollowingId: followingID,
	}

	resp, err := c.client.IsFollowing(ctx, req)
	if err != nil {
		return false, err
	}

	return resp.IsFollowing, nil
}

func (c *SocialClient) GetFollowersCount(ctx context.Context, userID string) (int32, error) {
	req := &proto.GetFollowersCountRequest{
		UserId: userID,
	}

	resp, err := c.client.GetFollowersCount(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.Count, nil
}

func (c *SocialClient) GetFollowingCount(ctx context.Context, userID string) (int32, error) {
	req := &proto.GetFollowingCountRequest{
		UserId: userID,
	}

	resp, err := c.client.GetFollowingCount(ctx, req)
	if err != nil {
		return 0, err
	}

	return resp.Count, nil
}

func (c *SocialClient) GetFollowing(ctx context.Context, userID string) ([]*proto.User, error) {
	req := &proto.GetFollowingRequest{
		UserId: userID,
	}

	resp, err := c.client.GetFollowing(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Users, nil
}
