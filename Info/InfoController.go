package info

import (
	HardWare "github.com/ahmedfargh/server-manager/HardWare"
	"github.com/gin-gonic/gin"
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
