package routes

import (
	authentication "github.com/ahmedfargh/server-manager/Authentication"
	info "github.com/ahmedfargh/server-manager/Info"
	"github.com/gin-gonic/gin"
)

func CpuRoute(router *gin.Engine) {
	router.Group("/info")
	{
		router.Use(authentication.AuthMiddleware()).GET("/cpu", authentication.CheckRole("read_cpu"), info.GetCputInfo())
		router.Use(authentication.AuthMiddleware()).GET("/gpu", authentication.CheckRole("read_gpu"), info.GetGpuInfo())
		router.Use(authentication.AuthMiddleware()).GET("/ram", authentication.CheckRole("read_ram"), info.GetRamInfo())
		router.Use(authentication.AuthMiddleware()).GET("/disk", authentication.CheckRole("read_disk"), info.GetDiskInfo())
		router.Use(authentication.AuthMiddleware()).GET("/report", authentication.CheckRole("read_cpu"), info.GetHardWareReport())
	}
}
