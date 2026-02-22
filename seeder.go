package main

import (
	config "github.com/ahmedfargh/server-manager/Config"
	crud "github.com/ahmedfargh/server-manager/Database/CRUD"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	repository "github.com/ahmedfargh/server-manager/Database/Repository"
	seeder "github.com/ahmedfargh/server-manager/Database/Seeder"
)

func main() {
	// Initialize Database Connection
	config.ConnectDB()

	// Ensure tables exist before seeding
	config.DB.AutoMigrate(&models.User{}, &models.Role{}, &models.Permission{})

	// Initialize Dependencies
	userRepo := repository.NewUserRepository(config.DB)
	userCRUD := crud.NewUserCRUD(userRepo)
	roleCRUD := crud.NewRoleCRUD(config.DB)
	permissionCRUD := crud.NewPermissionCRUD(config.DB)

	// Run Seeder
	seeder.SeedAll(userCRUD, roleCRUD, permissionCRUD)
}
