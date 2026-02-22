package main

import (
	service "github.com/ahmedfargh/server-manager/Authentication/Service"
	config "github.com/ahmedfargh/server-manager/Config"
	crud "github.com/ahmedfargh/server-manager/Database/CRUD"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	repository "github.com/ahmedfargh/server-manager/Database/Repository"
	routes "github.com/ahmedfargh/server-manager/Routes"
	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDB()
	config.DB.AutoMigrate(&models.User{}, &models.Role{}, &models.Permission{})
	userRepo := repository.NewUserRepository(config.DB)
	userCRUD := crud.NewUserCRUD(userRepo)
	roleCRUD := crud.NewRoleCRUD(config.DB)
	permissionCRUD := crud.NewPermissionCRUD(config.DB)
	authService := service.NewAuthService(userCRUD, roleCRUD)
	router := gin.Default()

	// Serve static files from the uploads directory
	router.Static("/uploads", "./uploads")

	routes.AuthRoutes(router, userCRUD, authService, roleCRUD)
	router.Run(config.PortNumber())
}
