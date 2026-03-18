package routes

import (
	authentication "github.com/ahmedfargh/server-manager/Authentication"
	functionalscontrollers "github.com/ahmedfargh/server-manager/FunctionalsControllers"
	"github.com/gin-gonic/gin"
)

func FireWallRoute(router *gin.Engine) {
	firewallGroup := router.Group("/firewall")
	{
		firewallGroup.GET("/enable", authentication.AuthMiddleware(), authentication.CheckRole("enable_firewall"), functionalscontrollers.EnableFireWallHandler())
		firewallGroup.GET("/disable", authentication.AuthMiddleware(), authentication.CheckRole("disable_firewall"), functionalscontrollers.DisableFireWallHandler())
		firewallGroup.GET("/status", authentication.AuthMiddleware(), authentication.CheckRole("view_firewall_status"), functionalscontrollers.StatusFireWallHandler())
		firewallGroup.GET("/rules", authentication.AuthMiddleware(), authentication.CheckRole("view_firewall_rules"), functionalscontrollers.RulesFireWallHandler())
		firewallGroup.GET("/list", authentication.AuthMiddleware(), authentication.CheckRole("view_firewall_rules"), functionalscontrollers.ListRulesFireWallHandler())
	}
}
