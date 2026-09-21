package routes

import (
	"net/http"
	"strings"

	MiddleWare "github.com/ahmedfargh/server-manager/Authentication"
	config "github.com/ahmedfargh/server-manager/Config"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	"github.com/ahmedfargh/server-manager/WebSockets"
	"github.com/gin-gonic/gin"
)

// RegisterRealTimeRoutes registers all WebSocket endpoints with the Gin router
func RegisterRealTimeRoutes(router *gin.Engine) {
	router.GET("/ws/processes", MiddleWare.AuthMiddleware(), gin.WrapH(http.HandlerFunc(ProcessRealTimeHandler)))
	router.GET("/ws/cpu-temperature", MiddleWare.AuthMiddleware(), gin.WrapH(http.HandlerFunc(CpuTemperatureRealTimeHandler)))
	router.GET("/ws/docker/:containerId", MiddleWare.AuthMiddleware(), gin.WrapH(http.HandlerFunc(DockerRealTimeHandler)))
	router.GET("/ws/docker/:containerId/logs", MiddleWare.AuthMiddleware(), gin.WrapH(http.HandlerFunc(DockerRealTimeLogsHandler)))
	router.GET("/ws/systemd/:unit/logs", MiddleWare.AuthMiddleware(), gin.WrapH(http.HandlerFunc(SystemdRealTimeLogsHandler)))
	router.GET("/ws/terminal", MiddleWare.AuthMiddleware(), TerminalRealTimeHandler)
	router.GET("/ws/ssh/terminal", MiddleWare.AuthMiddleware(), SSHTerminalRealTimeHandler)
	router.GET("/ws/:container_id/status", MiddleWare.AuthMiddleware(), gin.WrapH(http.HandlerFunc(DockerStatusHandler)))
	router.GET("/ws/file-system", MiddleWare.AuthMiddleware(), gin.WrapH(http.HandlerFunc(FileSystemRealTimeHandler)))
}

// ProcessRealTimeHandler handles process-specific WebSocket connections
func ProcessRealTimeHandler(w http.ResponseWriter, r *http.Request) {
	channel := WebSockets.NewChannel()
	err := channel.Connect(w, r)
	if err != nil {
		// Connection failed, send an error response if it hasn't been hijacked yet
		http.Error(w, "Failed to connect to WebSocket", http.StatusInternalServerError)
		return
	}
	defer channel.Disconnect()
}
func CpuTemperatureRealTimeHandler(w http.ResponseWriter, r *http.Request) {
	channel := WebSockets.NewCpuChannel()
	err := channel.Connect(w, r)
	if err != nil {
		// Connection failed, send an error response if it hasn't been hijacked yet
		http.Error(w, "Failed to connect to WebSocket", http.StatusInternalServerError)
		return
	}
	defer channel.Disconnect()
}

func DockerRealTimeHandler(w http.ResponseWriter, r *http.Request) {
	containerId := getContainerId(r)
	if containerId == "" {
		http.Error(w, "containerId is required", http.StatusBadRequest)
		return
	}
	hub := WebSockets.GetDockerHub(containerId)
	err := hub.Connect(w, r)
	if err != nil {
		http.Error(w, "Failed to connect to WebSocket", http.StatusInternalServerError)
		return
	}
}

func DockerRealTimeLogsHandler(w http.ResponseWriter, r *http.Request) {
	containerId := getContainerId(r)
	if containerId == "" {
		http.Error(w, "containerId is required", http.StatusBadRequest)
		return
	}
	hub := WebSockets.GetDockerLogHub(containerId)
	err := hub.Connect(w, r)
	if err != nil {
		http.Error(w, "Failed to connect to WebSocket", http.StatusInternalServerError)
		return
	}
}

func SystemdRealTimeLogsHandler(w http.ResponseWriter, r *http.Request) {
	unitName := getUnitName(r)
	if unitName == "" {
		http.Error(w, "unit name is required", http.StatusBadRequest)
		return
	}
	hub := WebSockets.GetSystemdLogHub(unitName)
	err := hub.Connect(w, r)
	if err != nil {
		http.Error(w, "Failed to connect to WebSocket", http.StatusInternalServerError)
		return
	}
}

func getUnitName(r *http.Request) string {
	unit := r.URL.Query().Get("unit")
	if unit == "" {
		parts := strings.Split(r.URL.Path, "/")
		for i, part := range parts {
			if part == "systemd" && i+1 < len(parts) {
				unit = parts[i+1]
				break
			}
		}
	}
	return unit
}

func getContainerId(r *http.Request) string {
	containerId := r.URL.Query().Get("containerId")
	if containerId == "" {
		parts := strings.Split(r.URL.Path, "/")
		for i, part := range parts {
			if part == "docker" && i+1 < len(parts) {
				containerId = parts[i+1]
				break
			}
		}
	}
	return containerId
}

func TerminalRealTimeHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: User ID not found in context"})
		return
	}

	user_id, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID type"})
		return
	}

	// Verify terminal authorization
	var user models.User
	if err := config.DB.Preload("Role").Preload("Role.Permissions").Preload("Permissions").First(&user, user_id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if !user.HasPermission("terminal_access") && user.Role.Name != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: user lacks terminal_access permission"})
		return
	}

	conn, err := WebSockets.DockerUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	clientIP := c.ClientIP()
	WebSockets.TerminalPool.ConnectSession(int32(user_id), clientIP, user.Username, conn)
}

func DockerStatusHandler(w http.ResponseWriter, r *http.Request) {
	container_id := getContainerId(r)
	WebSockets.DockerStateSocket.Connect(w, r, container_id)
}
func FileSystemRealTimeHandler(w http.ResponseWriter, r *http.Request) {
	channel := WebSockets.NewFileChannel()
	err := channel.Connect(w, r)
	if err != nil {
		// Connection failed, send an error response if it hasn't been hijacked yet
		http.Error(w, "Failed to connect to WebSocket", http.StatusInternalServerError)
		return
	}
}

func SSHTerminalRealTimeHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: User ID not found in context"})
		return
	}

	user_id, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID type"})
		return
	}

	// Verify terminal authorization (ssh_terminal_access or super_admin)
	var user models.User
	if err := config.DB.Preload("Role").Preload("Role.Permissions").Preload("Permissions").First(&user, user_id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if !user.HasPermission("ssh_terminal_access") && !user.HasPermission("terminal_access") && user.Role.Name != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: user lacks ssh_terminal_access permission"})
		return
	}

	clientIP := c.ClientIP()
	WebSockets.HandleSSHTerminal(c.Writer, c.Request, user.ID, user.Username, clientIP)
}
