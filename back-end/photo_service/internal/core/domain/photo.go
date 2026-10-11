package domain

type PhotoUpload struct {
	FileData []byte
	FileName string
	ContentType string
	UserID string
}

type PhotoResult struct {
	URL string
	FilePath string
	Error string
}
