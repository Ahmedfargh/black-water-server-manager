package routes

import (
	authentication "github.com/ahmedfargh/server-manager/Authentication"
	info "github.com/ahmedfargh/server-manager/Info"
	"github.com/gin-gonic/gin"
)

func CpuRoute(router *gin.Engine) {
	router.Group("/info")
	{
		router.Use(authentication.AuthMiddleware()).GET("/cpu", info.GetCputInfo())
		router.Use(authentication.AuthMiddleware()).GET("/gpu", info.GetGpuInfo())
		router.Use(authentication.AuthMiddleware()).GET("/ram", info.GetRamInfo())
	}
}
