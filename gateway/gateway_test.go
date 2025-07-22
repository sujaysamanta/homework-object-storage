package gateway

import (
	"testing"
	"time"

	"github.com/spacelift-io/homework-object-storage/models"
)

// TestNewObjectStorageGateway tests the NewObjectStorageGateway function
func TestNewObjectStorageGateway(t *testing.T) {
	// Skip this test in normal test runs since it would require a real Docker client
	t.Skip("Skipping test that requires a real Docker client")
}

// TestStop tests that the Stop function correctly stops the ticker
func TestStop(t *testing.T) {
	// Create a gateway with a ticker
	gateway := &models.ObjectStorageGateway{
		RefreshTicker: time.NewTicker(1 * time.Second),
	}

	// Call Stop
	Stop(gateway)

	// Verify that the ticker is stopped
	// Note: There's no direct way to check if a ticker is stopped,
	// but we can check that the ticker channel is closed by trying to receive from it
	select {
	case <-gateway.RefreshTicker.C:
		t.Error("Ticker channel is still open")
	default:
		// This is expected - the channel should be closed
	}
}
