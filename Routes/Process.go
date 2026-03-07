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
		router.Use(authentication.AuthMiddleware()).GET("/process/single/:pid", info.GetProcessByPID())
		router.Use(authentication.AuthMiddleware()).POST("/process/start", info.StartProcess())
		router.Use(authentication.AuthMiddleware()).GET("/process/log", info.GetProcessLog())
		router.Use(authentication.AuthMiddleware()).DELETE("/process/kill/:pid", info.KillProcess())
	}
}
