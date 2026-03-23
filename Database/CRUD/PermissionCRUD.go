package crud

import (
	"errors"

	"github.com/ahmedfargh/server-manager/Database/Models"
	"github.com/ahmedfargh/server-manager/Database/Repository"
	"gorm.io/gorm"
)

type PermissionCRUD struct {
	Repo *Repository.PermissionRepository
}

func NewPermissionCRUD(db *gorm.DB) *PermissionCRUD {
	return &PermissionCRUD{Repo: Repository.NewPermissionRepository(db)}
}

func (c *PermissionCRUD) CreatePermission(permission *Models.Permission) error {
	return c.Repo.CreatePermission(permission)
}

func (c *PermissionCRUD) GetPermissionByName(name string) (*Models.Permission, error) {
	return c.Repo.GetPermissionByName(name)
}

func (c *PermissionCRUD) FindOrCreatePermission(name string) (*Models.Permission, error) {
	permission, err := c.GetPermissionByName(name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newPermission := &Models.Permission{Name: name}
			if createErr := c.CreatePermission(newPermission); createErr != nil {
				return nil, createErr
			}
			return newPermission, nil
		}
		return nil, err
	}
	return permission, nil
}
