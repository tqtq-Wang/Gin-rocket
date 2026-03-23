package repository

import (
	"context"

	"gin-rocket/internal/model"

	"gorm.io/gorm"
)

type PermissionRepository interface {
	GetAPIByMethodPath(ctx context.Context, method, path string) (*model.Permission, error)
}

type GormPermissionRepository struct {
	db *gorm.DB
}

func NewGormPermissionRepository(db *gorm.DB) *GormPermissionRepository {
	return &GormPermissionRepository{db: db}
}

func (r *GormPermissionRepository) GetAPIByMethodPath(ctx context.Context, method, path string) (*model.Permission, error) {
	var permission model.Permission
	if err := r.db.WithContext(ctx).
		Where("type = ? AND method = ? AND path = ? AND status = ?", model.PermissionTypeAPI, method, path, model.StatusEnabled).
		First(&permission).Error; err != nil {
		return nil, err
	}

	return &permission, nil
}
