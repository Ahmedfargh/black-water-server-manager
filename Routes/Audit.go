package routes

import (
	authentication "github.com/ahmedfargh/server-manager/Authentication"
	functionalscontrollers "github.com/ahmedfargh/server-manager/FunctionalsControllers"
	"github.com/gin-gonic/gin"
)

func AuditRoutes(router *gin.Engine) {
	auditGroup := router.Group("/audit").Use(authentication.AuthMiddleware())
	{
		auditGroup.GET("/list", authentication.CheckRole("view_audit_logs"), functionalscontrollers.GetAuditLogHandler())
	}
}
