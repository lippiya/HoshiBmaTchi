package service

import (
	"fmt"
	"hoshiBmaTchi/back-end/photo_service/internal/core/domain"
	"hoshiBmaTchi/back-end/photo_service/internal/core/port"
)

type PhotoService struct {
	minioRepo port.MinioRepository
}

func NewPhotoService(minioRepo port.MinioRepository) *PhotoService {
	return &PhotoService{
		minioRepo: minioRepo,
	}
}

func (s *PhotoService) UploadPhoto(upload domain.PhotoUpload) (*domain.PhotoResult, error) {
	
	if len(upload.FileData) == 0 {
		return &domain.PhotoResult{Error: "file data is empty"}, fmt.Errorf("file data is empty")
	}


	url, filePath, err := s.minioRepo.Upload(upload)
	if err != nil {
		return &domain.PhotoResult{Error: err.Error()}, err
	}

	return &domain.PhotoResult{
		URL:      url,
		FilePath: filePath,
	}, nil
}

func (s *PhotoService) DeletePhoto(filePath string) error {
	return s.minioRepo.Delete(filePath)
}

func (s *PhotoService) GetPhotoURL(filePath string) (string, error) {
	return s.minioRepo.GetPhoto(filePath)
}
