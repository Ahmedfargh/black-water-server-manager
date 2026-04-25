package WebSockets

import (
	"context"
	"encoding/json"
	"fmt"
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

var DockerStateSocket DockerContainerStateSocket = DockerContainerStateSocket{}

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
			dss.mu.RLock()
			_, err := dss.Clients[client.ContainerId]
			dss.mu.RUnlock()
			if err {
				dss.Clients[client.ContainerId] = make([]*DockerContainerStateSocketClient, 5)
			} else {
				dss.Clients[client.ContainerId] = append(dss.Clients[client.ContainerId], &client)
			}
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
	timer := time.NewTimer(time.Duration(5) * time.Second)
	dockerServices, err := DockerService.NewDockerService()
	fmt.Println(err)
	for {
		select {
		case <-timer.C:
			for container_id, conns := range dss.Clients {
				ctx := context.Background()
				fmt.Println(len(conns))
				container_status, err := dockerServices.ContainerStatus(ctx, container_id)
				if err != nil {
					continue
				}
				for k, client := range dss.Clients[container_id] {
					data, err := json.Marshal(container_status)
					if err != nil {
						continue
					}
					client.Send <- data
					fmt.Println(k)
				}
			}
		}
	}
}
