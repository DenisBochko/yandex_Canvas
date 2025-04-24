package miniostorage

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
)

// MinioClientWrapper — адаптер, который приводит *minio.Client к интерфейсу MinioClient
type MinioClientWrapper struct {
	Client *minio.Client
}

func (w *MinioClientWrapper) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return w.Client.PutObject(ctx, bucketName, objectName, reader, objectSize, opts)
}

func (w *MinioClientWrapper) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error) {
	obj, err := w.Client.GetObject(ctx, bucketName, objectName, opts)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (w *MinioClientWrapper) RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error {
	return w.Client.RemoveObject(ctx, bucketName, objectName, opts)
}
