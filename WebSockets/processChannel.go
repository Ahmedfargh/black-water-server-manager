package WebSockets

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	processes "github.com/ahmedfargh/server-manager/Processes"
	"github.com/gorilla/websocket"
)

// Client represents a single WebSocket connection
type Client struct {
	hub  *ProcessHub
	conn *websocket.Conn
	send chan []byte
}

// ProcessHub manages multiple WebSocket clients and broadcasts process updates
type ProcessHub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

var (
	globalProcessHub *ProcessHub
	hubOnce          sync.Once
	upgrader         = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

// GetProcessHub returns the singleton instance of ProcessHub
func GetProcessHub() *ProcessHub {
	hubOnce.Do(func() {
		globalProcessHub = &ProcessHub{
			clients:    make(map[*Client]bool),
			broadcast:  make(chan []byte),
			register:   make(chan *Client),
			unregister: make(chan *Client),
		}
		go globalProcessHub.run()
		go globalProcessHub.startMonitoring()
	})
	return globalProcessHub
}

func (h *ProcessHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client registered: %s", client.conn.RemoteAddr())

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client unregistered: %s", client.conn.RemoteAddr())
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// If the client's send channel is full, they are likely too slow
					// or disconnected, so we unregister them.
					go func(c *Client) { h.unregister <- c }(client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *ProcessHub) startMonitoring() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Initial fetch
	h.updateProcesses()

	for {
		select {
		case <-ticker.C:
			h.updateProcesses()
		}
	}
}

func (h *ProcessHub) updateProcesses() {
	processList, err := processes.GetProcesses()
	if err != nil {
		log.Println("Error getting processes:", err)
		return
	}

	bytes, err := json.Marshal(processList)
	if err != nil {
		log.Println("Error marshaling processes:", err)
		return
	}

	h.broadcast <- bytes
}

func (h *ProcessHub) Connect(w http.ResponseWriter, r *http.Request) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}

	h.register <- client

	go client.writePump()
	go client.readPump()

	return nil
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// The hub closed the channel.
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

type ProcessChannel struct {
	hub *ProcessHub
}

func NewChannel() *ProcessChannel {
	return &ProcessChannel{
		hub: GetProcessHub(),
	}
}

func (p *ProcessChannel) Connect(w http.ResponseWriter, r *http.Request) error {
	return p.hub.Connect(w, r)
}

func (p *ProcessChannel) Disconnect() error {
	// Individual client disconnection is handled via readPump/writePump
	return nil
}

func (p *ProcessChannel) Send(data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	p.hub.broadcast <- bytes
	return nil
}

func (p *ProcessChannel) Receive() (<-chan []byte, error) {
	// Not really used in broadcasting hub pattern for now
	return nil, nil
}
