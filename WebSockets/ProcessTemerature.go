package WebSockets

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/host"
)

type TemperatureClient struct {
	hub  *TemperatureHub
	conn *websocket.Conn
	send chan []byte
}

type TemperatureHub struct {
	clients    map[*TemperatureClient]bool
	broadcast  chan []byte
	register   chan *TemperatureClient
	unregister chan *TemperatureClient
	mu         sync.RWMutex
}

var (
	globalTemperatureHub *TemperatureHub
	tempHubOnce          sync.Once
)

// GetTemperatureHub returns the singleton instance of TemperatureHub
func GetTemperatureHub() *TemperatureHub {
	tempHubOnce.Do(func() {
		globalTemperatureHub = &TemperatureHub{
			clients:    make(map[*TemperatureClient]bool),
			broadcast:  make(chan []byte),
			register:   make(chan *TemperatureClient),
			unregister: make(chan *TemperatureClient),
		}
		go globalTemperatureHub.run()
		go globalTemperatureHub.startMonitoring()
	})
	return globalTemperatureHub
}

func (h *TemperatureHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Temperature Client registered: %s", client.conn.RemoteAddr())

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Temperature Client unregistered: %s", client.conn.RemoteAddr())
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					go func(c *TemperatureClient) { h.unregister <- c }(client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *TemperatureHub) startMonitoring() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.updateTemperature()
		}
	}
}

func (h *TemperatureHub) updateTemperature() {
	temperatures, err := host.SensorsTemperatures()
	if err != nil {
		log.Println("Error getting temperatures:", err)
		return
	}

	bytes, err := json.Marshal(temperatures)
	if err != nil {
		log.Println("Error marshaling temperatures:", err)
		return
	}

	h.broadcast <- bytes
}

// Connect upgrades the HTTP connection and adds the client to the hub
func (h *TemperatureHub) Connect(w http.ResponseWriter, r *http.Request) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	client := &TemperatureClient{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}

	h.register <- client

	go client.writePump()
	go client.readPump()

	return nil
}

func (c *TemperatureClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *TemperatureClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				return
			}

		case <-ticker.C:
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// --- Compatibility Shims ---

type CpuTemperature struct {
	hub *TemperatureHub
}

func NewCpuChannel() *CpuTemperature {
	return &CpuTemperature{
		hub: GetTemperatureHub(),
	}
}

func (c *CpuTemperature) Connect(w http.ResponseWriter, r *http.Request) error {
	return c.hub.Connect(w, r)
}

func (c *CpuTemperature) Disconnect() error {
	return nil
}
