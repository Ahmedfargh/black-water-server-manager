package functionalscontrollers

import (
	"net/http"

	"github.com/ahmedfargh/server-manager/Services"
	"github.com/gin-gonic/gin"
)

func GetContainersHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		dockerService := Services.NewDockerService()
		containers, err := dockerService.GetContainers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, containers)
	}
}
