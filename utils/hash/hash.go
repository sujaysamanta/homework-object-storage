package hash

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/spacelift-io/homework-object-storage/models"
)

// GetMinioInstanceForID returns the Minio instance to use for a given object ID
func GetMinioInstanceForID(g *models.ObjectStorageGateway, objectID string) (*models.MinioInstance, error) {
	g.Mutex.RLock()
	defer g.Mutex.RUnlock()

	if len(g.MinioInstances) == 0 {
		return nil, fmt.Errorf("no Minio instances available")
	}

	// Use SHA-256 for better distribution across instances
	hash := sha256.Sum256([]byte(objectID))

	// Use all 32 bytes of the hash for better distribution
	// We'll use 4 uint64 values (8 bytes each) and XOR them together
	hashPart1 := binary.BigEndian.Uint64(hash[:8])
	hashPart2 := binary.BigEndian.Uint64(hash[8:16])
	hashPart3 := binary.BigEndian.Uint64(hash[16:24])
	hashPart4 := binary.BigEndian.Uint64(hash[24:32])

	// XOR the parts together to get a well-distributed uint64 value
	hashInt := hashPart1 ^ hashPart2 ^ hashPart3 ^ hashPart4

	instanceIndex := int(hashInt % uint64(len(g.MinioInstances)))

	return &g.MinioInstances[instanceIndex], nil
}
