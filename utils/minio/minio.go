package minio

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/spacelift-io/homework-object-storage/interfaces"
	"github.com/spacelift-io/homework-object-storage/models"
)

// SetupMinioClient creates a Minio client and ensures the bucket exists
func SetupMinioClient(ctx context.Context, ipAddress, accessKey, secretKey, bucketName string) (interfaces.MinioClientInterface, error) {
	// Create Minio client
	minioClient, err := minio.New(fmt.Sprintf("%s:9000", ipAddress), &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Minio client: %w", err)
	}

	// Check if a bucket exists, create if it doesn't
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	// Wrap the client in an adapter to implement the interface
	return NewMinioClientAdapter(minioClient), nil
}

// PutObject uploads an object to Minio
func PutObject(ctx context.Context, instance *models.MinioInstance, objectID string, data io.Reader, size int64) error {
	_, err := instance.Client.PutObject(
		ctx,
		instance.BucketName,
		objectID,
		data,
		size,
		minio.PutObjectOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to put object: %w", err)
	}
	return nil
}

// GetObjectFromMinio retrieves an object from Minio and handles errors
func GetObjectFromMinio(w http.ResponseWriter, ctx context.Context, instance *models.MinioInstance, objectID string) (interfaces.MinioObjectInterface, *minio.ObjectInfo, bool) {
	// Get object from Minio
	object, err := instance.Client.GetObject(
		ctx,
		instance.BucketName,
		objectID,
		minio.GetObjectOptions{},
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get object: %v", err), http.StatusInternalServerError)
		return nil, nil, false
	}

	// Check if an object exists
	stat, err := object.Stat()
	if err != nil {
		// Check if it's a "not found" error
		if errResp, ok := err.(minio.ErrorResponse); ok && errResp.Code == "NoSuchKey" {
			http.Error(w, "Object not found", http.StatusNotFound)
		} else if err.Error() == "The specified key does not exist." {
			// Fallback for backward compatibility
			http.Error(w, "Object not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to stat object: %v", err), http.StatusInternalServerError)
		}
		err := object.Close()
		if err != nil {
			log.Printf("Failed to close object: %v", err)
		}
		return nil, nil, false
	}

	return object, &stat, true
}

// WriteObjectToResponse writes the object data to the HTTP response
func WriteObjectToResponse(w http.ResponseWriter, object interfaces.MinioObjectInterface, stat *minio.ObjectInfo, objectID, instanceName string) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size))

	_, err := io.Copy(w, object)
	if err != nil {
		log.Printf("Failed to copy object data to response: %v", err)
		return
	}

	log.Printf("Successfully retrieved object %s from instance %s", objectID, instanceName)
}
