package routes

import (
	MiddleWare "github.com/ahmedfargh/server-manager/Authentication"
	functionController "github.com/ahmedfargh/server-manager/FunctionalsControllers"
	"github.com/gin-gonic/gin"
)

func RegisterFileSystemRoutes(router *gin.Engine) {
	file_system_router_group := router.Group("/filesystem")
	{
		file_system_router_group.GET("/browse", MiddleWare.AuthMiddleware(), MiddleWare.CheckRole("browse_filesystem"), functionController.FileBrowserHandler)
	}
}
