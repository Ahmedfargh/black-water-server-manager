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

	authenticated := router.Group("/").Use(authentication.AuthMiddleware())
	{
		fmt.Println("main grouping")
		rolesHandlers := authentication.NewRolesHandlers(roleCRUD)

		UserRoutes(authenticated.(*gin.RouterGroup), userCRUD)
		RoleRoutes(authenticated.(*gin.RouterGroup), rolesHandlers)
	}
}

func UserRoutes(router *gin.RouterGroup, userCRUD *crud.UserCRUD) {
	router.Group("/users").Use(authentication.CheckRole("read_users"))
	{
		router.GET("/", authentication.GetUsers(userCRUD))
		router.POST("/acount/update", authentication.UpdateProfile(userCRUD))
	}

	fmt.Println("crud user grouping")
	router.GET("/crud/users/:id", authentication.GetUser(userCRUD))
	router.POST("/crud/users/", authentication.CreateUser(userCRUD))
	router.PUT("/crud/users/:id", authentication.UpdateUser(userCRUD))
	router.DELETE("/crud/users/:id", authentication.DeleteUser(userCRUD))
	router.GET("/crud/users/list", authentication.ListAll(userCRUD))

	router.GET("/profile/me", authentication.GetProfile(userCRUD))
}

func RoleRoutes(router *gin.RouterGroup, rolesHandlers *authentication.RolesHandlers) {
	router.GET("/roles", rolesHandlers.GetRoles())
	router.POST("/roles", rolesHandlers.CreateRole)
	router.GET("/role/:id", rolesHandlers.GetRoleByID())
	router.POST("/roles/update/:id", rolesHandlers.UpdateRole)
}
