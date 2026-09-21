package routes

import (
	authentication "github.com/ahmedfargh/server-manager/Authentication"
	controller "github.com/ahmedfargh/server-manager/FunctionalsControllers"
	"github.com/gin-gonic/gin"
)

func SSLRoutes(router *gin.Engine) {
	sslGroup := router.Group("/ssl")
	{
		sslGroup.GET("/inspect", authentication.AuthMiddleware(), authentication.CheckRole("read_ssl"), controller.InspectSSLHandler())
		sslGroup.POST("/inspect-file", authentication.AuthMiddleware(), authentication.CheckRole("read_ssl"), controller.InspectCertFileHandler())
		sslGroup.POST("/issue-letsencrypt", authentication.AuthMiddleware(), authentication.CheckRole("manage_ssl"), controller.IssueCertbotHandler())
		sslGroup.POST("/sites/:id/check", authentication.AuthMiddleware(), authentication.CheckRole("manage_ssl"), controller.CheckSiteSSLHandler())
		sslGroup.POST("/sites/:id/toggle-renew", authentication.AuthMiddleware(), authentication.CheckRole("manage_ssl"), controller.ToggleSiteAutoRenewHandler())
	}
}
