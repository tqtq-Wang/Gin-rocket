package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	StatusDisabled int8 = 0
	StatusEnabled  int8 = 1
)

type User struct {
	ID          uint64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Username    string     `json:"username" gorm:"size:64;not null;uniqueIndex"`
	Password    string     `json:"-" gorm:"size:255;not null"`
	Nickname    string     `json:"nickname" gorm:"size:128;not null"`
	Email       string     `json:"email" gorm:"size:255;uniqueIndex"`
	Status      int8       `json:"status" gorm:"type:tinyint;not null;default:1"`
	LastLoginAt *time.Time `json:"last_login_at"`
	Roles       []Role     `json:"roles,omitempty" gorm:"many2many:user_roles;constraint:OnDelete:CASCADE;"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Role{},
		&Permission{},
		&UserRole{},
		&RolePermission{},
	)
}
