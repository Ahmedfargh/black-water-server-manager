package WebSockets

import (
	"bufio"
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ahmedfargh/server-manager/Services"
	"github.com/gorilla/websocket"
)

type DockerContainerLogHub struct {
	ContainerId string
	Clients     map[*DockerLogClient]bool
	Broadcast   chan []byte
	Register    chan *DockerLogClient
	Unregister  chan *DockerLogClient
	mu          sync.RWMutex
}

type DockerLogClient struct {
	Hub         *DockerContainerLogHub
	Conn        *websocket.Conn
	Send        chan []byte
	ContainerId string
}

var (
	logHubs   = make(map[string]*DockerContainerLogHub)
	logHubsMu sync.Mutex
)

func GetDockerLogHub(containerId string) *DockerContainerLogHub {
	logHubsMu.Lock()
	defer logHubsMu.Unlock()

	if hub, ok := logHubs[containerId]; ok {
		return hub
	}

	hub := &DockerContainerLogHub{
		ContainerId: containerId,
		Clients:     make(map[*DockerLogClient]bool),
		Broadcast:   make(chan []byte),
		Register:    make(chan *DockerLogClient),
		Unregister:  make(chan *DockerLogClient),
	}
	logHubs[containerId] = hub
	go hub.Run()
	return hub
}

func (hub *DockerContainerLogHub) Run() {
	// Start a goroutine to stream logs for this hub
	ctx, cancel := context.WithCancel(context.Background())
	go hub.streamLogs(ctx)

	defer func() {
		cancel()
		logHubsMu.Lock()
		delete(logHubs, hub.ContainerId)
		logHubsMu.Unlock()
	}()

	for {
		select {
		case client := <-hub.Register:
			hub.mu.Lock()
			hub.Clients[client] = true
			hub.mu.Unlock()
		case client := <-hub.Unregister:
			hub.mu.Lock()
			if _, ok := hub.Clients[client]; ok {
				delete(hub.Clients, client)
				close(client.Send)
			}
			if len(hub.Clients) == 0 {
				hub.mu.Unlock()
				return
			}
			hub.mu.Unlock()
		case message := <-hub.Broadcast:
			hub.mu.RLock()
			for client := range hub.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(hub.Clients, client)
				}
			}
			hub.mu.RUnlock()
		}
	}
}

func (hub *DockerContainerLogHub) streamLogs(ctx context.Context) {
	dockerService, _ := Services.NewDockerService()
	logs, err := dockerService.ContainerLogs(ctx, hub.ContainerId)
	if err != nil {
		log.Printf("error starting log stream for %s: %v", hub.ContainerId, err)
		return
	}
	defer logs.Close()

	scanner := bufio.NewScanner(logs)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			hub.Broadcast <- scanner.Bytes()
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("log stream error for %s: %v", hub.ContainerId, err)
	}
}

func (hub *DockerContainerLogHub) Connect(w http.ResponseWriter, r *http.Request) error {
	conn, err := DockerUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	client := &DockerLogClient{
		Hub:         hub,
		Conn:        conn,
		Send:        make(chan []byte, 256),
		ContainerId: hub.ContainerId,
	}
	hub.Register <- client

	go client.writePump()
	go client.readPump()

	return nil
}

func (c *DockerLogClient) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *DockerLogClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
