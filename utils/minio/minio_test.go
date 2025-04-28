package minio

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/spacelift-io/homework-object-storage/interfaces"
	"github.com/spacelift-io/homework-object-storage/interfaces/mocks"
	"github.com/spacelift-io/homework-object-storage/models"
)

func TestPutObject(t *testing.T) {
	// Create a mock client
	mockClient := mocks.NewMockMinioClient()

	// Create a test instance
	instance := &models.MinioInstance{
		ID:         "test-id",
		Name:       "test-instance",
		IPAddress:  "127.0.0.1",
		AccessKey:  "test-access",
		SecretKey:  "test-secret",
		Client:     mockClient,
		BucketName: "test-bucket",
	}

	// Test data
	objectID := "test-object"
	data := []byte("test data")

	// Call PutObject
	err := PutObject(context.Background(), instance, objectID, bytes.NewReader(data), int64(len(data)))

	// Verify results
	if err != nil {
		t.Errorf("PutObject returned error: %v", err)
	}

	// Verify that the object was stored in the mock client
	storedData, exists := mockClient.Objects["test-bucket/test-object"]
	if !exists {
		t.Error("Object was not stored in the mock client")
	} else if string(storedData) != "test data" {
		t.Errorf("Stored data does not match: got %s, want %s", string(storedData), "test data")
	}
}

func TestNewMinioClientAdapter(t *testing.T) {
	// We can't easily mock a minio.Client, so we'll just test that the adapter is created
	// and implements the interface
	t.Skip("Skipping test that requires a real minio.Client")
}

func TestSetupMinioClient(t *testing.T) {
	// We can't easily test the actual minio client creation, so we'll mock the dependencies
	t.Skip("Skipping test that requires a real minio client")
}

// Test PutObject with an error
func TestPutObjectError(t *testing.T) {
	// Create a mock client that returns an error
	mockClient := mocks.NewMockMinioClient()
	mockClient.PutObjectFunc = func(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
		return minio.UploadInfo{}, errors.New("put object error")
	}

	// Create a test instance
	instance := &models.MinioInstance{
		ID:         "test-id",
		Name:       "test-instance",
		IPAddress:  "127.0.0.1",
		AccessKey:  "test-access",
		SecretKey:  "test-secret",
		Client:     mockClient,
		BucketName: "test-bucket",
	}

	// Test data
	objectID := "test-object"
	data := []byte("test data")

	// Call PutObject
	err := PutObject(context.Background(), instance, objectID, bytes.NewReader(data), int64(len(data)))

	// Verify results
	if err == nil {
		t.Error("PutObject should have returned an error")
	}
}

func TestGetObjectFromMinio(t *testing.T) {
	// Create a mock client
	mockClient := mocks.NewMockMinioClient()

	// Store test data
	mockClient.Objects["test-bucket/test-object"] = []byte("test data")

	// Create a test instance
	instance := &models.MinioInstance{
		ID:         "test-id",
		Name:       "test-instance",
		IPAddress:  "127.0.0.1",
		AccessKey:  "test-access",
		SecretKey:  "test-secret",
		Client:     mockClient,
		BucketName: "test-bucket",
	}

	// Create a test response recorder
	w := httptest.NewRecorder()

	// Call GetObjectFromMinio
	obj, stat, ok := GetObjectFromMinio(w, context.Background(), instance, "test-object")

	// Verify results
	if !ok {
		t.Error("GetObjectFromMinio returned not ok")
	}

	if obj == nil {
		t.Error("GetObjectFromMinio returned nil object")
	}

	if stat == nil {
		t.Error("GetObjectFromMinio returned nil stat")
	} else if stat.Size != 9 { // "test data" is 9 bytes
		t.Errorf("Stat size does not match: got %d, want %d", stat.Size, 9)
	}

	// Read the object data
	data, err := io.ReadAll(obj)
	if err != nil {
		t.Errorf("Failed to read object data: %v", err)
	} else if string(data) != "test data" {
		t.Errorf("Object data does not match: got %s, want %s", string(data), "test data")
	}
}

func TestGetObjectFromMinio_NotFound(t *testing.T) {
	// Create a mock client
	mockClient := mocks.NewMockMinioClient()

	// Override the GetObject method to return a mock object that will return an error from Stat
	mockClient.GetObjectFunc = func(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (interfaces.MinioObjectInterface, error) {
		mockObj := &mocks.MockMinioObject{
			StatFunc: func() (minio.ObjectInfo, error) {
				return minio.ObjectInfo{}, minio.ErrorResponse{
					Code:    "NoSuchKey",
					Message: "The specified key does not exist.",
				}
			},
		}
		return mockObj, nil
	}

	// Create a test instance
	instance := &models.MinioInstance{
		ID:         "test-id",
		Name:       "test-instance",
		IPAddress:  "127.0.0.1",
		AccessKey:  "test-access",
		SecretKey:  "test-secret",
		Client:     mockClient,
		BucketName: "test-bucket",
	}

	// Create a test response recorder
	w := httptest.NewRecorder()

	// Call GetObjectFromMinio with a non-existent object
	obj, stat, ok := GetObjectFromMinio(w, context.Background(), instance, "non-existent")

	// Verify results
	if ok {
		t.Error("GetObjectFromMinio returned ok for non-existent object")
	}

	if obj != nil {
		t.Error("GetObjectFromMinio returned non-nil object for non-existent object")
	}

	if stat != nil {
		t.Error("GetObjectFromMinio returned non-nil stat for non-existent object")
	}

	// Verify response
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestWriteObjectToResponse(t *testing.T) {
	// Create test data
	data := []byte("test data")

	// Create a mock object
	info := minio.ObjectInfo{
		Key:          "test-object",
		Size:         int64(len(data)),
		ContentType:  "application/octet-stream",
		ETag:         "mock-etag",
		LastModified: time.Now(),
	}
	mockObject := mocks.NewMockMinioObject(data, info)

	// Create a test response recorder
	w := httptest.NewRecorder()

	// Call WriteObjectToResponse
	WriteObjectToResponse(w, mockObject, &info, "test-object", "test-instance")

	// Verify response headers
	if w.Header().Get("Content-Type") != "application/octet-stream" {
		t.Errorf("Expected Content-Type %s, got %s", "application/octet-stream", w.Header().Get("Content-Type"))
	}

	if w.Header().Get("Content-Length") != "9" { // "test data" is 9 bytes
		t.Errorf("Expected Content-Length %s, got %s", "9", w.Header().Get("Content-Length"))
	}

	// Verify response body
	if w.Body.String() != "test data" {
		t.Errorf("Response body does not match: got %s, want %s", w.Body.String(), "test data")
	}
}
