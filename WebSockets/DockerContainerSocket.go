package WebSockets

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ahmedfargh/server-manager/Services"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var DockerUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type DockerWebSocketHub struct {
	ContainerId string
	Clients     map[*DockerWebSocketClient]bool
	Broadcast   chan []byte
	Register    chan *DockerWebSocketClient
	Unregister  chan *DockerWebSocketClient
	mu          sync.RWMutex
}

var (
	hubs   = make(map[string]*DockerWebSocketHub)
	hubsMu sync.Mutex
)

// GetDockerHub returns a hub for a specific containerId, creating and starting it if necessary.
func GetDockerHub(containerId string) *DockerWebSocketHub {
	hubsMu.Lock()
	defer hubsMu.Unlock()

	if hub, ok := hubs[containerId]; ok {
		return hub
	}

	hub := &DockerWebSocketHub{
		ContainerId: containerId,
		Clients:     make(map[*DockerWebSocketClient]bool),
		Broadcast:   make(chan []byte),
		Register:    make(chan *DockerWebSocketClient),
		Unregister:  make(chan *DockerWebSocketClient),
	}
	hubs[containerId] = hub
	go hub.Run()
	return hub
}

type DockerWebSocketClient struct {
	Hub         *DockerWebSocketHub
	Conn        *websocket.Conn
	Send        chan []byte
	ContainerId string
}

func (hub *DockerWebSocketHub) Run() {
	go hub.startMonitoring()
	
	// Ensure hub is removed from registry when it stops
	defer func() {
		hubsMu.Lock()
		delete(hubs, hub.ContainerId)
		hubsMu.Unlock()
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
			// If no more clients, exit Run loop (stopping the hub)
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

func (hub *DockerWebSocketHub) startMonitoring() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hub.mu.RLock()
			clientCount := len(hub.Clients)
			hub.mu.RUnlock()

			// Stop monitoring if there are no clients
			if clientCount == 0 {
				return
			}

			if err := hub.WritePumpMessage(); err != nil {
				log.Printf("error broadcasting docker message for %s: %v", hub.ContainerId, err)
			}
		}
	}
}

func (hub *DockerWebSocketHub) WritePumpMessage() error {
	dockerService := &Services.DockerService{}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	// Fetch status only for this specific container
	status, err := dockerService.ContainerStatus(ctx, hub.ContainerId)
	if err != nil {
		return err
	}

	message, err := json.Marshal(status)
	if err != nil {
		return err
	}

	hub.Broadcast <- message
	return nil
}

func (c *DockerWebSocketClient) readPump() {
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
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
	}
}

func (c *DockerWebSocketClient) writePump() {
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

			// Add queued chat messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

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

func (hub *DockerWebSocketHub) Connect(w http.ResponseWriter, r *http.Request) error {
	conn, err := DockerUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	client := &DockerWebSocketClient{
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

func (hub *DockerWebSocketHub) Disconnect() error {
	return nil
}

func (hub *DockerWebSocketHub) BroadcastMessage(message []byte) {
	hub.Broadcast <- message
}

func (hub *DockerWebSocketHub) SendMessageToClient(conn *websocket.Conn, message []byte) error {
	return conn.WriteMessage(websocket.TextMessage, message)
}

func NewDockerWebSocketHub() *DockerWebSocketHub {
	return GetDockerHub("default")
}
