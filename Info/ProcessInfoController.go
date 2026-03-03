package info

import (
	"fmt"

	processes "github.com/ahmedfargh/server-manager/Processes"
	"github.com/gin-gonic/gin"
)

func GetProcessInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		info, err := processes.GetProcesses()
		fmt.Println(info)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})

		}
		c.JSON(200, info)
	}
}
