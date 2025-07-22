package hash

import (
	"testing"

	"github.com/spacelift-io/homework-object-storage/models"
)

func TestGetMinioInstanceForID(t *testing.T) {
	// Create test instances
	testInstances := []models.MinioInstance{
		{ID: "instance1", Name: "instance1"},
		{ID: "instance2", Name: "instance2"},
		{ID: "instance3", Name: "instance3"},
	}

	// Create a gateway with test instances
	gateway := &models.ObjectStorageGateway{
		MinioInstances: testInstances,
	}

	// Test with various object IDs
	tests := []struct {
		name      string
		objectID  string
		wantError bool
	}{
		{"Valid object ID 1", "test1", false},
		{"Valid object ID 2", "test2", false},
		{"Valid object ID 3", "abcdef123456", false},
		{"Empty object ID", "", false}, // This is valid for the hash function, validation happens elsewhere
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance, err := GetMinioInstanceForID(gateway, tt.objectID)

			if tt.wantError {
				if err == nil {
					t.Errorf("GetMinioInstanceForID() error = nil, wantError = true")
				}
				return
			}

			if err != nil {
				t.Errorf("GetMinioInstanceForID() error = %v, wantError = false", err)
				return
			}

			if instance == nil {
				t.Errorf("GetMinioInstanceForID() returned nil instance")
				return
			}

			// Verify the instance is one of our test instances
			found := false
			for _, testInstance := range testInstances {
				if instance.ID == testInstance.ID {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("GetMinioInstanceForID() returned instance with ID %s, which is not in our test instances", instance.ID)
			}
		})
	}

	// Test with empty instances
	t.Run("No instances available", func(t *testing.T) {
		emptyGateway := &models.ObjectStorageGateway{
			MinioInstances: []models.MinioInstance{},
		}

		_, err := GetMinioInstanceForID(emptyGateway, "test")
		if err == nil {
			t.Errorf("GetMinioInstanceForID() with empty instances, expected error but got nil")
		}
	})

	// Test consistency - same object ID should always map to same instance
	t.Run("Consistency check", func(t *testing.T) {
		objectID := "consistent-test-id"

		instance1, err := GetMinioInstanceForID(gateway, objectID)
		if err != nil {
			t.Errorf("GetMinioInstanceForID() error = %v", err)
			return
		}

		instance2, err := GetMinioInstanceForID(gateway, objectID)
		if err != nil {
			t.Errorf("GetMinioInstanceForID() error = %v", err)
			return
		}

		if instance1.ID != instance2.ID {
			t.Errorf("GetMinioInstanceForID() not consistent: got %s and %s for same object ID",
				instance1.ID, instance2.ID)
		}
	})
}
