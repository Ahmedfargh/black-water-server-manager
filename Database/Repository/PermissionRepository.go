package repository

import (
	"github.com/ahmedfargh/server-manager/Database/Models"
	"gorm.io/gorm"
)

type PermissionRepository struct {
	DB *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{DB: db}
}

func (r *PermissionRepository) CreatePermission(permission *models.Permission) error {
	return r.DB.Create(permission).Error
}

func (r *PermissionRepository) GetPermissionByName(name string) (*models.Permission, error) {
	var permission models.Permission
	err := r.DB.Where("name = ?", name).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}
