package crud

import (
	"errors"
	"github.com/ahmedfargh/server-manager/Database/Models"
	"github.com/ahmedfargh/server-manager/Database/Repository"
	"gorm.io/gorm"
)

type PermissionCRUD struct {
	Repo *repository.PermissionRepository
}

func NewPermissionCRUD(db *gorm.DB) *PermissionCRUD {
	return &PermissionCRUD{Repo: repository.NewPermissionRepository(db)}
}

func (c *PermissionCRUD) CreatePermission(permission *models.Permission) error {
	return c.Repo.CreatePermission(permission)
}

func (c *PermissionCRUD) GetPermissionByName(name string) (*models.Permission, error) {
	return c.Repo.GetPermissionByName(name)
}

func (c *PermissionCRUD) FindOrCreatePermission(name string) (*models.Permission, error) {
	permission, err := c.GetPermissionByName(name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newPermission := &models.Permission{Name: name}
			if createErr := c.CreatePermission(newPermission); createErr != nil {
				return nil, createErr
			}
			return newPermission, nil
		}
		return nil, err
	}
	return permission, nil
}
