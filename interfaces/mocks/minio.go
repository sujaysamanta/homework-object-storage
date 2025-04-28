package mocks

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/spacelift-io/homework-object-storage/interfaces"
)

// MockMinioClient is a mock implementation of the MinioClientInterface
type MockMinioClient struct {
	PutObjectFunc    func(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	GetObjectFunc    func(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (interfaces.MinioObjectInterface, error)
	BucketExistsFunc func(ctx context.Context, bucketName string) (bool, error)
	MakeBucketFunc   func(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error
	Objects          map[string][]byte // Map of object data by key (bucketName/objectName)
}

// Ensure MockMinioClient implements MinioClientInterface
var _ interfaces.MinioClientInterface = (*MockMinioClient)(nil)

// PutObject implements the MinioClientInterface.PutObject method
func (m *MockMinioClient) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	if m.PutObjectFunc != nil {
		return m.PutObjectFunc(ctx, bucketName, objectName, reader, objectSize, opts)
	}

	// Default implementation: store the object data in memory
	if m.Objects == nil {
		m.Objects = make(map[string][]byte)
	}

	key := bucketName + "/" + objectName
	data, err := io.ReadAll(reader)
	if err != nil {
		return minio.UploadInfo{}, err
	}

	m.Objects[key] = data

	return minio.UploadInfo{
		Bucket:       bucketName,
		Key:          objectName,
		ETag:         "mock-etag",
		Size:         int64(len(data)),
		LastModified: time.Now(),
	}, nil
}

// GetObject implements the MinioClientInterface.GetObject method
func (m *MockMinioClient) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (interfaces.MinioObjectInterface, error) {
	if m.GetObjectFunc != nil {
		return m.GetObjectFunc(ctx, bucketName, objectName, opts)
	}

	// Default implementation: retrieve the object data from memory
	key := bucketName + "/" + objectName
	data, exists := m.Objects[key]
	if !exists {
		return nil, minio.ErrorResponse{
			Code:    "NoSuchKey",
			Message: "The specified key does not exist.",
		}
	}

	// Create a mock object
	info := minio.ObjectInfo{
		Key:          objectName,
		Size:         int64(len(data)),
		ContentType:  "application/octet-stream",
		ETag:         "mock-etag",
		LastModified: time.Now(),
	}

	return NewMockMinioObject(data, info), nil
}

// BucketExists implements the MinioClientInterface.BucketExists method
func (m *MockMinioClient) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	if m.BucketExistsFunc != nil {
		return m.BucketExistsFunc(ctx, bucketName)
	}

	// Default implementation: always return true
	return true, nil
}

// MakeBucket implements the MinioClientInterface.MakeBucket method
func (m *MockMinioClient) MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error {
	if m.MakeBucketFunc != nil {
		return m.MakeBucketFunc(ctx, bucketName, opts)
	}

	// Default implementation: do nothing
	return nil
}

// MockMinioObject is a mock implementation of the MinioObjectInterface
type MockMinioObject struct {
	StatFunc  func() (minio.ObjectInfo, error)
	ReadFunc  func(p []byte) (n int, err error)
	CloseFunc func() error
	Data      []byte
	Position  int
	Info      minio.ObjectInfo
	Reader    *bytes.Reader
}

// Ensure MockMinioObject implements MinioObjectInterface
var _ interfaces.MinioObjectInterface = (*MockMinioObject)(nil)

// Stat implements the MinioObjectInterface.Stat method
func (m *MockMinioObject) Stat() (minio.ObjectInfo, error) {
	if m.StatFunc != nil {
		return m.StatFunc()
	}

	// Default implementation: return the stored info
	return m.Info, nil
}

// Read implements the MinioObjectInterface.Read method
func (m *MockMinioObject) Read(p []byte) (n int, err error) {
	if m.ReadFunc != nil {
		return m.ReadFunc(p)
	}

	// Default implementation: read from the reader
	return m.Reader.Read(p)
}

// Close implements the MinioObjectInterface.Close method
func (m *MockMinioObject) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}

	// Default implementation: do nothing
	return nil
}

// NewMockMinioClient creates a new MockMinioClient
func NewMockMinioClient() *MockMinioClient {
	return &MockMinioClient{
		Objects: make(map[string][]byte),
	}
}

// NewMockMinioObject creates a new MockMinioObject
func NewMockMinioObject(data []byte, info minio.ObjectInfo) *MockMinioObject {
	return &MockMinioObject{
		Data:     data,
		Position: 0,
		Info:     info,
		Reader:   bytes.NewReader(data),
	}
}
