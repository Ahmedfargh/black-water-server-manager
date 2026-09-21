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
		"enable_firewall",
		"disable_firewall",
		"view_firewall_status",
		"view_firewall_rules",
		"view_firewall_list",
		"read_containers",
		"manage_containers",
		"read_packages",
		"manage_packages",
		"terminal_access",
		"read_systemd",
		"manage_systemd",
		"read_cron",
		"manage_cron",
		"view_audit_logs",
		"site_create",
		"site_checks",
		"site_delete",
		"site_update",
		"site_read",
		"site_analytics",
		"read_file",
		"write_file",
		"delete_file",
		"browse_filesystem",
		"manage_ssh_keys",
		"read_ssh_keys",
		"ssh_terminal_access",
		"manage_ssl",
		"read_ssl",
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
