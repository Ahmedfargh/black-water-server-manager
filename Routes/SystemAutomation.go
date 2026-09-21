package routes

import (
	authentication "github.com/ahmedfargh/server-manager/Authentication"
	functionalscontrollers "github.com/ahmedfargh/server-manager/FunctionalsControllers"
	"github.com/gin-gonic/gin"
)

func SystemAutomationRoutes(router *gin.Engine) {
	systemdCtrl := functionalscontrollers.NewSystemdController(nil, nil)
	cronCtrl := functionalscontrollers.NewCronController(nil, nil)

	// Systemd Units Group
	systemdGroup := router.Group("/systemd")
	systemdGroup.Use(authentication.AuthMiddleware())
	{
		systemdGroup.GET("/units", authentication.CheckRole("read_systemd"), systemdCtrl.GetOverview)
		systemdGroup.GET("/unit/:unit/status", authentication.CheckRole("read_systemd"), systemdCtrl.GetUnitStatus)
		systemdGroup.POST("/unit/:unit/action", authentication.CheckRole("manage_systemd"), systemdCtrl.ExecuteAction)
	}

	// Cron Tasks Group
	cronGroup := router.Group("/cron")
	cronGroup.Use(authentication.AuthMiddleware())
	{
		cronGroup.GET("/jobs", authentication.CheckRole("read_cron"), cronCtrl.ListJobs)
		cronGroup.POST("/jobs", authentication.CheckRole("manage_cron"), cronCtrl.CreateOrUpdateJob)
		cronGroup.DELETE("/jobs/:id", authentication.CheckRole("manage_cron"), cronCtrl.DeleteJob)
		cronGroup.POST("/jobs/:id/toggle", authentication.CheckRole("manage_cron"), cronCtrl.ToggleJob)
		cronGroup.POST("/jobs/:id/run", authentication.CheckRole("manage_cron"), cronCtrl.RunJob)
	}
}
