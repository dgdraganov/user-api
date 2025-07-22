package minio

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	minioClient *minio.Client
}

func NewMinioClient(endpoint, accessKey, secretKey string) (*MinioClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	return &MinioClient{
		minioClient: client,
	}, nil
}

func (m *MinioClient) CreateBucket(ctx context.Context, bucketName string) error {
	exists, err := m.minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("check if bucket exists: %w", err)
	}
	if exists {
		return nil
	}

	err = m.minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		return fmt.Errorf("create bucket: %w", err)
	}

	return nil
}

func (m *MinioClient) UploadFile(ctx context.Context, bucketName, objectName string, file io.Reader, fileSize int64) error {
	_, err := m.minioClient.PutObject(ctx, bucketName, objectName, file, fileSize, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("bucket put object: %w", err)
	}

	return nil
}
