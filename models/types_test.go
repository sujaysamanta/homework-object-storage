package models

import (
	"sync"
	"testing"
	"time"
)

// TestMinioInstanceInitialization tests that a MinioInstance can be properly initialized
func TestMinioInstanceInitialization(t *testing.T) {
	instance := MinioInstance{
		ID:         "test-id",
		Name:       "test-name",
		IPAddress:  "127.0.0.1",
		AccessKey:  "test-access-key",
		SecretKey:  "test-secret-key",
		Client:     nil, // We don't need a real client for this test
		BucketName: "test-bucket",
	}

	// Verify that the fields are set correctly
	if instance.ID != "test-id" {
		t.Errorf("Expected ID to be %s, got %s", "test-id", instance.ID)
	}
	if instance.Name != "test-name" {
		t.Errorf("Expected Name to be %s, got %s", "test-name", instance.Name)
	}
	if instance.IPAddress != "127.0.0.1" {
		t.Errorf("Expected IPAddress to be %s, got %s", "127.0.0.1", instance.IPAddress)
	}
	if instance.AccessKey != "test-access-key" {
		t.Errorf("Expected AccessKey to be %s, got %s", "test-access-key", instance.AccessKey)
	}
	if instance.SecretKey != "test-secret-key" {
		t.Errorf("Expected SecretKey to be %s, got %s", "test-secret-key", instance.SecretKey)
	}
	if instance.BucketName != "test-bucket" {
		t.Errorf("Expected BucketName to be %s, got %s", "test-bucket", instance.BucketName)
	}
}

// TestObjectStorageGatewayInitialization tests that an ObjectStorageGateway can be properly initialized
func TestObjectStorageGatewayInitialization(t *testing.T) {
	// Create a test instance
	instance := MinioInstance{
		ID:         "test-id",
		Name:       "test-name",
		IPAddress:  "127.0.0.1",
		AccessKey:  "test-access-key",
		SecretKey:  "test-secret-key",
		Client:     nil, // We don't need a real client for this test
		BucketName: "test-bucket",
	}

	// Create a gateway with the test instance
	gateway := ObjectStorageGateway{
		MinioInstances: []MinioInstance{instance},
		Mutex:          sync.RWMutex{},
		DockerClient:   nil, // We don't need a real client for this test
		RefreshTicker:  time.NewTicker(1 * time.Second),
	}

	// Verify that the fields are set correctly
	if len(gateway.MinioInstances) != 1 {
		t.Errorf("Expected 1 Minio instance, got %d", len(gateway.MinioInstances))
	}
	if gateway.MinioInstances[0].ID != "test-id" {
		t.Errorf("Expected instance ID to be %s, got %s", "test-id", gateway.MinioInstances[0].ID)
	}
	if gateway.DockerClient != nil {
		t.Errorf("Expected DockerClient to be nil")
	}
	if gateway.RefreshTicker == nil {
		t.Errorf("Expected RefreshTicker to be non-nil")
	}

	// Clean up
	gateway.RefreshTicker.Stop()
}

// TestObjectStorageGatewayMutex tests that the mutex in ObjectStorageGateway works correctly
func TestObjectStorageGatewayMutex(t *testing.T) {
	gateway := ObjectStorageGateway{
		Mutex: sync.RWMutex{},
	}

	// Test that we can lock and unlock the mutex
	gateway.Mutex.Lock()
	gateway.Mutex.Unlock()

	// Test that we can RLock and RUnlock the mutex
	gateway.Mutex.RLock()
	gateway.Mutex.RUnlock()

	// Test that multiple RLocks work
	gateway.Mutex.RLock()
	gateway.Mutex.RLock()
	gateway.Mutex.RUnlock()
	gateway.Mutex.RUnlock()
}
