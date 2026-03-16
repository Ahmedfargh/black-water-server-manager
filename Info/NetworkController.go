package info

import (
	"net/http"

	"github.com/ahmedfargh/server-manager/HardWare"
	"github.com/gin-gonic/gin"
)

func GetNetworkInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		networkService := HardWare.NewNetworkService()
		networkInfo, err := networkService.GetNetworkInfo()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, networkInfo)
	}
}
func GetNetworkConnections() gin.HandlerFunc {
	return func(c *gin.Context) {
		networkService := HardWare.NewNetworkService()
		networkConnections, err := networkService.GetNetworkConnections()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, networkConnections)
	}
}
