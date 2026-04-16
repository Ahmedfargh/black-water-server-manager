package routes

import (
	authentication "github.com/ahmedfargh/server-manager/Authentication"
	FunctionalsControllers "github.com/ahmedfargh/server-manager/FunctionalsControllers"
	"github.com/gin-gonic/gin"
)

func RegisterHardwareReportRoutes(router *gin.Engine) {
	hwReportGroup := router.Group("/hardware-report")
	{
		hwReportGroup.GET("/latest", authentication.AuthMiddleware(), FunctionalsControllers.GetHardwareReportHandler())
		hwReportGroup.POST("/by-time-range", authentication.AuthMiddleware(), FunctionalsControllers.GetHardwareReportByTimeRangeHandler())
		hwReportGroup.POST("/average-usage-by-time-range", authentication.AuthMiddleware(), FunctionalsControllers.GetAverageHardwareUsageByTimeRangeHandler())
	}
}
