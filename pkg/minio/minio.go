package minio

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

type MinioConfig struct {
	Endpoint  string `yaml:"MINIO_HOST" env-required:"true"`
	Port      string `yaml:"MINIO_PORT" env-required:"true"`
	AccessKey string `yaml:"MINIO_USER" env-required:"true"`
	SecretKey string `yaml:"MINIO_PASS" env-required:"true"`
	Bucket    string `yaml:"MINIO_BUCKET" env-required:"true"`
	Sslmode   bool   `yaml:"MINIO_SSLMODE" env-required:"true"`
}

func New(ctx context.Context, log *zap.Logger, cfg MinioConfig) (*minio.Client, error) {
	endpoint := fmt.Sprintf("%s:%s", cfg.Endpoint, cfg.Port)
	accessKeyID := cfg.AccessKey
	secretAccessKey := cfg.SecretKey
	useSSL := false

	// Инициализируем новый клиент MinIO
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})

	if err != nil {
		log.Error("failed to create MinIO client", zap.Error(err))
		return nil, fmt.Errorf("unable to connect to MinIO: %w", err)
	}

	// Проверяем бакет, если он не существует, создаем его
	bucketName := cfg.Bucket
	// location := "us-east-1" // MinIO не требует указания региона, но мы можем указать его для совместимости с AWS S3

	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Info("We already own", zap.String("bucket", bucketName))
		} else {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	log.Info("Successfully created minIO", zap.String("bucket", bucketName))

	// Устанавливаем публичную политику
	publicPolicy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect":    "Allow",
				"Principal": "*",
				"Action":    []string{"s3:GetObject"},
				"Resource":  fmt.Sprintf("arn:aws:s3:::%s/*", bucketName),
			},
		},
	}

	policyBytes, err := json.Marshal(publicPolicy)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public bucket policy: %w", err)
	}

	err = minioClient.SetBucketPolicy(ctx, bucketName, string(policyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to set bucket policy: %w", err)
	}

	log.Info("Bucket policy set to public", zap.String("bucket", bucketName))

	return minioClient, nil
}
