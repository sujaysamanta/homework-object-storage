package gateway

import (
	"time"

	"github.com/spacelift-io/homework-object-storage/models"
	"github.com/spacelift-io/homework-object-storage/utils/docker"
)

// NewObjectStorageGateway creates a new gateway instance
func NewObjectStorageGateway() (*models.ObjectStorageGateway, error) {
	// Create a Docker client
	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}

	gateway := &models.ObjectStorageGateway{
		DockerClient:  dockerClient,
		RefreshTicker: time.NewTicker(30 * time.Second),
	}

	// Initial discovery of Minio instances
	if err := docker.DiscoverMinioInstances(gateway); err != nil {
		return nil, err
	}

	// Start background refresh of Minio instances
	go docker.RefreshMinioInstances(gateway)

	return gateway, nil
}

// Stop stops the gateway server
func Stop(g *models.ObjectStorageGateway) {
	g.RefreshTicker.Stop()
}
