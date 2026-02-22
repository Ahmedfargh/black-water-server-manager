package seeder

import (
	"fmt"

	crud "github.com/ahmedfargh/server-manager/Database/CRUD"
)

func SeedPermissions(permissionCRUD *crud.PermissionCRUD) {
	permissions := []string{
		"create_user",
		"read_user",
		"update_user",
		"delete_user",
		"manage_roles",
		"manage_permissions",
	}

	for _, p := range permissions {
		_, err := permissionCRUD.FindOrCreatePermission(p)
		if err != nil {
			fmt.Printf("Error seeding permission %s: %v\n", p, err)
		} else {
			fmt.Printf("Seeded permission: %s\n", p)
		}
	}
}
