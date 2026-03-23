package repository

import (
	"context"
	"time"

	"gin-rocket/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserRepository 封装用户及其 RBAC 关联的持久化操作。
type UserRepository interface {
	GetByID(ctx context.Context, id uint64) (*model.User, error)
	GetProfileByID(ctx context.Context, id uint64) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	Register(ctx context.Context, user *model.User) (bool, error)
	UpdateLastLoginAt(ctx context.Context, id uint64, lastLoginAt time.Time) error
	ListPermissionCodesByUserID(ctx context.Context, userID uint64) ([]string, error)
	ListMenusByUserID(ctx context.Context, userID uint64) ([]model.Permission, error)
}

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).
		Preload("Roles", "status = ?", model.StatusEnabled).
		First(&user, id).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *GormUserRepository) GetProfileByID(ctx context.Context, id uint64) (*model.User, error) {
	return r.GetByID(ctx, id)
}

func (r *GormUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).
		Preload("Roles", "status = ?", model.StatusEnabled).
		Where("username = ? AND status = ?", username, model.StatusEnabled).
		First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// Register 在一个事务里完成用户创建和首用户超管绑定，避免并发下出现多个首用户。
func (r *GormUserRepository) Register(ctx context.Context, user *model.User) (bool, error) {
	var isSuperAdmin bool

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.User{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Count(&count).Error; err != nil {
			return err
		}

		if err := tx.Create(user).Error; err != nil {
			return err
		}

		if count > 0 {
			return nil
		}

		role := &model.Role{}
		if err := tx.Where("code = ? AND status = ?", model.RoleCodeSuperAdmin, model.StatusEnabled).
			First(role).Error; err != nil {
			return err
		}

		if err := tx.Model(user).Association("Roles").Append(role); err != nil {
			return err
		}

		isSuperAdmin = true
		return nil
	})
	if err != nil {
		if isDuplicateKeyError(err) {
			return false, ErrUserAlreadyExists
		}
		return false, err
	}

	if isSuperAdmin {
		return true, r.GetRolePreloadedUser(ctx, user.ID, user)
	}

	return false, r.GetRolePreloadedUser(ctx, user.ID, user)
}

func (r *GormUserRepository) UpdateLastLoginAt(ctx context.Context, id uint64, lastLoginAt time.Time) error {
	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", id).
		Update("last_login_at", lastLoginAt).Error
}

func (r *GormUserRepository) ListPermissionCodesByUserID(ctx context.Context, userID uint64) ([]string, error) {
	codes := make([]string, 0)
	err := r.db.WithContext(ctx).
		Table("permissions").
		Distinct("permissions.code").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN roles ON roles.id = role_permissions.role_id").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Where("roles.status = ?", model.StatusEnabled).
		Where("permissions.status = ?", model.StatusEnabled).
		Pluck("permissions.code", &codes).Error
	if err != nil {
		return nil, err
	}

	return codes, nil
}

func (r *GormUserRepository) ListMenusByUserID(ctx context.Context, userID uint64) ([]model.Permission, error) {
	menus := make([]model.Permission, 0)
	err := r.db.WithContext(ctx).
		Model(&model.Permission{}).
		Select("permissions.*").
		Distinct().
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN roles ON roles.id = role_permissions.role_id").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Where("roles.status = ?", model.StatusEnabled).
		Where("permissions.status = ?", model.StatusEnabled).
		Where("permissions.type = ?", model.PermissionTypeMenu).
		Order("permissions.sort ASC, permissions.id ASC").
		Find(&menus).Error
	if err != nil {
		return nil, err
	}

	return menus, nil
}

func (r *GormUserRepository) GetRolePreloadedUser(ctx context.Context, userID uint64, target *model.User) error {
	if target == nil {
		return gorm.ErrInvalidData
	}

	return r.db.WithContext(ctx).
		Preload("Roles", "status = ?", model.StatusEnabled).
		First(target, userID).Error
}
