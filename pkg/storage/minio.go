package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"gin-rocket/pkg/configx"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

var ErrInvalidObjectKey = errors.New("invalid object key")

type FileInfo struct {
	Bucket       string `json:"bucket"`
	ObjectKey    string `json:"object_key"`
	Category     string `json:"category"`
	OriginalName string `json:"original_name"`
	ContentType  string `json:"content_type"`
	Size         int64  `json:"size"`
	URL          string `json:"url,omitempty"`
}

// MinIOStore wraps the MinIO client and centralizes object-key generation and URL logic.
type MinIOStore struct {
	client *minio.Client
	cfg    configx.MinIOConfig
}

// NewMinIO initializes a MinIO client and creates the bucket if configured.
func NewMinIO(ctx context.Context, cfg configx.MinIOConfig, log *zap.Logger) (*MinIOStore, error) {
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

	return &MinIOStore{
		client: client,
		cfg:    cfg,
	}, nil
}

// Upload stores a file under a business category like avatar/ or docs/.
func (s *MinIOStore) Upload(ctx context.Context, category, originalName string, reader io.Reader, size int64, contentType string) (*FileInfo, error) {
	normalizedCategory := normalizeCategory(category)
	objectKey := buildObjectKey(normalizedCategory, originalName)
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(originalName))
	}

	info, err := s.client.PutObject(ctx, s.cfg.Bucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("put object: %w", err)
	}

	return &FileInfo{
		Bucket:       s.cfg.Bucket,
		ObjectKey:    objectKey,
		Category:     normalizedCategory,
		OriginalName: originalName,
		ContentType:  contentType,
		Size:         info.Size,
		URL:          s.objectURL(objectKey),
	}, nil
}

// Delete removes an object by its full object key.
func (s *MinIOStore) Delete(ctx context.Context, objectKey string) error {
	objectKey = strings.TrimSpace(strings.TrimPrefix(objectKey, "/"))
	if objectKey == "" {
		return ErrInvalidObjectKey
	}

	return s.client.RemoveObject(ctx, s.cfg.Bucket, objectKey, minio.RemoveObjectOptions{})
}

// PresignGetURL generates a temporary URL for file access.
func (s *MinIOStore) PresignGetURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	objectKey = strings.TrimSpace(strings.TrimPrefix(objectKey, "/"))
	if objectKey == "" {
		return "", ErrInvalidObjectKey
	}

	presignedURL, err := s.client.PresignedGetObject(ctx, s.cfg.Bucket, objectKey, expiry, url.Values{})
	if err != nil {
		return "", fmt.Errorf("presign get object: %w", err)
	}

	return presignedURL.String(), nil
}

func (s *MinIOStore) objectURL(objectKey string) string {
	if s.cfg.BaseURL == "" {
		return ""
	}

	return strings.TrimRight(s.cfg.BaseURL, "/") + "/" + strings.TrimLeft(objectKey, "/")
}

func buildObjectKey(category, originalName string) string {
	ext := strings.ToLower(filepath.Ext(originalName))
	datePrefix := time.Now().UTC().Format("2006/01/02")
	fileName := uuid.NewString() + ext
	return path.Join(category, datePrefix, fileName)
}

func normalizeCategory(category string) string {
	category = strings.TrimSpace(category)
	if category == "" {
		return "misc"
	}

	cleaned := path.Clean("/" + strings.ReplaceAll(category, "\\", "/"))
	cleaned = strings.Trim(cleaned, "/")
	if cleaned == "" || cleaned == "." {
		return "misc"
	}

	return cleaned
}
