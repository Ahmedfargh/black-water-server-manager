package routes

import (
	authentication "github.com/ahmedfargh/server-manager/Authentication"
	info "github.com/ahmedfargh/server-manager/Info"
	"github.com/gin-gonic/gin"
)

func ProcessRoute(router *gin.Engine) {
	router.Group("/info")
	{
		router.Use(authentication.AuthMiddleware()).GET("/processes", authentication.CheckRole("read_processes"), info.GetProcessInfo())
		router.Use(authentication.AuthMiddleware()).GET("/process/single/:pid", authentication.CheckRole("read_process"), info.GetProcessByPID())
		router.Use(authentication.AuthMiddleware()).POST("/process/start", authentication.CheckRole("start_process"), info.StartProcess())
		router.Use(authentication.AuthMiddleware()).GET("/process/log", authentication.CheckRole("read_process_log"), info.GetProcessLog())
		router.Use(authentication.AuthMiddleware()).DELETE("/process/kill/:pid", authentication.CheckRole("kill_process"), info.KillProcess())
	}
}
