package miniostorage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
)

type MinioStorage struct {
	minioClient *minio.Client
	bucketName  string
}

func New(minioClient *minio.Client, bucketName string) *MinioStorage {
	return &MinioStorage{
		minioClient: minioClient,
		bucketName:  bucketName,
	}
}

// SaveCanvas сохраняет canvas в MinIO
func (s *MinioStorage) SaveCanvas(ctx context.Context, canvasID string, imageData []byte) (string, error) {
	objectName := fmt.Sprintf("%s.png", canvasID)
	_, err := s.minioClient.PutObject(ctx, s.bucketName, objectName, bytes.NewReader(imageData), int64(len(imageData)), minio.PutObjectOptions{
		ContentType: "image/png",
	})

	url := fmt.Sprintf("%s/%s/%s", s.minioClient.EndpointURL(), s.bucketName, objectName)
	
	return url, err
}

// GetCanvas получает canvas по ID
func (s *MinioStorage) GetCanvas(ctx context.Context, canvasID string) ([]byte, error) {
	objectName := fmt.Sprintf("%s.png", canvasID)
	obj, err := s.minioClient.GetObject(ctx, s.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// DeleteCanvas удаляет canvas по ID
func (s *MinioStorage) DeleteCanvas(ctx context.Context, canvasID string) error {
	objectName := fmt.Sprintf("%s.png", canvasID)
	return s.minioClient.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
}
