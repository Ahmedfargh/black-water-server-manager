package routes

import (
	"fmt"
	"net/http"
	"strings"

	MiddleWare "github.com/ahmedfargh/server-manager/Authentication"
	"github.com/ahmedfargh/server-manager/WebSockets"
	"github.com/gin-gonic/gin"
)

// RegisterRealTimeRoutes registers all WebSocket endpoints with the Gin router
func RegisterRealTimeRoutes(router *gin.Engine) {
	router.GET("/ws/processes", MiddleWare.AuthMiddleware(), gin.WrapH(http.HandlerFunc(ProcessRealTimeHandler)))
	router.GET("/ws/cpu-temperature", MiddleWare.AuthMiddleware(), gin.WrapH(http.HandlerFunc(CpuTemperatureRealTimeHandler)))
	router.GET("/ws/docker/:containerId", MiddleWare.AuthMiddleware(), gin.WrapH(http.HandlerFunc(DockerRealTimeHandler)))
	router.GET("/ws/docker/:containerId/logs", MiddleWare.AuthMiddleware(), gin.WrapH(http.HandlerFunc(DockerRealTimeLogsHandler)))
	router.GET("/ws/terminal", MiddleWare.AuthMiddleware(), TerminalRealTimeHandler)
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
	fmt.Println("TerminalRealTimeHandler called")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	user_id, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid User ID type"})
		return
	}

	fmt.Println("User ID from middleware:", user_id)

	conn, err := WebSockets.DockerUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// http.Error(w, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
		return
	}
	WebSockets.TerminalPool.ConnectSession(int32(user_id), conn)

}
