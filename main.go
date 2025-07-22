package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/spacelift-io/homework-object-storage/gateway"
	"github.com/spacelift-io/homework-object-storage/handlers"
	"github.com/spacelift-io/homework-object-storage/server"
)

func main() {
	log.Println("Starting Amazin Object Storage Gateway")

	// Create a new gateway instance
	gatewayInstance, err := gateway.NewObjectStorageGateway()
	if err != nil {
		log.Fatalf("Failed to create gateway: %v", err)
	}
	defer gateway.Stop(gatewayInstance)

	// Print discovered Minio instances
	gatewayInstance.Mutex.RLock()
	for _, instance := range gatewayInstance.MinioInstances {
		log.Printf("Discovered Minio instance: %s at %s", instance.Name, instance.IPAddress)
	}
	gatewayInstance.Mutex.RUnlock()

	// Create HTTP handlers
	putObjectHandler := handlers.HandlePutObject(gatewayInstance)
	getObjectHandler := handlers.HandleGetObject(gatewayInstance)

	// Set up routes
	router := server.SetupRoutes(putObjectHandler, getObjectHandler)

	// Start the gateway server
	if err := server.Start(router); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Failed to start gateway: %v", err)
	}
}
