package miniostorage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ---- МОК КЛИЕНТА ----
type MockMinioClient struct {
	mock.Mock
}

func (m *MockMinioClient) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	args := m.Called(ctx, bucketName, objectName, reader, objectSize, opts)
	return minio.UploadInfo{}, args.Error(1)
}

func (m *MockMinioClient) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, bucketName, objectName, opts)
	if rc, ok := args.Get(0).(io.ReadCloser); ok {
		return rc, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockMinioClient) RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error {
	args := m.Called(ctx, bucketName, objectName, opts)
	return args.Error(0)
}

// ---- ТЕСТЫ ----

func TestSaveAndGetCanvas(t *testing.T) {
	ctx := context.Background()
	canvasID := "mock-canvas"
	bucket := "mock-bucket"
	mockClient := new(MockMinioClient)
	storage := New(mockClient, bucket)

	imageData := []byte{0x89, 0x50, 0x4E, 0x47} // PNG signature start

	mockClient.On("PutObject", ctx, bucket, canvasID+".png", mock.Anything, int64(len(imageData)), mock.Anything).Return(minio.UploadInfo{}, nil)
	mockClient.On("GetObject", ctx, bucket, canvasID+".png", mock.Anything).
		Return(io.NopCloser(bytes.NewReader(imageData)), nil)

	url, err := storage.SaveCanvas(ctx, canvasID, imageData)
	assert.NoError(t, err)
	assert.Contains(t, url, canvasID+".png")

	result, err := storage.GetCanvas(ctx, canvasID)
	assert.NoError(t, err)
	assert.Equal(t, imageData, result)
}

func TestDeleteCanvas(t *testing.T) {
	ctx := context.Background()
	canvasID := "mock-delete"
	bucket := "mock-bucket"
	mockClient := new(MockMinioClient)
	storage := New(mockClient, bucket)

	mockClient.On("RemoveObject", ctx, bucket, canvasID+".png", mock.Anything).Return(nil)

	err := storage.DeleteCanvas(ctx, canvasID)
	assert.NoError(t, err)
}

func TestGetCanvas_Error(t *testing.T) {
	ctx := context.Background()
	canvasID := "missing"
	bucket := "mock-bucket"
	mockClient := new(MockMinioClient)
	storage := New(mockClient, bucket)

	mockClient.On("GetObject", ctx, bucket, canvasID+".png", mock.Anything).Return(nil, errors.New("not found"))

	data, err := storage.GetCanvas(ctx, canvasID)
	assert.Error(t, err)
	assert.Nil(t, data)
}
