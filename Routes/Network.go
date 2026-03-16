package routes

import (
	authentication "github.com/ahmedfargh/server-manager/Authentication"
	info "github.com/ahmedfargh/server-manager/Info"
	"github.com/gin-gonic/gin"
)

func NetworkRoutes(router *gin.Engine) {
	router.GET("/network", authentication.AuthMiddleware(), authentication.CheckRole("read_network"), info.GetNetworkInfo())
	router.GET("/network/connections", authentication.AuthMiddleware(), authentication.CheckRole("read_network"), info.GetNetworkConnections())
}
