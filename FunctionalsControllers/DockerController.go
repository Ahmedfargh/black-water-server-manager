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
