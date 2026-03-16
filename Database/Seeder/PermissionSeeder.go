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
		"read_processes",
		"read_process",
		"start_process",
		"read_process_log",
		"kill_process",
		"read_cpu",
		"read_gpu",
		"read_ram",
		"read_disk",
		"read_network",
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
