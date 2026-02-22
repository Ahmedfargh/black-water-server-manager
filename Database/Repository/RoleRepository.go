package repository

import (
	"errors"

	models "github.com/ahmedfargh/server-manager/Database/Models"
	"gorm.io/gorm"
)

type RoleRepository struct {
	DB *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{DB: db}
}

func (r *RoleRepository) CreateRole(role *models.Role) error {
	tx := r.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	permissionsToAssociate := role.IncomingPermissionIDs
	role.Permissions = nil

	// Check if role with the same name already exists
	var existingRole models.Role
	if err := tx.Where("name = ?", role.Name).First(&existingRole).Error; err == nil {
		tx.Rollback()
		return errors.New("role with this name already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return err
	}

	if err := tx.Create(role).Error; err != nil {
		tx.Rollback()
		return err
	}
	if _, err := r.syncRoles(tx, role, permissionsToAssociate); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
func (r *RoleRepository) syncRoles(tx *gorm.DB, role *models.Role, permissionsToAssociate []uint) (bool, error) {
	if len(permissionsToAssociate) > 0 {
		var existingPermissions []models.Permission

		if err := tx.Where("id IN ?", permissionsToAssociate).Find(&existingPermissions).Error; err != nil {
			return false, err
		}

		if err := tx.Model(role).Association("Permissions").Replace(existingPermissions); err != nil {
			return false, err
		}
	} else {
		if err := tx.Model(role).Association("Permissions").Clear(); err != nil {
			return false, err
		}
	}
	return true, nil
}
func (r *RoleRepository) GetRoleByName(name string) (*models.Role, error) {
	var role models.Role
	err := r.DB.Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) GetRoles() ([]models.Role, error) {
	var roles []models.Role
	err := r.DB.Preload("Permissions").Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}
func (r *RoleRepository) GetRoleByID(id uint) (*models.Role, error) {
	var role models.Role
	err := r.DB.Preload("Permissions").First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) UpdateRole(role_id uint, name string, permission_ids []uint) (bool, error) {
	tx := r.DB.Begin()
	if tx.Error != nil {
		return false, tx.Error
	}
	// Check if another role with the same name already exists
	var existingRole models.Role
	if err := tx.Where("name = ? AND id != ?", name, role_id).First(&existingRole).Error; err == nil {
		tx.Rollback()
		return false, errors.New("another role with this name already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return false, err
	}
	var role models.Role
	if err := tx.First(&role, role_id).Error; err != nil {
		tx.Rollback()
		return false, err
	}

	if err := tx.Model(&role).Update("name", name).Error; err != nil {
		tx.Rollback()
		return false, err
	}

	if _, err := r.syncRoles(tx, &role, permission_ids); err != nil {
		tx.Rollback()
		return false, err
	}

	return true, tx.Commit().Error
}
