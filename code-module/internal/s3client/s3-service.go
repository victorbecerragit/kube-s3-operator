// S3 API operations.

package s3client

import (
    "context"
)

// BucketManager defines bucket lifecycle operations.
type BucketManager interface {
    EnsureBucket(ctx context.Context, name string) error
    DeleteBucket(ctx context.Context, name string) error
    BucketExists(ctx context.Context, name string) (bool, error)
}

// NewService returns the real implementation.
func NewService(cfg Config) BucketManager {
    // initialize AWS SDK client with cfg
}

type service struct {
    // aws client, logger, etc.
}

func (s *service) EnsureBucket(ctx context.Context, name string) error {
    // Create bucket if not exists
}

func (s *service) DeleteBucket(ctx context.Context, name string) error {
    // Delete bucket logic
}

func (s *service) BucketExists(ctx context.Context, name string) (bool, error) {
    // HeadBucket call
}
