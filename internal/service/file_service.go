package service

import (
	"context"
	"errors"
	"mime"
	"mime/multipart"
	"path/filepath"
	"time"

	"gin-rocket/pkg/storage"
)

var ErrFileStorageDisabled = errors.New("file storage disabled")

type FileService interface {
	Upload(ctx context.Context, category string, fileHeader *multipart.FileHeader) (*storage.FileInfo, error)
	UploadMultiple(ctx context.Context, category string, fileHeaders []*multipart.FileHeader) ([]*storage.FileInfo, error)
	Delete(ctx context.Context, objectKey string) error
	GetPresignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error)
}

type DefaultFileService struct {
	store         *storage.MinIOStore
	presignExpiry time.Duration
}

func NewFileService(store *storage.MinIOStore, presignExpiry time.Duration) *DefaultFileService {
	return &DefaultFileService{
		store:         store,
		presignExpiry: presignExpiry,
	}
}

func (s *DefaultFileService) Upload(ctx context.Context, category string, fileHeader *multipart.FileHeader) (*storage.FileInfo, error) {
	if s.store == nil {
		return nil, ErrFileStorageDisabled
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(fileHeader.Filename))
	}

	return s.store.Upload(ctx, category, fileHeader.Filename, file, fileHeader.Size, contentType)
}

func (s *DefaultFileService) UploadMultiple(ctx context.Context, category string, fileHeaders []*multipart.FileHeader) ([]*storage.FileInfo, error) {
	if s.store == nil {
		return nil, ErrFileStorageDisabled
	}

	result := make([]*storage.FileInfo, 0, len(fileHeaders))
	for _, fileHeader := range fileHeaders {
		info, err := s.Upload(ctx, category, fileHeader)
		if err != nil {
			return nil, err
		}
		result = append(result, info)
	}

	return result, nil
}

func (s *DefaultFileService) Delete(ctx context.Context, objectKey string) error {
	if s.store == nil {
		return ErrFileStorageDisabled
	}

	return s.store.Delete(ctx, objectKey)
}

func (s *DefaultFileService) GetPresignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	if s.store == nil {
		return "", ErrFileStorageDisabled
	}

	if expiry <= 0 {
		expiry = s.presignExpiry
	}

	return s.store.PresignGetURL(ctx, objectKey, expiry)
}
