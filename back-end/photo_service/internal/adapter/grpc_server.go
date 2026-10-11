package grpc_server

import (
	"context"
	"hoshiBmaTchi/back-end/photo_service/internal/core/domain"
	"hoshiBmaTchi/back-end/photo_service/internal/core/port"
	pb "hoshiBmaTchi/back-end/proto"
)

type PhotoServer struct {
	pb.UnimplementedPhotoServiceServer
	photoService port.PhotoService
}

func NewPhotoServer(photoService port.PhotoService) *PhotoServer {
	return &PhotoServer{
		photoService: photoService,
	}
}

func (s *PhotoServer) UploadPhoto(ctx context.Context, req *pb.UploadPhotoRequest) (*pb.UploadPhotoResponse, error) {
	upload := domain.PhotoUpload{
		FileData:    req.FileData,
		FileName:    req.FileName,
		ContentType: req.ContentType,
		UserID:      req.UserId,
	}

	result, err := s.photoService.UploadPhoto(upload)
	if err != nil {
		return &pb.UploadPhotoResponse{
			Error: err.Error(),
		}, nil
	}

	return &pb.UploadPhotoResponse{
		Url:      result.URL,
		FilePath: result.FilePath,
	}, nil
}

func (s *PhotoServer) DeletePhoto(ctx context.Context, req *pb.DeletePhotoRequest) (*pb.DeletePhotoResponse, error) {
	err := s.photoService.DeletePhoto(req.FilePath)
	if err != nil {
		return &pb.DeletePhotoResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.DeletePhotoResponse{
		Success: true,
	}, nil
}

func (s *PhotoServer) GetPhotoURL(ctx context.Context, req *pb.GetPhotoURLRequest) (*pb.GetPhotoURLResponse, error) {
	url, err := s.photoService.GetPhotoURL(req.FilePath)
	if err != nil {
		return &pb.GetPhotoURLResponse{
			Error: err.Error(),
		}, nil
	}

	return &pb.GetPhotoURLResponse{
		Url: url,
	}, nil
}
