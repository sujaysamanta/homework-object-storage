package minio

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/spacelift-io/homework-object-storage/interfaces"
)

// MinioClientAdapter adapts a minio.Client to implement interfaces.MinioClientInterface
type MinioClientAdapter struct {
	client *minio.Client
}

// NewMinioClientAdapter creates a new adapter for a minio.Client
func NewMinioClientAdapter(client *minio.Client) interfaces.MinioClientInterface {
	return &MinioClientAdapter{
		client: client,
	}
}

// PutObject implements interfaces.MinioClientInterface.PutObject
func (a *MinioClientAdapter) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return a.client.PutObject(ctx, bucketName, objectName, reader, objectSize, opts)
}

// GetObject implements interfaces.MinioClientInterface.GetObject
func (a *MinioClientAdapter) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (interfaces.MinioObjectInterface, error) {
	obj, err := a.client.GetObject(ctx, bucketName, objectName, opts)
	if err != nil {
		return nil, err
	}

	// The minio.Object already implements interfaces.MinioObjectInterface
	return obj, nil
}

// BucketExists implements interfaces.MinioClientInterface.BucketExists
func (a *MinioClientAdapter) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	return a.client.BucketExists(ctx, bucketName)
}

// MakeBucket implements interfaces.MinioClientInterface.MakeBucket
func (a *MinioClientAdapter) MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error {
	return a.client.MakeBucket(ctx, bucketName, opts)
}
