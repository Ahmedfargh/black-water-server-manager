package seeder

import (
	"fmt"

	crud "github.com/ahmedfargh/server-manager/Database/CRUD"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedAll orchestrates the seeding of permissions, roles, and users
func SeedAll(userCRUD *crud.UserCRUD, roleCRUD *crud.RoleCRUD, permissionCRUD *crud.PermissionCRUD) {
	fmt.Println("Starting database seeding...")

	// 1. Seed Permissions
	SeedPermissions(permissionCRUD)

	// 2. Define Roles and their Permissions
	superAdminPermissions := []string{
		"create_user", "read_user", "update_user", "delete_user",
		"manage_roles", "manage_permissions",
	}
	developerPermissions := []string{
		"read_user", "update_user",
	}

	// 3. Create Roles and assign permissions
	seedRole(roleCRUD, permissionCRUD, "super_admin", superAdminPermissions)
	seedRole(roleCRUD, permissionCRUD, "developer", developerPermissions)
	seedRole(roleCRUD, permissionCRUD, "user", []string{"read_user"}) // Ensure 'user' role exists with basic permission

	// 4. Seed Users
	users := []struct {
		Username  string
		Email     string
		Password  string
		RoleName  string
		ImagePath string
		Status    bool
	}{
		{Username: "admin", Email: "admin@example.com", Password: "password", RoleName: "super_admin", ImagePath: "", Status: true},
		{Username: "user1", Email: "user1@example.com", Password: "password", RoleName: "user", ImagePath: "", Status: true},
		{Username: "dev1", Email: "dev1@example.com", Password: "password", RoleName: "developer", ImagePath: "", Status: true},
		{Username: "moderator", Email: "mod@example.com", Password: "password", RoleName: "developer", ImagePath: "", Status: true},
		{Username: "testuser", Email: "test@example.com", Password: "password", RoleName: "user", ImagePath: "", Status: true},
	}

	for _, userData := range users {
		role, err := roleCRUD.GetRoleByName(userData.RoleName)
		if err != nil {
			fmt.Printf("Error getting role %s for user %s: %v\n", userData.RoleName, userData.Username, err)
			continue
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Printf("Error hashing password for user %s: %v\n", userData.Username, err)
			continue
		}

		// Check if user already exists
		existingUser, err := userCRUD.GetUserByEmail(userData.Email)
		if err == nil && existingUser != nil {
			// Update existing user
			existingUser.Username = userData.Username
			existingUser.Password = string(hashedPassword)
			existingUser.RoleID = role.ID
			existingUser.ImagePath = userData.ImagePath
			existingUser.Status = userData.Status

			err = userCRUD.UpdateUser(existingUser, existingUser.ID)
			if err != nil {
				fmt.Printf("Error updating user %s: %v\n", existingUser.Username, err)
			} else {
				fmt.Printf("Updated user: %s\n", existingUser.Username)
			}
			continue
		}

		user := &models.User{
			Username:  userData.Username,
			Email:     userData.Email,
			Password:  string(hashedPassword),
			RoleID:    role.ID,
			ImagePath: userData.ImagePath,
			Status:    userData.Status,
		}

		err = userCRUD.CreateUser(user)
		if err != nil {
			fmt.Printf("Error seeding user %s: %v\n", user.Username, err)
		} else {
			fmt.Printf("Seeded user: %s\n", user.Username)
		}
	}

	fmt.Println("Database seeding completed.")
}

func seedRole(roleCRUD *crud.RoleCRUD, permissionCRUD *crud.PermissionCRUD, roleName string, permissionNames []string) {
	role, err := roleCRUD.FindOrCreateRole(roleName)
	if err != nil {
		fmt.Printf("Error finding or creating role %s: %v\n", roleName, err)
		return
	}

	// Clear existing permissions for the role if any, then add new ones
	// This part needs more sophisticated handling for many2many, for now, we'll just associate.
	// GORM will handle duplicates if a permission is already associated.

	var permissionsToAdd []models.Permission
	for _, pName := range permissionNames {
		permission, err := permissionCRUD.FindOrCreatePermission(pName)
		if err != nil {
			fmt.Printf("Error finding or creating permission %s for role %s: %v\n", pName, roleName, err)
			continue
		}
		permissionsToAdd = append(permissionsToAdd, *permission)
	}

	// This is a basic association, GORM will add new ones if they don't exist
	// For actual "clear then add", you'd need to manipulate the join table directly.
	role.Permissions = permissionsToAdd
	err = roleCRUD.Repo.DB.Session(&gorm.Session{FullSaveAssociations: true}).Save(role).Error
	if err != nil {
		fmt.Printf("Error associating permissions with role %s: %v\n", roleName, err)
	} else {
		fmt.Printf("Seeded role: %s with %d permissions\n", roleName, len(permissionsToAdd))
	}
}
