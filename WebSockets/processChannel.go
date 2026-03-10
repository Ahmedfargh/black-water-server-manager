package WebSockets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	processes "github.com/ahmedfargh/server-manager/Processes"

	"github.com/gorilla/websocket"
)

type ProcessChannel struct {
	Conn    *websocket.Conn
	mu      sync.Mutex
	message chan []byte
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewChannel() *ProcessChannel {
	return &ProcessChannel{
		message: make(chan []byte),
	}
}

func (p *ProcessChannel) Connect(w http.ResponseWriter, r *http.Request) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	p.Conn = conn

	// We block here to keep the connection alive.
	// The handler in Routes will trigger Disconnect when this returns.
	for {
		// Placeholder for actual process data
		data := map[string]interface{}{"status": "alive", "time": time.Now().Format(time.RFC3339)}
		process, err := processes.GetProcesses()
		data["processes"] = process
		processes, err := json.Marshal(data)
		if err != nil {
			return err
		}

		p.mu.Lock()
		if p.Conn == nil {
			p.mu.Unlock()
			return nil
		}
		err = p.Conn.WriteMessage(websocket.TextMessage, processes)
		p.mu.Unlock()

		if err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
	}
}

func (p *ProcessChannel) Disconnect() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Conn != nil {
		err := p.Conn.Close()
		p.Conn = nil
		return err
	}
	return nil
}

func (p *ProcessChannel) Send(data interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Conn == nil {
		return fmt.Errorf("connection not established")
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return p.Conn.WriteMessage(websocket.TextMessage, bytes)
}

func (p *ProcessChannel) Receive() (<-chan []byte, error) {
	if p.Conn == nil {
		return nil, fmt.Errorf("connection not established")
	}

	go func() {
		for {
			p.mu.Lock()
			conn := p.Conn
			p.mu.Unlock()
			if conn == nil {
				return
			}

			_, msg, err := conn.ReadMessage()
			if err != nil {
				close(p.message)
				return
			}
			p.message <- msg
		}
	}()

	return p.message, nil
}
