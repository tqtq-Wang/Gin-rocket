package service

import (
	"context"
	"errors"
	"time"

	"gin-rocket/internal/model"
	"gin-rocket/internal/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MenuItem struct {
	ID        uint64      `json:"id"`
	Name      string      `json:"name"`
	Code      string      `json:"code"`
	Path      string      `json:"path"`
	RouteName string      `json:"route_name,omitempty"`
	Component string      `json:"component,omitempty"`
	Redirect  string      `json:"redirect,omitempty"`
	Icon      string      `json:"icon,omitempty"`
	Sort      int         `json:"sort"`
	Hidden    bool        `json:"hidden"`
	Children  []*MenuItem `json:"children,omitempty"`
}

type PermissionService interface {
	CheckAccess(ctx context.Context, userID uint64, roleCodes []string, method, path string) (bool, *model.Permission, error)
	GetCurrentUserMenus(ctx context.Context, userID uint64) ([]*MenuItem, error)
	GetCurrentUserPermissionCodes(ctx context.Context, userID uint64) ([]string, error)
}

type DefaultPermissionService struct {
	userRepo               repository.UserRepository
	permissionRepo         repository.PermissionRepository
	authorizationCacheRepo repository.AuthorizationCacheRepository
	permissionCacheTTL     time.Duration
	logger                 *zap.Logger
}

func NewPermissionService(
	userRepo repository.UserRepository,
	permissionRepo repository.PermissionRepository,
	authorizationCacheRepo repository.AuthorizationCacheRepository,
	permissionCacheTTL time.Duration,
	logger *zap.Logger,
) *DefaultPermissionService {
	return &DefaultPermissionService{
		userRepo:               userRepo,
		permissionRepo:         permissionRepo,
		authorizationCacheRepo: authorizationCacheRepo,
		permissionCacheTTL:     permissionCacheTTL,
		logger:                 logger,
	}
}

// CheckAccess 先按路由模板匹配接口权限，再校验当前用户是否拥有对应权限码。
// 未配置到权限表的接口默认放行，便于按需逐步接入权限控制。
func (s *DefaultPermissionService) CheckAccess(ctx context.Context, userID uint64, roleCodes []string, method, path string) (bool, *model.Permission, error) {
	if hasRoleCode(roleCodes, model.RoleCodeSuperAdmin) {
		return true, nil, nil
	}

	permission, err := s.permissionRepo.GetAPIByMethodPath(ctx, method, path)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil, nil
		}
		return false, nil, err
	}

	codes, err := s.GetCurrentUserPermissionCodes(ctx, userID)
	if err != nil {
		return false, nil, err
	}

	for _, code := range codes {
		if code == permission.Code {
			return true, permission, nil
		}
	}

	return false, permission, nil
}

// GetCurrentUserMenus 返回当前用户可见的菜单树，供前端动态路由渲染。
func (s *DefaultPermissionService) GetCurrentUserMenus(ctx context.Context, userID uint64) ([]*MenuItem, error) {
	menus, err := s.userRepo.ListMenusByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return buildMenuTree(menus), nil
}

// GetCurrentUserPermissionCodes 优先读取缓存，未命中时回源数据库。
func (s *DefaultPermissionService) GetCurrentUserPermissionCodes(ctx context.Context, userID uint64) ([]string, error) {
	codes, err := s.authorizationCacheRepo.GetPermissionCodes(ctx, userID)
	if err == nil {
		return codes, nil
	}
	if err != nil && !errors.Is(err, repository.ErrCacheMiss) {
		s.logger.Warn("load permission codes from cache failed", zap.Uint64("user_id", userID), zap.Error(err))
	}

	codes, err = s.userRepo.ListPermissionCodesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := s.authorizationCacheRepo.SetPermissionCodes(ctx, userID, codes, s.permissionCacheTTL); err != nil {
		s.logger.Warn("write permission codes to cache failed", zap.Uint64("user_id", userID), zap.Error(err))
	}

	return codes, nil
}

// buildMenuTree 将扁平菜单权限构造成前端可直接消费的树结构。
func buildMenuTree(permissions []model.Permission) []*MenuItem {
	menuMap := make(map[uint64]*MenuItem, len(permissions))
	rootMenus := make([]*MenuItem, 0)

	for _, permission := range permissions {
		menuMap[permission.ID] = &MenuItem{
			ID:        permission.ID,
			Name:      permission.Name,
			Code:      permission.Code,
			Path:      permission.Path,
			RouteName: permission.RouteName,
			Component: permission.Component,
			Redirect:  permission.Redirect,
			Icon:      permission.Icon,
			Sort:      permission.Sort,
			Hidden:    permission.Hidden,
			Children:  make([]*MenuItem, 0),
		}
	}

	for _, permission := range permissions {
		current := menuMap[permission.ID]
		if permission.ParentID != nil {
			parent, ok := menuMap[*permission.ParentID]
			if ok {
				parent.Children = append(parent.Children, current)
				continue
			}
		}

		rootMenus = append(rootMenus, current)
	}

	return rootMenus
}

func hasRoleCode(roleCodes []string, expected string) bool {
	for _, roleCode := range roleCodes {
		if roleCode == expected {
			return true
		}
	}

	return false
}
