package arc

import (
	"context"

	"github.com/docker/distribution/registry/storage/driver"
	"github.com/docker/distribution/registry/storage/driver/s3-aws"
)

const (
	root     = "/docker/registry/v2"
	repoRoot = root + "/repositories"
)

var (
	metadataComponents = []string{"_manifests", "_layers", "_uploads"}
)

// RegistryClient interacts with the docker registry storage
type RegistryClient struct {
	storage driver.StorageDriver
}

// RegistryRepo represents an image repository inside the registry
type RegistryRepo string

// NewRegistryClientS3 returns a new RegistryClient configured for a S3 storage
// backend
func NewRegistryClientS3(accessKey, secretKey, endpoint, bucket string) (*RegistryClient, error) {
	s, err := s3.New(s3.DriverParameters{
		AccessKey:      accessKey,
		SecretKey:      secretKey,
		Bucket:         bucket,
		Region:         "us-east-1",
		RegionEndpoint: endpoint,
	})
	if err != nil {
		return nil, err
	}

	return &RegistryClient{storage: s}, nil
}

// DeleteRepo deletes the given repository metadata, but NOT its associated
/// BLOBs
func (r RegistryClient) DeleteRepo(name string) error {
	ctx := context.Background()

	for _, part := range metadataComponents {
		path := repoRoot + "/" + name + "/" + part
		if err := r.storage.Delete(ctx, path); err != nil {
			if _, ok := err.(driver.PathNotFoundError); ok {
				continue
			}
			return err
		}
	}

	return nil
}
