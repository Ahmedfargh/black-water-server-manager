package authentication

import (
	"fmt"

	crud "github.com/ahmedfargh/server-manager/Database/CRUD"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	"github.com/gin-gonic/gin"
)

type RolesHandlers struct {
	RoleCRUD *crud.RoleCRUD
}

func NewRolesHandlers(roleCRUD *crud.RoleCRUD) *RolesHandlers {
	return &RolesHandlers{RoleCRUD: roleCRUD}
}

func (h *RolesHandlers) GetRoles() gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, err := h.RoleCRUD.GetRoles()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, roles)
	}
}
func (h *RolesHandlers) GetRoleByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		role_id_params := c.Param("id")
		var role_id uint
		if _, err := fmt.Sscan(role_id_params, &role_id); err != nil {
			c.JSON(400, gin.H{"error": "invalid id"})
			return
		}

		role, err := h.RoleCRUD.GetRoleByID(role_id)

		if err != nil {
			if err.Error() == "record not found" {
				c.JSON(404, gin.H{"error": "role not found"})
				return
			}
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, role)
	}
}
func (h *RolesHandlers) CreateRole(c *gin.Context) {
	type CreateRoleRequest struct {
		Name          string `json:"name" binding:"required"`
		PermissionIDs []uint `json:"permission_ids" binding:"required"`
	}
	var input CreateRoleRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Manually populate role.Permissions from IncomingPermissionIDs
	var role models.Role
	role.Name = input.Name

	for index := range input.PermissionIDs {

		role.IncomingPermissionIDs = append(role.IncomingPermissionIDs, input.PermissionIDs[index])
	}

	if err := h.RoleCRUD.CreateRole(&role); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, role)
}
func (h *RolesHandlers) UpdateRole(c *gin.Context) {
	role_id_params := c.Param("id")
	var role_data struct {
		Name           string `json:"name" binding:"required"`
		Permission_ids []uint `json:"permission_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&role_data); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	fmt.Println(role_data)
	var role_id uint
	if _, err := fmt.Sscan(role_id_params, &role_id); err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}
	results, err := h.RoleCRUD.UpdateRole(role_id, role_data.Name, role_data.Permission_ids)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(results)
	c.JSON(200, role_data)
}
