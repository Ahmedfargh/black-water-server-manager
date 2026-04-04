package info

import (
	reports "github.com/ahmedfargh/server-manager/BackGround/Reports"
	HardWare "github.com/ahmedfargh/server-manager/HardWare"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

func GetCputInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		info, err := HardWare.GetCpuInfo()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, info)
	}
}

func GetGpuInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		info, err := HardWare.GetGpuInfo()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, info)
	}
}

func GetRamInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		info, err := HardWare.GetRamInfo()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, info)
	}
}

func GetDiskInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		info, err := HardWare.GetDiskInfo()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, info)
	}
}
func GetHardWareReport() gin.HandlerFunc {
	return func(c *gin.Context) {
		cpuUsage, _ := cpu.Percent(0, false)
		memUsage, _ := mem.VirtualMemory()
		diskUsage, _ := disk.Usage("/")
		// Collect hardware report data
		report := reports.BackgroundHardwareReport{
			Report: reports.HardwareReport{
				CPUUsage:    cpuUsage[0],
				MemoryUsage: memUsage.UsedPercent,
				DiskUsage:   diskUsage.UsedPercent,
			},
		}
		c.JSON(200, report)

	}
}
