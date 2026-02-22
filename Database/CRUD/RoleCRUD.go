package crud

import (
	"errors" // Added import for errors package

	models "github.com/ahmedfargh/server-manager/Database/Models"
	repository "github.com/ahmedfargh/server-manager/Database/Repository"
	"gorm.io/gorm"
)

type RoleCRUD struct {
	Repo *repository.RoleRepository
}

func NewRoleCRUD(db *gorm.DB) *RoleCRUD {
	return &RoleCRUD{Repo: repository.NewRoleRepository(db)}
}

func (c *RoleCRUD) CreateRole(role *models.Role) error {
	return c.Repo.CreateRole(role)
}

func (c *RoleCRUD) GetRoleByName(name string) (*models.Role, error) {
	return c.Repo.GetRoleByName(name)
}

func (c *RoleCRUD) FindOrCreateRole(name string) (*models.Role, error) {
	role, err := c.GetRoleByName(name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newRole := &models.Role{Name: name}
			if createErr := c.CreateRole(newRole); createErr != nil {
				return nil, createErr
			}
			return newRole, nil
		}
		return nil, err
	}
	return role, nil
}
func (c *RoleCRUD) GetRoles() ([]models.Role, error) {
	return c.Repo.GetRoles()
}
func (c *RoleCRUD) GetRoleByID(id uint) (*models.Role, error) {
	return c.Repo.GetRoleByID(id)
}
func (c *RoleCRUD) UpdateRole(role_id uint, name string, permission_ids []uint) (bool, error) {
	return c.Repo.UpdateRole(role_id, name, permission_ids)
}
func (c *RoleCRUD) UserHasPermission(user *models.User, permissionName string) (bool, error) {
	for _, p := range user.Role.Permissions {
		if p.Name == permissionName {
			return true, nil
		}
	}
	return false, nil
}
