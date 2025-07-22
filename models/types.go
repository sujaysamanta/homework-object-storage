package models

import (
	"sync"
	"time"

	"github.com/docker/docker/client"
	"github.com/spacelift-io/homework-object-storage/interfaces"
)

// MinioInstance represents a single instance of a Minio server used for object storage operations.
type MinioInstance struct {
	ID         string                          // ID uniquely identifies the Minio server instance.
	Name       string                          // Name represents the name of the Minio server instance.
	IPAddress  string                          // IPAddress represents the IP address of the Minio server instance.
	AccessKey  string                          // AccessKey represents the access key for the Minio server instance.
	SecretKey  string                          // SecretKey represents the secret key required to authenticate with the Minio server.
	Client     interfaces.MinioClientInterface // Client represents the Minio client used to communicate with the Minio server instance.
	BucketName string                          // BucketName represents the name of the bucket associated with the Minio instance.
}

// ObjectStorageGateway represents a gateway for managing multiple Minio instances and facilitating object storage operations.
type ObjectStorageGateway struct {
	MinioInstances []MinioInstance // MinioInstances holds a list of Minio server instances managed by the ObjectStorageGateway.
	Mutex          sync.RWMutex    // Mutex is used to synchronize read/write access to the MinioInstances slice in the ObjectStorageGateway type.
	DockerClient   *client.Client  // DockerClient represents the Docker client used for managing containerized Minio instances.
	RefreshTicker  *time.Ticker    // RefreshTicker triggers periodic refresh operations for Minio instances in the ObjectStorageGateway.
}
