package grpcclient

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"hoshiBmaTchi/back-end/proto"
)

type PhotoClient struct {
	client proto.PhotoServiceClient
	conn   *grpc.ClientConn
}

func NewPhotoClient(grpcAddr string) (*PhotoClient, error) {
	conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to photo service: %w", err)
	}

	client := proto.NewPhotoServiceClient(conn)

	return &PhotoClient{
		client: client,
		conn: conn,
	}, nil
}

func (c *PhotoClient) Close() error {
	return c.conn.Close()
}

func (c *PhotoClient) UploadPhoto(file interface{}, userID string) (string, error) {
	fileHeader, ok := file.(*multipart.FileHeader)
	if !ok {
		return "", fmt.Errorf("invalid file type: expected *multipart.FileHeader")
	}
	
	f, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()
	fileData, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	resp, err := c.client.UploadPhoto(context.Background(), &proto.UploadPhotoRequest{
		FileData: fileData,
		FileName: fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		UserId: userID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload photo: %w", err)
	}

	if resp.Error != "" {
		return "", fmt.Errorf("photo service error: %s", resp.Error)
	}

	return resp.FilePath, nil
}

func (c *PhotoClient) DeletePhoto(filePath string) error {
	resp, err := c.client.DeletePhoto(context.Background(), &proto.DeletePhotoRequest{
		FilePath: filePath,
	})
	if err != nil {
		return fmt.Errorf("failed to delete photo: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("photo service error: %s", resp.Error)
	}

	return nil
}

func (c *PhotoClient) GetPhotoURL(filePath string) (string, error) {
	resp, err := c.client.GetPhotoURL(context.Background(), &proto.GetPhotoURLRequest{
		FilePath: filePath,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get photo URL: %w", err)
	}

	if resp.Error != "" {
		return "", fmt.Errorf("photo service error: %s", resp.Error)
	}

	return resp.Url, nil
}
