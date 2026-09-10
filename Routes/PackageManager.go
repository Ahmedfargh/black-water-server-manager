package routes

import (
	authentication "github.com/ahmedfargh/server-manager/Authentication"
	functionalscontrollers "github.com/ahmedfargh/server-manager/FunctionalsControllers"
	"github.com/gin-gonic/gin"
)

// RegisterPackageManagerRoutes registers package inspection endpoints
func RegisterPackageManagerRoutes(router *gin.Engine) {
	ctrl := functionalscontrollers.NewPackageManagerController(nil)

	pkgGroup := router.Group("/packages").Use(authentication.AuthMiddleware())
	{
		pkgGroup.GET("/overview", authentication.CheckRole("read_packages"), ctrl.GetOverview)
		pkgGroup.GET("/:manager/updates", authentication.CheckRole("read_packages"), ctrl.GetUpdates)
		pkgGroup.GET("/:manager/list", authentication.CheckRole("read_packages"), ctrl.GetPackages)
	}
}
