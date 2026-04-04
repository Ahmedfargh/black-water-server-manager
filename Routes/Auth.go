package routes

import (
	"fmt"

	authentication "github.com/ahmedfargh/server-manager/Authentication"
	service "github.com/ahmedfargh/server-manager/Authentication/Service"
	crud "github.com/ahmedfargh/server-manager/Database/CRUD"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.Engine, userCRUD *crud.UserCRUD, authService *service.AuthService, roleCRUD *crud.RoleCRUD) {
	router.POST("/login", authentication.Login(authService))
	router.POST("/register", authentication.Register(authService))

	authenticated := router.Group("/users").Use(authentication.AuthMiddleware())
	{
		fmt.Println("main grouping")
		rolesHandlers := authentication.NewRolesHandlers(roleCRUD)

		UserRoutes(authenticated.(*gin.RouterGroup), userCRUD, authService)
		RoleRoutes(authenticated.(*gin.RouterGroup), rolesHandlers)
	}
}

func UserRoutes(router *gin.RouterGroup, userCRUD *crud.UserCRUD, authService *service.AuthService) {
	usersGroup := router.Group("/users")
	{
		usersGroup.GET("/", authentication.CheckRole("read_user"), authentication.GetUsers(userCRUD))
		usersGroup.POST("/acount/update", authentication.UpdateProfile(userCRUD))
		usersGroup.POST("/notifications/settings", authentication.UpdateNotificationSettings(authService))
	}

	crudGroup := router.Group("/crud/users")
	{
		crudGroup.GET("/:id", authentication.CheckRole("read_user"), authentication.GetUser(userCRUD))
		crudGroup.POST("/", authentication.CheckRole("create_user"), authentication.CreateUser(userCRUD))
		crudGroup.PUT("/:id", authentication.CheckRole("update_user"), authentication.UpdateUser(userCRUD))
		crudGroup.DELETE("/:id", authentication.CheckRole("delete_user"), authentication.DeleteUser(userCRUD))
		crudGroup.GET("/list", authentication.CheckRole("read_user"), authentication.ListAll(userCRUD))
	}

	router.GET("/profile/me", authentication.GetProfile(userCRUD))
}

func RoleRoutes(router *gin.RouterGroup, rolesHandlers *authentication.RolesHandlers) {
	rolesGroup := router.Group("")
	{
		rolesGroup.Use(authentication.CheckRole("manage_roles"))
		rolesGroup.GET("/roles", rolesHandlers.GetRoles())
		rolesGroup.POST("/roles", rolesHandlers.CreateRole)
		rolesGroup.GET("/role/:id", rolesHandlers.GetRoleByID())
		rolesGroup.POST("/roles/update/:id", rolesHandlers.UpdateRole)
	}
}
