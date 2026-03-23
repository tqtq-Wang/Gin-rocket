package storage

import (
	"context"
	"fmt"

	"gin-rocket/pkg/configx"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

// NewMinIO initializes a MinIO client and creates the bucket if configured.
func NewMinIO(ctx context.Context, cfg configx.MinIOConfig, log *zap.Logger) (*minio.Client, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("check minio bucket: %w", err)
	}

	if !exists && cfg.AutoCreateBucket {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{
			Region: cfg.Location,
		}); err != nil {
			return nil, fmt.Errorf("create minio bucket: %w", err)
		}

		log.Info("minio bucket created", zap.String("bucket", cfg.Bucket))
	}

	if !exists && !cfg.AutoCreateBucket {
		return nil, fmt.Errorf("minio bucket %s does not exist", cfg.Bucket)
	}

	return client, nil
}
