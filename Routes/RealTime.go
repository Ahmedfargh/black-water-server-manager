package routes

import (
	"net/http"
	"strings"

	"github.com/ahmedfargh/server-manager/WebSockets"
	"github.com/gin-gonic/gin"
)

// RegisterRealTimeRoutes registers all WebSocket endpoints with the Gin router
func RegisterRealTimeRoutes(router *gin.Engine) {
	router.GET("/ws/processes", gin.WrapH(http.HandlerFunc(ProcessRealTimeHandler)))
	router.GET("/ws/cpu-temperature", gin.WrapH(http.HandlerFunc(CpuTemperatureRealTimeHandler)))
	router.GET("/ws/docker/:containerId", gin.WrapH(http.HandlerFunc(DockerRealTimeHandler)))
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
	// Extract containerId from URL path or query
	// In Gin, router.GET("/ws/docker/:containerId", ...) doesn't directly pass param to http.HandlerFunc easily with gin.WrapH
	// However, we can use r.Context() or just parse the path if we know the pattern.
	// Alternatively, we can use query param if simpler, but let's try to get it from path.
	
	// A simpler way with gin.WrapH is to use a closure or change the handler to gin.HandlerFunc
	// But let's assume we can get it from the query for now if path is tricky, 
	// or better, let's fix the route registration.
	
	containerId := r.URL.Query().Get("containerId")
	if containerId == "" {
		// Fallback to path parsing if not in query
		// path is /ws/docker/ID
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 0 {
			containerId = parts[len(parts)-1]
		}
	}

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
