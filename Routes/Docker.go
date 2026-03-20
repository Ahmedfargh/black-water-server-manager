package routes

import (
	Authentication "github.com/ahmedfargh/server-manager/Authentication"
	functionalscontrollers "github.com/ahmedfargh/server-manager/FunctionalsControllers"
	"github.com/gin-gonic/gin"
)

func SetupDockerRoutes(router *gin.Engine) {
	dockerGroup := router.Group("/docker")
	{
		dockerGroup.GET("/containers", Authentication.AuthMiddleware(), Authentication.CheckRole("read_containers"), functionalscontrollers.GetContainersHandler())
		dockerGroup.GET("/container/:id", Authentication.AuthMiddleware(), Authentication.CheckRole("read_containers"), functionalscontrollers.GetContainerByIDHandler())
	}
}
