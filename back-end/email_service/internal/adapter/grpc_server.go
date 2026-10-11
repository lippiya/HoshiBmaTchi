package adapter

import (
	"context"
	"log"
	"net"

	emailpb "hoshiBmaTchi/back-end/proto"
	service "hoshiBmaTchi/back-end/email_service/internal/core/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	emailpb.UnimplementedEmailServiceServer
	service *service.EmailService
}

func NewGRPCServer(s *service.EmailService) *GRPCServer {
	return &GRPCServer{service: s}
}

func (g *GRPCServer) QueueOTPEmail(ctx context.Context, req *emailpb.OTPRequest) (*emailpb.EmptyResponse, error) {
	if req.GetRecipient() == "" || req.GetCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "recipient and code are required")
	}

	if err := g.service.QueueOTPEmail(req.GetRecipient(), req.GetCode()); err != nil {
		log.Printf("QueueOTPEmail failed: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emailpb.EmptyResponse{Ok: true, Message: "queued"}, nil
}

func (g *GRPCServer) QueueWelcomeEmail(ctx context.Context, req *emailpb.GenericEmail) (*emailpb.EmptyResponse, error) {
	if req.GetRecipient() == "" || req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "recipient and name are required")
	}

	if err := g.service.QueueWelcomeEmail(req.GetRecipient(), req.GetName()); err != nil {
		log.Printf("QueueWelcomeEmail failed: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emailpb.EmptyResponse{Ok: true, Message: "queued"}, nil
}

func (g *GRPCServer) QueueLoginEmail(ctx context.Context, req *emailpb.GenericEmail) (*emailpb.EmptyResponse, error) {
	if req.GetRecipient() == "" || req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "recipient and name are required")
	}

	if err := g.service.QueueLoginEmail(req.GetRecipient(), req.GetName()); err != nil {
		log.Printf("QueueLoginEmail failed: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emailpb.EmptyResponse{Ok: true, Message: "queued"}, nil
}

func RunGRPC(addr string, gs *GRPCServer) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s := grpc.NewServer()
	emailpb.RegisterEmailServiceServer(s, gs)
	log.Printf("gRPC email server listening %s", addr)
	return s.Serve(lis)
}