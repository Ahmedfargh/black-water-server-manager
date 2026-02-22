package models

import "gorm.io/gorm"

type Role struct {
	gorm.Model
	Name        string       `gorm:"unique;not null" json:"name"`
	Permissions []Permission `gorm:"many2many:role_permissions;"`
	IncomingPermissionIDs []uint `json:"role_permissions" gorm:"-"`
}
