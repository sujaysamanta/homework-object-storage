package interfaces

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
)

// MinioClientInterface defines the interface for Minio client operations
type MinioClientInterface interface {
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (MinioObjectInterface, error)
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error
}

// MinioObjectInterface defines the interface for Minio object operations
type MinioObjectInterface interface {
	Stat() (minio.ObjectInfo, error)
	Read(p []byte) (n int, err error)
	Close() error
}

// Note: *minio.Client doesn't directly implement MinioClientInterface
// because GetObject returns *minio.Object instead of MinioObjectInterface.
// We need to use an adapter to make it compatible.

// Ensure that *minio.Object implements MinioObjectInterface
var _ MinioObjectInterface = (*minio.Object)(nil)
