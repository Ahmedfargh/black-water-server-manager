package models

import (
	"fmt"

	config "github.com/ahmedfargh/server-manager/Config"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username    string       `gorm:"unique;not null" json:"username" form:"username" validate:"required,min=3,max=32"`
	Email       string       `gorm:"unique;not null" json:"email" form:"email" validate:"required,email"`
	Password    string       `gorm:"not null" json:"password" form:"password" validate:"required,min=8"`
	RoleID      uint         `gorm:"not null" json:"role_id" form:"role_id"`
	Role        Role         // Belongs to Role
	ImagePath   string       `json:"image_path"`
	Status      bool         `gorm:"default:false" json:"status" form:"status"`
	Permissions []Permission `gorm:"many2many:user_permissions;"`
}

func (User) TableName() string {
	return "users"
}

func (u User) ToMap() map[string]interface{} {
	fmt.Println(u.Role.Permissions)
	return map[string]interface{}{
		"id":               u.ID,
		"username":         u.Username,
		"email":            u.Email,
		"role":             u.Role.Name,
		"image_path":       config.GetKey("APP_URL") + u.ImagePath,
		"status":           u.Status,
		"permissions":      u.Permissions,
		"roles_permssions": u.Role.Permissions,
	}
}
func (u User) HasPermission(permission string) bool {
	for _, p := range u.Permissions {
		if p.Name == permission {
			return true
		}
	}
	for _, p := range u.Role.Permissions {
		if p.Name == permission {
			return true
		}
	}

	return false
}
