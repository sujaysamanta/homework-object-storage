package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"github.com/spacelift-io/homework-object-storage/interfaces"
	"github.com/spacelift-io/homework-object-storage/interfaces/mocks"
	"github.com/spacelift-io/homework-object-storage/models"
)

// createTestGateway creates a test gateway with mock Minio instances
func createTestGateway() *models.ObjectStorageGateway {
	// Create mock Minio instances
	instances := []models.MinioInstance{
		{
			ID:         "instance1",
			Name:       "instance1",
			IPAddress:  "127.0.0.1",
			AccessKey:  "test-access",
			SecretKey:  "test-secret",
			Client:     mocks.NewMockMinioClient(),
			BucketName: "test-bucket",
		},
	}

	// Create a gateway with the mock instances
	return &models.ObjectStorageGateway{
		MinioInstances: instances,
		Mutex:          sync.RWMutex{},
	}
}

// Test the getMinioInstance function
func TestGetMinioInstance(t *testing.T) {
	// Create a test gateway
	gateway := createTestGateway()

	// Test cases
	tests := []struct {
		name       string
		objectID   string
		wantResult bool
	}{
		{"Valid object ID", "test123", true},
		{"Empty gateway instances", "test123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			// For the "Empty gateway instances" test, use an empty gateway
			testGateway := gateway
			if tt.name == "Empty gateway instances" {
				testGateway = &models.ObjectStorageGateway{
					MinioInstances: []models.MinioInstance{},
					Mutex:          sync.RWMutex{},
				}
			}

			instance, result := getMinioInstance(w, testGateway, tt.objectID)

			if result != tt.wantResult {
				t.Errorf("getMinioInstance() result = %v, want %v", result, tt.wantResult)
			}

			if tt.wantResult && instance == nil {
				t.Errorf("getMinioInstance() returned nil instance when result is true")
			}

			if !tt.wantResult && w.Code != http.StatusInternalServerError {
				t.Errorf("Expected status %d for error, got %d", http.StatusInternalServerError, w.Code)
			}
		})
	}
}

