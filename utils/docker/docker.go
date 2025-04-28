package docker

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	"github.com/spacelift-io/homework-object-storage/models"
	"github.com/spacelift-io/homework-object-storage/utils/minio"
)

// NewDockerClient creates a new Docker client
func NewDockerClient() (*client.Client, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	return dockerClient, nil
}

// extractIPAddress extracts the IP address from a container
func extractIPAddress(containerJSON container.InspectResponse) (string, error) {
	for networkName, network := range containerJSON.NetworkSettings.Networks {
		if strings.Contains(networkName, "amazin-object-storage") {
			return network.IPAddress, nil
		}
	}
	return "", fmt.Errorf("could not find IP address")
}

// extractCredentials extracts access and secret keys from container environment variables
func extractCredentials(containerJSON container.InspectResponse) (string, string, error) {
	var accessKey, secretKey string
	for _, env := range containerJSON.Config.Env {
		if strings.HasPrefix(env, "MINIO_ACCESS_KEY=") {
			accessKey = strings.TrimPrefix(env, "MINIO_ACCESS_KEY=")
		} else if strings.HasPrefix(env, "MINIO_SECRET_KEY=") {
			secretKey = strings.TrimPrefix(env, "MINIO_SECRET_KEY=")
		}
	}

	if accessKey == "" || secretKey == "" {
		return "", "", fmt.Errorf("missing access or secret key")
	}
	return accessKey, secretKey, nil
}

// processContainer processes a single container and returns a MinioInstance if valid
func processContainer(ctx context.Context, dockerClient *client.Client, container container.Summary) (*models.MinioInstance, error) {
	// Check if this is a Minio container
	if !strings.Contains(container.Names[0], "amazin-object-storage-node") {
		return nil, fmt.Errorf("not a Minio container")
	}

	// Get container details
	containerJSON, err := dockerClient.ContainerInspect(ctx, container.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	// Extract IP address
	ipAddress, err := extractIPAddress(containerJSON)
	if err != nil {
		return nil, fmt.Errorf("IP address extraction failed: %w", err)
	}

	// Extract credentials
	accessKey, secretKey, err := extractCredentials(containerJSON)
	if err != nil {
		return nil, fmt.Errorf("credential extraction failed: %w", err)
	}

	// Create a unique bucket name for this instance
	containerName := strings.TrimPrefix(container.Names[0], "/")
	bucketName := fmt.Sprintf("bucket-%s", containerName)

	// Setup Minio client and bucket
	minioClient, err := minio.SetupMinioClient(ctx, ipAddress, accessKey, secretKey, bucketName)
	if err != nil {
		return nil, fmt.Errorf("minio setup failed: %w", err)
	}

	// Create and return the instance
	return &models.MinioInstance{
		ID:         container.ID,
		Name:       containerName,
		IPAddress:  ipAddress,
		AccessKey:  accessKey,
		SecretKey:  secretKey,
		Client:     minioClient,
		BucketName: bucketName,
	}, nil
}

// DiscoverMinioInstances finds all Minio instances from Docker
func DiscoverMinioInstances(g *models.ObjectStorageGateway) error {
	ctx := context.Background()

	// List containers
	containers, err := g.DockerClient.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	var newInstances []models.MinioInstance

	// Process each docker container
	for _, dockerContainer := range containers {
		instance, err := processContainer(ctx, g.DockerClient, dockerContainer)
		if err != nil {
			log.Printf("Skipping dockerContainer %s: %v", dockerContainer.ID, err)
			continue
		}
		newInstances = append(newInstances, *instance)
	}

	if len(newInstances) == 0 {
		return fmt.Errorf("no Minio instances found")
	}

	// Sort instances by name for consistent ordering
	sort.Slice(newInstances, func(i, j int) bool {
		return newInstances[i].Name < newInstances[j].Name
	})

	// Update instances with lock
	g.Mutex.Lock()
	g.MinioInstances = newInstances
	g.Mutex.Unlock()

	log.Printf("Discovered %d Minio instances", len(newInstances))
	return nil
}

// RefreshMinioInstances periodically refreshes the list of Minio instances
func RefreshMinioInstances(g *models.ObjectStorageGateway) {
	for range g.RefreshTicker.C {
		if err := DiscoverMinioInstances(g); err != nil {
			log.Printf("Failed to refresh Minio instances: %v", err)
		}
	}
}
