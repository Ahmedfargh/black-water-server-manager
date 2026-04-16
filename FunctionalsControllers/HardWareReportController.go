package functionalscontrollers

import (
	"fmt"
	"time"

	Config "github.com/ahmedfargh/server-manager/Config"
	crud "github.com/ahmedfargh/server-manager/Database/CRUD"
	"github.com/gin-gonic/gin"
)

func GetHardwareReportHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		hwReportCRUD := crud.NewHardWareReportCRUD(Config.DB)
		report, err := hwReportCRUD.GetLatest()
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve hardware report"})
			return
		}
		c.JSON(200, gin.H{"report": report})
	}
}

func GetHardwareReportByTimeRangeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Start string `json:"start"`
			End   string `json:"end"`
		}
		if err := c.BindJSON(&request); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}
		const layout = "2006-01-02 15:04:05.000"
		start, err := time.Parse(layout, request.Start)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid start time format. Use YYYY-MM-DD HH:MM:SS.mmm"})
			return
		}
		end, err := time.Parse(layout, request.End)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid end time format. Use YYYY-MM-DD HH:MM:SS.mmm"})
			return
		}
		hwReportCRUD := crud.NewHardWareReportCRUD(Config.DB)
		reports, err := hwReportCRUD.GetReportsByTimeRange(start, end)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve hardware reports"})
			return
		}
		c.JSON(200, gin.H{"reports": reports})
	}
}

func GetAverageHardwareUsageByTimeRangeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Start string `json:"start"`
			End   string `json:"end"`
		}
		if err := c.BindJSON(&request); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}
		const layout = "2006-01-02 15:04:05.000"
		start, err := time.Parse(layout, request.Start)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid start time format. Use YYYY-MM-DD HH:MM:SS.mmm"})
			return
		}
		end, err := time.Parse(layout, request.End)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid end time format. Use YYYY-MM-DD HH:MM:SS.mmm"})
			return
		}
		hwReportCRUD := crud.NewHardWareReportCRUD(Config.DB)
		avgCPU, avgMemory, avgDisk, err := hwReportCRUD.GetAverageUsageByTimeRange(start, end)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve average hardware usage"})
			return
		}
		fmt.Println(avgCPU, avgDisk, avgMemory)
		c.JSON(200, gin.H{
			"average_cpu_usage":    avgCPU,
			"average_memory_usage": avgMemory,
			"average_disk_usage":   avgDisk,
		})
	}
}
