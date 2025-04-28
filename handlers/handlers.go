package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/spacelift-io/homework-object-storage/interfaces"
	"github.com/spacelift-io/homework-object-storage/models"
	"github.com/spacelift-io/homework-object-storage/utils"
	"github.com/spacelift-io/homework-object-storage/utils/hash"
	minioUtil "github.com/spacelift-io/homework-object-storage/utils/minio"
)

// HandlePutObject handles PUT /object/{id} requests
func HandlePutObject(g *models.ObjectStorageGateway) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		objectID := vars["id"]

		// Validate object ID
		if !utils.IsValidObjectID(objectID) {
			http.Error(w, "Invalid object ID", http.StatusBadRequest)
			return
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusInternalServerError)
			return
		}

		// Check if the request body is empty
		if len(body) == 0 {
			http.Error(w, "Request body is empty", http.StatusBadRequest)
			return
		}

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Printf("Failed to close request body: %v", err)
			}
		}(r.Body)

		// Get Minio instance for this object ID
		instance, err := hash.GetMinioInstanceForID(g, objectID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get Minio instance: %v", err), http.StatusInternalServerError)
			return
		}

		// Upload an object to Minio
		err = minioUtil.PutObject(
			r.Context(),
			instance,
			objectID,
			strings.NewReader(string(body)),
			int64(len(body)),
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to put object: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		log.Printf("Successfully stored object %s in instance %s", objectID, instance.Name)
	}
}

// validateObjectID validates the object ID and returns an error response if invalid
func validateObjectID(w http.ResponseWriter, objectID string) bool {
	if !utils.IsValidObjectID(objectID) {
		http.Error(w, "Invalid object ID", http.StatusBadRequest)
		return false
	}
	return true
}

// getMinioInstance gets the Minio instance for the given object ID
func getMinioInstance(w http.ResponseWriter, g *models.ObjectStorageGateway, objectID string) (*models.MinioInstance, bool) {
	instance, err := hash.GetMinioInstanceForID(g, objectID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get Minio instance: %v", err), http.StatusInternalServerError)
		return nil, false
	}
	return instance, true
}

// HandleGetObject handles GET /object/{id} requests
func HandleGetObject(g *models.ObjectStorageGateway) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		objectID := vars["id"]

		// Validate object ID
		if !validateObjectID(w, objectID) {
			return
		}

		// Get Minio instance for this object ID
		instance, ok := getMinioInstance(w, g, objectID)
		if !ok {
			return
		}

		// Get an object from Minio and check if it exists
		object, stat, ok := minioUtil.GetObjectFromMinio(w, r.Context(), instance, objectID)
		if !ok {
			return
		}
		defer func(object interfaces.MinioObjectInterface) {
			err := object.Close()
			if err != nil {
				log.Printf("Failed to close object: %v", err)
			}
		}(object)

		// Write object data to response
		minioUtil.WriteObjectToResponse(w, object, stat, objectID, instance.Name)
	}
}
