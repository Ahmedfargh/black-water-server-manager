package info

import (
	"fmt"
	"strconv"
	"strings"

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
func GetProcessByPID() gin.HandlerFunc {
	return func(c *gin.Context) {
		pid, err := strconv.Atoi(c.Param("pid"))
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		info, err := processes.GetProcessByPID(int32(pid))
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, info)
	}
}

func GetProcessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

		if page < 1 {
			page = 1
		}
		if pageSize < 1 {
			pageSize = 10
		}

		data, total, err := processes.GetPaginatedProcesses(page, pageSize)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"data":      data,
			"total":     total,
			"page":      page,
			"pageSize":  pageSize,
			"last_page": (total + int64(pageSize) - 1) / int64(pageSize),
		})
	}
}

func StartProcess() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Command string `json:"command"`
			Args    string `json:"args"`
		}
		if err := c.BindJSON(&request); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		parts := strings.Fields(request.Command)
		if len(parts) == 0 {
			c.JSON(400, gin.H{"error": "Command cannot be empty"})
			return
		}

		cmd := parts[0]
		// Use explicit args if provided, otherwise use parts[1:]
		var args []string
		if request.Args != "" {
			args = strings.Fields(request.Args)
		} else {
			args = parts[1:]
		}

		if _, err := processes.StartProcess(cmd, args...); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"message": "Process started"})
	}
}
