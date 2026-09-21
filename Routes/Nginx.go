package routes

import (
	authentication "github.com/ahmedfargh/server-manager/Authentication"
	controller "github.com/ahmedfargh/server-manager/FunctionalsControllers"
	"github.com/gin-gonic/gin"
)

func NginxRoutes(router *gin.Engine) {
	ctrl := controller.NewNginxController(nil)

	nginxGroup := router.Group("/nginx")
	nginxGroup.Use(authentication.AuthMiddleware())
	{
		// Overview & Status
		nginxGroup.GET("/overview", ctrl.GetOverview)

		// Sites / Virtual Hosts
		nginxGroup.GET("/sites", ctrl.GetSites)
		nginxGroup.GET("/sites/:name", ctrl.GetSite)
		nginxGroup.POST("/sites", ctrl.SaveSite)
		nginxGroup.DELETE("/sites/:name", ctrl.DeleteSite)
		nginxGroup.POST("/sites/:name/toggle", ctrl.ToggleSite)

		// Service & Config Controls
		nginxGroup.POST("/test", ctrl.TestConfig)
		nginxGroup.POST("/reload", ctrl.Reload)
		nginxGroup.POST("/restart", ctrl.Restart)

		// Logs & Analytics
		nginxGroup.GET("/logs/files", ctrl.GetLogFiles)
		nginxGroup.GET("/logs/access", ctrl.GetAccessLogs)
		nginxGroup.GET("/logs/error", ctrl.GetErrorLogs)
		nginxGroup.GET("/logs/analytics", ctrl.GetLogAnalytics)
	}
}
