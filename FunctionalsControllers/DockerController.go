package functionalscontrollers

import (
	"context"
	"net/http"

	"github.com/ahmedfargh/server-manager/Services"
	"github.com/gin-gonic/gin"
)

func GetContainersHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		dockerService, err := Services.NewDockerService()
		containers, err := dockerService.GetContainers(context.Background())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, containers)
	}
}

func GetContainerByIDHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		dockerService, err := Services.NewDockerService()
		container, err := dockerService.GetContainerByID(context.Background(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, container)
	}
}

func ContainerStatusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		dockerService, err := Services.NewDockerService()
		container, err := dockerService.ContainerStatus(context.Background(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, container)
	}
}
func ActionContainerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		dockerService, err := Services.NewDockerService()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		action := c.Param("action")
		if action != "start" && action != "stop" && action != "restart" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
			return
		} else if action == "start" {
			err = dockerService.StartContainer(context.Background(), id)
		} else if action == "stop" {
			err = dockerService.StopContainer(context.Background(), id)
		} else if action == "restart" {
			err = dockerService.RestartContainer(context.Background(), id)
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Container " + action + "ed successfully"})
	}
}
