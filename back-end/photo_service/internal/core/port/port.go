package port

import "hoshiBmaTchi/back-end/photo_service/internal/core/domain"

type PhotoService interface {
	UploadPhoto(upload domain.PhotoUpload) (*domain.PhotoResult, error)
	DeletePhoto(filePath string) error
	GetPhotoURL(filePath string) (string, error)
}

type MinioRepository interface {
	Upload(upload domain.PhotoUpload) (string, string, error)
	Delete(filePath string) error
	GetPhoto(filePath string) (string, error)
}
