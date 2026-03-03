package routes

import (
	authentication "github.com/ahmedfargh/server-manager/Authentication"
	info "github.com/ahmedfargh/server-manager/Info"
	"github.com/gin-gonic/gin"
)

func ProcessRoute(router *gin.Engine) {
	router.Group("/info")
	{
		router.Use(authentication.AuthMiddleware()).GET("/processes", info.GetProcessInfo())
	}
}
