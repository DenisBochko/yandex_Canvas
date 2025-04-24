package miniostorage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
)

type MinioClient interface {
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error)
	RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error
}

type MinioStorage struct {
	client     MinioClient
	bucketName string
}

func New(client MinioClient, bucketName string) *MinioStorage {
	return &MinioStorage{
		client:     client,
		bucketName: bucketName,
	}
}

// SaveCanvas сохраняет canvas в MinIO
func (s *MinioStorage) SaveCanvas(ctx context.Context, canvasID string, imageData []byte) (string, error) {
	objectName := fmt.Sprintf("%s.png", canvasID)
	_, err := s.client.PutObject(ctx, s.bucketName, objectName, bytes.NewReader(imageData), int64(len(imageData)), minio.PutObjectOptions{
		ContentType: "image/png",
	})

	url := fmt.Sprintf("http://localhost:9000/%s/%s", s.bucketName, objectName)
	return url, err
}

// GetCanvas получает canvas по ID
func (s *MinioStorage) GetCanvas(ctx context.Context, canvasID string) ([]byte, error) {
	objectName := fmt.Sprintf("%s.png", canvasID)
	obj, err := s.client.GetObject(ctx, s.bucketName, objectName, minio.GetObjectOptions{})
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
	return s.client.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
}
