package routes

import (
	authentication "github.com/ahmedfargh/server-manager/Authentication"
	functionalscontrollers "github.com/ahmedfargh/server-manager/FunctionalsControllers"
	"github.com/gin-gonic/gin"
)

// RegisterPackageManagerRoutes registers package inspection and lifecycle endpoints
func RegisterPackageManagerRoutes(router *gin.Engine) {
	ctrl := functionalscontrollers.NewPackageManagerController(nil)

	pkgGroup := router.Group("/packages").Use(authentication.AuthMiddleware())
	{
		// Query & Inspection
		pkgGroup.GET("/overview", authentication.CheckRole("read_packages"), ctrl.GetOverview)
		pkgGroup.GET("/:manager/updates", authentication.CheckRole("read_packages"), ctrl.GetUpdates)
		pkgGroup.GET("/:manager/list", authentication.CheckRole("read_packages"), ctrl.GetPackages)

		// Lifecycle Mutations & Maintenance
		pkgGroup.POST("/:manager/clean-cache", authentication.CheckRole("manage_packages"), ctrl.CleanCache)
		pkgGroup.POST("/:manager/refresh", authentication.CheckRole("manage_packages"), ctrl.RefreshRepositories)
		pkgGroup.POST("/:manager/upgrade-system", authentication.CheckRole("manage_packages"), ctrl.UpgradeSystem)
		pkgGroup.POST("/:manager/install", authentication.CheckRole("manage_packages"), ctrl.InstallPackage)
		pkgGroup.POST("/:manager/remove", authentication.CheckRole("manage_packages"), ctrl.RemovePackage)
		pkgGroup.POST("/:manager/upgrade-package", authentication.CheckRole("manage_packages"), ctrl.UpgradePackage)
	}
}
