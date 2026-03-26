package routes

import (
	Authentication "github.com/ahmedfargh/server-manager/Authentication"
	functionalscontrollers "github.com/ahmedfargh/server-manager/FunctionalsControllers"
	"github.com/gin-gonic/gin"
)

func SetupDockerRoutes(router *gin.Engine) {
	dockerGroup := router.Group("/docker")
	{
		dockerGroup.POST("/container", Authentication.AuthMiddleware(), Authentication.CheckRole("manage_containers"), functionalscontrollers.CreateDockerHandler())
		dockerGroup.PUT("/container/:id", Authentication.AuthMiddleware(), Authentication.CheckRole("manage_containers"), functionalscontrollers.UpdateDockerHandler())
		dockerGroup.DELETE("/container/:id", Authentication.AuthMiddleware(), Authentication.CheckRole("manage_containers"), functionalscontrollers.DeleteDockerHandler())
		// dockerGroup.POST("/container/:id/limits", Authentication.AuthMiddleware(), Authentication.CheckRole("manage_containers"), functionalscontrollers.SetDockerLimitsHandler())
		dockerGroup.GET("/containers", Authentication.AuthMiddleware(), Authentication.CheckRole("read_containers"), functionalscontrollers.GetContainersHandler())
		dockerGroup.GET("/container/:id", Authentication.AuthMiddleware(), Authentication.CheckRole("read_containers"), functionalscontrollers.GetContainerByIDHandler())
		dockerGroup.GET("/container/:id/status", Authentication.AuthMiddleware(), Authentication.CheckRole("read_containers"), functionalscontrollers.ContainerStatusHandler())
		dockerGroup.POST("/container/:id/:action", Authentication.AuthMiddleware(), Authentication.CheckRole("manage_containers"), functionalscontrollers.ActionContainerHandler())
	}
}
