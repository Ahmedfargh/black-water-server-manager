package routes

import (
	"net/http"

	"github.com/ahmedfargh/server-manager/WebSockets"
	"github.com/gin-gonic/gin"
)

// RegisterRealTimeRoutes registers all WebSocket endpoints with the Gin router
func RegisterRealTimeRoutes(router *gin.Engine) {
	router.GET("/ws/processes", gin.WrapH(http.HandlerFunc(ProcessRealTimeHandler)))
	router.GET("/ws/cpu-temperature", gin.WrapH(http.HandlerFunc(CpuTemperatureRealTimeHandler)))
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
