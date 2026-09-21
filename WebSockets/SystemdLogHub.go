package WebSockets

import (
	"bufio"
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	systemd "github.com/ahmedfargh/server-manager/Services/Systemd"
	"github.com/gorilla/websocket"
)

type SystemdLogHub struct {
	UnitName   string
	Clients    map[*SystemdLogClient]bool
	Broadcast  chan []byte
	Register   chan *SystemdLogClient
	Unregister chan *SystemdLogClient
	mu         sync.RWMutex
}

type SystemdLogClient struct {
	Hub      *SystemdLogHub
	Conn     *websocket.Conn
	Send     chan []byte
	UnitName string
}

var (
	systemdHubs   = make(map[string]*SystemdLogHub)
	systemdHubsMu sync.Mutex
)

func GetSystemdLogHub(unitName string) *SystemdLogHub {
	systemdHubsMu.Lock()
	defer systemdHubsMu.Unlock()

	if hub, ok := systemdHubs[unitName]; ok {
		return hub
	}

	hub := &SystemdLogHub{
		UnitName:   unitName,
		Clients:    make(map[*SystemdLogClient]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *SystemdLogClient),
		Unregister: make(chan *SystemdLogClient),
	}
	systemdHubs[unitName] = hub
	go hub.Run()
	return hub
}

func (hub *SystemdLogHub) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	go hub.streamJournalLogs(ctx)

	defer func() {
		cancel()
		systemdHubsMu.Lock()
		delete(systemdHubs, hub.UnitName)
		systemdHubsMu.Unlock()
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

func (hub *SystemdLogHub) streamJournalLogs(ctx context.Context) {
	svc := systemd.NewService()
	logs, err := svc.StreamUnitLogs(ctx, hub.UnitName, 100)
	if err != nil {
		log.Printf("journalctl log stream error for %s: %v", hub.UnitName, err)
		hub.Broadcast <- []byte("[SYSTEM] Journalctl log stream unavailable: " + err.Error())
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
}

func (hub *SystemdLogHub) Connect(w http.ResponseWriter, r *http.Request) error {
	conn, err := DockerUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	client := &SystemdLogClient{
		Hub:      hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UnitName: hub.UnitName,
	}
	hub.Register <- client

	go client.writePump()
	go client.readPump()

	return nil
}

func (c *SystemdLogClient) readPump() {
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

func (c *SystemdLogClient) writePump() {
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
