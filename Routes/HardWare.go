package routes

import (
	info "github.com/ahmedfargh/server-manager/Info"
	"github.com/gin-gonic/gin"
)

func CpuRoute(router *gin.Engine) {
	router.GET("/cpu", info.GetCputInfo())
	router.GET("/gpu", info.GetGpuInfo())
}
