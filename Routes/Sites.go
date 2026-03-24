package routes

import (
	authentication "github.com/ahmedfargh/server-manager/Authentication"
	controller "github.com/ahmedfargh/server-manager/FunctionalsControllers"
	"github.com/gin-gonic/gin"
)

func SiteRoutes(router *gin.Engine) {
	siteRouters := router.Group("/site")
	{
		siteRouters.POST("/create", authentication.AuthMiddleware(), authentication.CheckRole("site_create"), controller.CreateSiteHandler())
		siteRouters.GET("/list", authentication.AuthMiddleware(), authentication.CheckRole("site_read"), controller.GetSitesHandler())
	}
}
