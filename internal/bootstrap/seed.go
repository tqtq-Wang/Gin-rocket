package bootstrap

import (
	"errors"

	"gin-rocket/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// seedRBACData 只初始化 RBAC 基础元数据，不预置管理员用户。
// 首个注册用户会在注册事务里自动绑定 super-admin 角色。
func seedRBACData(db *gorm.DB, log *zap.Logger) error {
	return db.Transaction(func(tx *gorm.DB) error {
		role := &model.Role{}
		err := tx.Where("code = ?", model.RoleCodeSuperAdmin).First(role).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			role = &model.Role{
				Name:        "超级管理员",
				Code:        model.RoleCodeSuperAdmin,
				Description: "拥有全部系统权限",
				Status:      model.StatusEnabled,
			}
			if err := tx.Create(role).Error; err != nil {
				return err
			}
		}

		systemMenu := model.Permission{
			Name:        "系统管理",
			Code:        "system:manage",
			Type:        model.PermissionTypeMenu,
			Path:        "/system",
			RouteName:   "SystemRoot",
			Component:   "Layout",
			Redirect:    "/system/users",
			Icon:        "setting",
			Sort:        1,
			Status:      model.StatusEnabled,
			Description: "系统管理菜单",
		}
		if err := upsertPermission(tx, &systemMenu); err != nil {
			return err
		}

		userMenu := model.Permission{
			Name:        "用户管理",
			Code:        "system:user:menu",
			Type:        model.PermissionTypeMenu,
			Path:        "/system/users",
			RouteName:   "SystemUsers",
			Component:   "system/user/index",
			ParentID:    &systemMenu.ID,
			Sort:        10,
			Status:      model.StatusEnabled,
			Description: "用户管理菜单",
		}
		if err := upsertPermission(tx, &userMenu); err != nil {
			return err
		}

		userReadPermission := model.Permission{
			Name:        "查看用户详情",
			Code:        "system:user:read",
			Type:        model.PermissionTypeAPI,
			Method:      "GET",
			Path:        "/api/v1/users/:id",
			Sort:        110,
			Status:      model.StatusEnabled,
			Description: "查看用户详情接口权限",
		}
		if err := upsertPermission(tx, &userReadPermission); err != nil {
			return err
		}

		userCreateButtonPermission := model.Permission{
			Name:        "新增用户按钮",
			Code:        "system:user:create",
			Type:        model.PermissionTypeButton,
			Path:        "system:user:create",
			Sort:        120,
			Status:      model.StatusEnabled,
			Description: "前端新增用户按钮权限",
		}
		if err := upsertPermission(tx, &userCreateButtonPermission); err != nil {
			return err
		}

		fileUploadPermission := model.Permission{
			Name:        "上传文件",
			Code:        "system:file:upload",
			Type:        model.PermissionTypeAPI,
			Method:      "POST",
			Path:        "/api/v1/files/upload",
			Sort:        130,
			Status:      model.StatusEnabled,
			Description: "上传单文件接口权限",
		}
		if err := upsertPermission(tx, &fileUploadPermission); err != nil {
			return err
		}

		fileUploadMultiplePermission := model.Permission{
			Name:        "批量上传文件",
			Code:        "system:file:upload-multiple",
			Type:        model.PermissionTypeAPI,
			Method:      "POST",
			Path:        "/api/v1/files/upload-multiple",
			Sort:        131,
			Status:      model.StatusEnabled,
			Description: "批量上传文件接口权限",
		}
		if err := upsertPermission(tx, &fileUploadMultiplePermission); err != nil {
			return err
		}

		fileDeletePermission := model.Permission{
			Name:        "删除文件",
			Code:        "system:file:delete",
			Type:        model.PermissionTypeAPI,
			Method:      "DELETE",
			Path:        "/api/v1/files",
			Sort:        132,
			Status:      model.StatusEnabled,
			Description: "删除文件接口权限",
		}
		if err := upsertPermission(tx, &fileDeletePermission); err != nil {
			return err
		}

		filePresignPermission := model.Permission{
			Name:        "获取文件访问链接",
			Code:        "system:file:presign",
			Type:        model.PermissionTypeAPI,
			Method:      "GET",
			Path:        "/api/v1/files/presign",
			Sort:        133,
			Status:      model.StatusEnabled,
			Description: "获取文件预签名链接接口权限",
		}
		if err := upsertPermission(tx, &filePresignPermission); err != nil {
			return err
		}

		var permissions []model.Permission
		if err := tx.Where("status = ?", model.StatusEnabled).Find(&permissions).Error; err != nil {
			return err
		}

		if err := tx.Model(role).Association("Permissions").Replace(&permissions); err != nil {
			return err
		}

		log.Info("rbac metadata seeded", zap.String("role_code", model.RoleCodeSuperAdmin))
		return nil
	})
}

// upsertPermission 通过 code 保持权限元数据可重复执行。
func upsertPermission(tx *gorm.DB, permission *model.Permission) error {
	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "code"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name",
			"type",
			"method",
			"path",
			"route_name",
			"component",
			"redirect",
			"icon",
			"parent_id",
			"sort",
			"hidden",
			"status",
			"description",
			"updated_at",
		}),
	}).Create(permission).Error; err != nil {
		return err
	}

	return tx.Where("code = ?", permission.Code).First(permission).Error
}