// Test the HandlePutObject function
func TestHandlePutObject(t *testing.T) {
	// Create a test gateway
	gateway := createTestGateway()

	// Create a handler function
	handler := HandlePutObject(gateway)

	// Test cases
	tests := []struct {
		name       string
		objectID   string
		body       string
		wantStatus int
	}{
		{"Valid request", "test123", "test data", http.StatusOK},
		{"Invalid object ID", "test-123", "test data", http.StatusBadRequest},
		{"Empty body", "test123", "", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request
			req, err := http.NewRequest("PUT", "/object/"+tt.objectID, strings.NewReader(tt.body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Set up the gorilla/mux vars
			req = mux.SetURLVars(req, map[string]string{"id": tt.objectID})

			// Create a response recorder
			w := httptest.NewRecorder()

			// Call the handler
			handler(w, req)

			// Check the status code
			if w.Code != tt.wantStatus {
				t.Errorf("HandlePutObject() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

// Test the HandleGetObject function
func TestHandleGetObject(t *testing.T) {
	// Create a test gateway
	gateway := createTestGateway()

	// Store test data in the mock client
	mockClient := gateway.MinioInstances[0].Client.(*mocks.MockMinioClient)
	mockClient.Objects["test-bucket/test123"] = []byte("test data")

	// Create a test gateway with a failing GetObject function
	failingGateway := createTestGateway()
	failingClient := failingGateway.MinioInstances[0].Client.(*mocks.MockMinioClient)
	failingClient.GetObjectFunc = func(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (interfaces.MinioObjectInterface, error) {
		return nil, errors.New("GetObject failed")
	}

	// Create a handler function
	handler := HandleGetObject(gateway)

	// Create a handler function for the failing gateway
	failingHandler := HandleGetObject(failingGateway)

	// Create a test gateway with a failing Stat function
	statFailGateway := createTestGateway()
	statFailClient := statFailGateway.MinioInstances[0].Client.(*mocks.MockMinioClient)
	statFailClient.GetObjectFunc = func(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (interfaces.MinioObjectInterface, error) {
		mockObj := &mocks.MockMinioObject{
			StatFunc: func() (minio.ObjectInfo, error) {
				return minio.ObjectInfo{}, errors.New("Stat failed")
			},
			ReadFunc: func(p []byte) (n int, err error) {
				return 0, io.EOF
			},
			CloseFunc: func() error {
				return nil
			},
		}
		return mockObj, nil
	}

	// Create a handler function for the stat-failing gateway
	statFailHandler := HandleGetObject(statFailGateway)

	// Create a test gateway with a "not found" error from Stat
	notFoundGateway := createTestGateway()
	notFoundClient := notFoundGateway.MinioInstances[0].Client.(*mocks.MockMinioClient)
	notFoundClient.GetObjectFunc = func(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (interfaces.MinioObjectInterface, error) {
		mockObj := &mocks.MockMinioObject{
			StatFunc: func() (minio.ObjectInfo, error) {
				return minio.ObjectInfo{}, minio.ErrorResponse{
					Code:    "NoSuchKey",
					Message: "The specified key does not exist.",
				}
			},
			ReadFunc: func(p []byte) (n int, err error) {
				return 0, io.EOF
			},
			CloseFunc: func() error {
				return nil
			},
		}
		return mockObj, nil
	}

	// Create a handler function for the not-found gateway
	notFoundHandler := HandleGetObject(notFoundGateway)

	// Test cases
	tests := []struct {
		name       string
		objectID   string
		handler    http.HandlerFunc
		wantStatus int
		wantBody   string
	}{
		{"Valid request", "test123", handler, http.StatusOK, "test data"},
		{"Invalid object ID", "test-123", handler, http.StatusBadRequest, ""},
		{"Object not found", "nonexistent", notFoundHandler, http.StatusNotFound, ""},
		{"GetObject failure", "test123", failingHandler, http.StatusInternalServerError, ""},
		{"Stat failure", "test123", statFailHandler, http.StatusInternalServerError, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request
			req, err := http.NewRequest("GET", "/object/"+tt.objectID, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Set up the gorilla/mux vars
			req = mux.SetURLVars(req, map[string]string{"id": tt.objectID})

			// Create a response recorder
			w := httptest.NewRecorder()

			// Call the handler from the test case
			tt.handler(w, req)

			// Check the status code
			if w.Code != tt.wantStatus {
				t.Errorf("HandleGetObject() status = %v, want %v", w.Code, tt.wantStatus)
			}

			// For successful requests, check the response body
			if tt.wantStatus == http.StatusOK {
				if w.Body.String() != tt.wantBody {
					t.Errorf("HandleGetObject() body = %v, want %v", w.Body.String(), tt.wantBody)
				}
			}
		})
	}

	// Test with a request body that uses bytes.Buffer
	t.Run("Test with bytes.Buffer", func(t *testing.T) {
		// Create a request with a bytes.Buffer body
		body := bytes.NewBufferString("test data from buffer")
		req, err := http.NewRequest("PUT", "/object/test123", body)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Set up the gorilla/mux vars
		req = mux.SetURLVars(req, map[string]string{"id": "test123"})

		// Create a response recorder
		w := httptest.NewRecorder()

		// Call the handler
		putHandler := HandlePutObject(gateway)
		putHandler(w, req)

		// Check the status code
		if w.Code != http.StatusOK {
			t.Errorf("HandlePutObject() status = %v, want %v", w.Code, http.StatusOK)
		}
	})
}

// Test the validateObjectID function
func TestValidateObjectID(t *testing.T) {
	tests := []struct {
		name       string
		objectID   string
		wantResult bool
	}{
		{"Valid ID", "test123", true},
		{"Invalid ID", "test-123", false},
		{"Empty ID", "", false},
		{"ID too long", "12345678901234567890123456789012345", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			result := validateObjectID(w, tt.objectID)
			if result != tt.wantResult {
				t.Errorf("validateObjectID() = %v, want %v", result, tt.wantResult)
			}
			if !tt.wantResult && w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d for invalid ID, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}
