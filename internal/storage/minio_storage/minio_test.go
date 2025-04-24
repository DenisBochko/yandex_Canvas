package miniostorage

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
)

func TestSaveAndGetCanvas_Integration(t *testing.T) {
	ctx := context.Background()
	endpoint := "localhost:9000"
	accessKey := "minio"
	secretKey := "minio123"
	useSSL := false
	bucketName := "test-bucket"
	canvasID := "test-canvas"

	// Инициализация MinIO клиента
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	assert.NoError(t, err)

	// Проверка и создание бакета
	exists, err := client.BucketExists(ctx, bucketName)
	assert.NoError(t, err)
	if !exists {
		err := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		assert.NoError(t, err)
	}

	storage := New(client, bucketName)

	// Создание изображения
	img := image.NewRGBA(image.Rect(0, 0, 1280, 720))
	for x := 100; x < 400; x++ {
		for y := 100; y < 400; y++ {
			img.Set(x, y, color.RGBA{0, 255, 0, 255})
		}
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	assert.NoError(t, err)

	// Сохраняем изображение
	url, err := storage.SaveCanvas(ctx, canvasID, buf.Bytes())
	assert.NoError(t, err)
	assert.Contains(t, url, canvasID+".png")

	// Получаем изображение обратно
	data, err := storage.GetCanvas(ctx, canvasID)
	assert.NoError(t, err)
	assert.True(t, len(data) > 0)

	// Сравниваем байты PNG
	assert.Equal(t, buf.Bytes(), data)

	// Декодируем и проверяем цвет пикселя
	imgFromStorage, err := png.Decode(bytes.NewReader(data))
	assert.NoError(t, err)
	r, g, b, a := imgFromStorage.At(150, 150).RGBA()
	assert.Equal(t, uint32(0), r)
	assert.Equal(t, uint32(65535), g)
	assert.Equal(t, uint32(0), b)
	assert.Equal(t, uint32(65535), a)
}

func TestDeleteCanvas_Integration(t *testing.T) {
	ctx := context.Background()
	endpoint := "localhost:9000"
	accessKey := "minio"
	secretKey := "minio123"
	useSSL := false
	bucketName := "test-bucket"
	canvasID := "test-delete-canvas"

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	assert.NoError(t, err)

	storage := New(client, bucketName)

	// Подготовка и сохранение объекта
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	assert.NoError(t, err)

	_, err = storage.SaveCanvas(ctx, canvasID, buf.Bytes())
	assert.NoError(t, err)

	// Удаление
	err = storage.DeleteCanvas(ctx, canvasID)
	assert.NoError(t, err)

	// Проверка, что объект действительно удалён
	_, err = client.StatObject(ctx, bucketName, canvasID+".png", minio.StatObjectOptions{})
	assert.Error(t, err) // Должна быть ошибка: объект не найден
}