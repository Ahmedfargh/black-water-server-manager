package WebSockets

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	DockerService "github.com/ahmedfargh/server-manager/Services"
	"github.com/gorilla/websocket"
)

type DockerContainerStateSocketClient struct {
	Client      *websocket.Conn
	ContainerId string
	Send        chan []byte
}
type DockerContainerStateSocket struct {
	Clients  map[string][]*DockerContainerStateSocketClient
	mu       sync.RWMutex
	Register chan DockerContainerStateSocketClient
}

var DockerStateSocket = DockerContainerStateSocket{
	Clients:  make(map[string][]*DockerContainerStateSocketClient),
	Register: make(chan DockerContainerStateSocketClient, 10),
}

func init() {
	go DockerStateSocket.RegisterClient()
	go DockerStateSocket.WritePump()
}

func GetDocketContainerState() DockerContainerStateSocket {
	return DockerStateSocket
}
func (dc *DockerContainerStateSocketClient) SendToClient() {
	for {
		select {
		case message := <-dc.Send:
			message_body, err := dc.Client.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			message_body.Write(message)
			if err := message_body.Close(); err != nil {
				return
			}
		}
	}
}
func (dss *DockerContainerStateSocket) LunchGoRoutines() {
	go dss.WritePump()
}
func (dss *DockerContainerStateSocket) RegisterClient() {
	for {
		select {
		case client := <-dss.Register:
			dss.mu.Lock()
			if _, exists := dss.Clients[client.ContainerId]; !exists {
				dss.Clients[client.ContainerId] = make([]*DockerContainerStateSocketClient, 0)
			}
			client.Send = make(chan []byte, 256)
			dss.Clients[client.ContainerId] = append(dss.Clients[client.ContainerId], &client)
			go client.SendToClient()
			dss.mu.Unlock()
		}
	}
}
func (dss *DockerContainerStateSocket) Connect(w http.ResponseWriter, r *http.Request, container_id string) error {
	conn, err := DockerUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	client := DockerContainerStateSocketClient{
		Client:      conn,
		ContainerId: container_id,
	}
	dss.Register <- client
	return nil
}
func (dss *DockerContainerStateSocket) WritePump() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	dockerServices, _ := DockerService.NewDockerService()
	
	for {
		select {
		case <-ticker.C:
			dss.mu.RLock()
			for container_id, conns := range dss.Clients {
				if len(conns) == 0 {
					continue
				}
				ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
				container_status, err := dockerServices.ContainerStatus(ctx, container_id)
				cancel()
				
				if err != nil {
					continue
				}
				
				data, err := json.Marshal(container_status)
				if err != nil {
					continue
				}
				
				for _, client := range conns {
					select {
					case client.Send <- data:
					default:
						// Client buffer full or disconnected
					}
				}
			}
			dss.mu.RUnlock()
		}
	}
}
