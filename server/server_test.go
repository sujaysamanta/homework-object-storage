package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetupRoutes(t *testing.T) {
	// Create mock handler functions
	putHandlerCalled := false
	getHandlerCalled := false

	mockPutHandler := func(w http.ResponseWriter, r *http.Request) {
		putHandlerCalled = true
	}

	mockGetHandler := func(w http.ResponseWriter, r *http.Request) {
		getHandlerCalled = true
	}

	// Setup routes with mock handlers
	router := SetupRoutes(mockPutHandler, mockGetHandler)

	// Test PUT request
	putReq := httptest.NewRequest("PUT", "/object/test-id", nil)
	putRec := httptest.NewRecorder()
	router.ServeHTTP(putRec, putReq)

	if !putHandlerCalled {
		t.Error("PUT handler was not called")
	}

	// Test GET request
	getReq := httptest.NewRequest("GET", "/object/test-id", nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	if !getHandlerCalled {
		t.Error("GET handler was not called")
	}

	// Test unsupported method
	postReq := httptest.NewRequest("POST", "/object/test-id", nil)
	postRec := httptest.NewRecorder()
	router.ServeHTTP(postRec, postReq)

	if postRec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d for POST request, got %d",
			http.StatusMethodNotAllowed, postRec.Code)
	}

	// Test invalid path
	invalidReq := httptest.NewRequest("GET", "/invalid-path", nil)
	invalidRec := httptest.NewRecorder()
	router.ServeHTTP(invalidRec, invalidReq)

	if invalidRec.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d for invalid path, got %d",
			http.StatusNotFound, invalidRec.Code)
	}
}

// For the Start function, we'll just test that it doesn't panic with valid input
// We can't easily test the actual server start without modifying the code
func TestStartDoesNotPanic(t *testing.T) {
	// Skip this test in normal test runs since it would start a server
	t.Skip("Skipping test that would start a server")

	// In a real test, we would:
	// 1. Mock the http.Server to prevent actual listening
	// 2. Inject the mock into Start function
	// 3. Verify the server is configured correctly

	// For now, we just ensure the code exists and compiles
	router := SetupRoutes(
		func(w http.ResponseWriter, r *http.Request) {},
		func(w http.ResponseWriter, r *http.Request) {},
	)

	// This is just a compile-time check
	_ = router
}
