package functionalscontrollers

import (
	"context"
	"net/http"
	"strconv"

	Config "github.com/ahmedfargh/server-manager/Config"
	crud "github.com/ahmedfargh/server-manager/Database/CRUD"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	repo "github.com/ahmedfargh/server-manager/Database/Repository"
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

func CreateDockerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var docker models.Docker
		if err := c.ShouldBindJSON(&docker); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		rep := repo.NewDockerRepository(Config.DB)
		dc := crud.NewDockerCrud(rep)
		if err := dc.CreateDocker(docker); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Docker record created"})
	}
}

func UpdateDockerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id64, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		rep := repo.NewDockerRepository(Config.DB)
		dc := crud.NewDockerCrud(rep)
		existing, err := dc.GetDockerByID(uint(id64))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "docker not found"})
			return
		}
		var payload models.Docker
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		payload.ID = existing.ID
		if err := dc.UpdateDocker(&payload); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Docker updated"})
	}
}

func DeleteDockerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id64, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		rep := repo.NewDockerRepository(Config.DB)
		dc := crud.NewDockerCrud(rep)
		existing, err := dc.GetDockerByID(uint(id64))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "docker not found"})
			return
		}
		if err := dc.DeleteDocker(existing); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Docker deleted"})
	}
}

func SetDockerLimitsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id64, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		var payload struct {
			MaxCPU    float64 `json:"max_cpu_consumation"`
			MaxMemory float64 `json:"max_memory_consumation"`
			OnMaxCPU  string  `json:"on_max_cpu_consumation"`
			OnMaxMem  string  `json:"on_max_memory_consumation"`
			Action    string  `json:"action"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if payload.MaxCPU < 0 || payload.MaxCPU > 100 || payload.MaxMemory < 0 || payload.MaxMemory > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "max values must be between 0 and 100"})
			return
		}
		validActions := map[string]bool{"restart": true, "start": true, "stop": true, "remove": true, "nothing": true}
		if payload.OnMaxCPU != "" && !validActions[payload.OnMaxCPU] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid on_max_cpu_consumation action"})
			return
		}
		if payload.OnMaxMem != "" && !validActions[payload.OnMaxMem] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid on_max_memory_consumation action"})
			return
		}
		if payload.Action != "restart" && payload.Action != "start" && payload.Action != "stop" && payload.Action != "remove" && payload.Action != "nothing" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action"})
			return
		}
		rep := repo.NewDockerRepository(Config.DB)
		dc := crud.NewDockerCrud(rep)
		existing, err := dc.GetDockerByID(uint(id64))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "docker not found"})
			return
		}
		existing.MaxCpuConsumation = payload.MaxCPU
		existing.MaxMemoryConsumation = payload.MaxMemory
		results, err := false, nil
		if payload.OnMaxCPU != "" {
			results, err = dc.AddEventAction(existing, "max_cpu_consumation", payload.Action, payload.MaxCPU)

		}
		if payload.OnMaxMem != "" {
			results, err = dc.AddEventAction(existing, "max_memory_consumation", payload.Action, payload.MaxMemory)
		}
		if err := dc.UpdateDocker(existing); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !results {

			c.JSON(http.StatusOK, gin.H{"message": "Docker limits updated"})
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no update"})
	}
}
