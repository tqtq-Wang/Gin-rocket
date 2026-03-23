package model

import "time"

const (
	PermissionTypeAPI  = "api"
	PermissionTypeMenu = "menu"

	RoleCodeSuperAdmin = "super-admin"
)

type Role struct {
	ID          uint64       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string       `json:"name" gorm:"size:128;not null"`
	Code        string       `json:"code" gorm:"size:64;not null;uniqueIndex"`
	Description string       `json:"description" gorm:"size:255"`
	Status      int8         `json:"status" gorm:"type:tinyint;not null;default:1"`
	Permissions []Permission `json:"-" gorm:"many2many:role_permissions;constraint:OnDelete:CASCADE;"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type Permission struct {
	ID          uint64       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string       `json:"name" gorm:"size:128;not null"`
	Code        string       `json:"code" gorm:"size:128;not null;uniqueIndex"`
	Type        string       `json:"type" gorm:"size:32;not null;index"`
	Method      string       `json:"method,omitempty" gorm:"size:16;index"`
	Path        string       `json:"path" gorm:"size:255;not null;index"`
	RouteName   string       `json:"route_name,omitempty" gorm:"size:128"`
	Component   string       `json:"component,omitempty" gorm:"size:255"`
	Redirect    string       `json:"redirect,omitempty" gorm:"size:255"`
	Icon        string       `json:"icon,omitempty" gorm:"size:128"`
	ParentID    *uint64      `json:"parent_id,omitempty" gorm:"index"`
	Sort        int          `json:"sort" gorm:"not null;default:0"`
	Hidden      bool         `json:"hidden" gorm:"not null;default:false"`
	Status      int8         `json:"status" gorm:"type:tinyint;not null;default:1"`
	Description string       `json:"description" gorm:"size:255"`
	Children    []Permission `json:"children,omitempty" gorm:"-"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type UserRole struct {
	UserID    uint64    `gorm:"primaryKey;autoIncrement:false"`
	RoleID    uint64    `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt time.Time `json:"created_at"`
}

func (UserRole) TableName() string {
	return "user_roles"
}

type RolePermission struct {
	RoleID       uint64    `gorm:"primaryKey;autoIncrement:false"`
	PermissionID uint64    `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt    time.Time `json:"created_at"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
