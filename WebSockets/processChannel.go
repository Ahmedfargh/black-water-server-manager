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

type Hub struct {
}
type ProcessChannel struct {
	Conn    *websocket.Conn
	mu      sync.Mutex
	send    chan []byte
	receive chan []byte
	stop    chan struct{}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewChannel() *ProcessChannel {
	return &ProcessChannel{
		send:    make(chan []byte, 100),
		receive: make(chan []byte, 100),
		stop:    make(chan struct{}),
	}
}
func (p *ProcessChannel) Connect(w http.ResponseWriter, r *http.Request) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	p.Conn = conn

	// Create the signaling channel
	done := make(chan struct{})

	// Start process monitoring
	go p.StartMonitoring()

	// Pass 'done' to the pumps so they can signal when they exit
	go p.WritePump(done)
	// go p.ReadPump(done)

	// This now waits correctly until ReadPump or WritePump finishes
	<-done
	log.Println("Connection closed gracefully")
	return nil
}

func (p *ProcessChannel) WritePump(done chan struct{}) {
	defer func() {
		p.Disconnect()
		close(done) // This triggers the <-done in Connect
	}()
	for {
		select {
		case message := <-p.send:
			// Gorilla requires a lock if multiple goroutines write,
			// but since only WritePump writes, we are safe here.
			err := p.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				return
			}
		case <-p.stop:
			return
		}
	}
}
func (p *ProcessChannel) Disconnect() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	select {
	case <-p.stop:
		// already closed
		return nil
	default:
		close(p.stop) // Signal all goroutines to stop
	}

	if p.Conn != nil {
		err := p.Conn.Close()
		p.Conn = nil
		return err
	}
	return nil
}
func (p *ProcessChannel) MonitorConnection() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.mu.Lock()
			conn := p.Conn
			p.mu.Unlock()
			if conn == nil {
				return
			}
			err := conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				p.Disconnect()
				return
			}
		case <-p.stop:
			return
		}
	}
}
func (p *ProcessChannel) Send(data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	select {
	case p.send <- bytes:
	default:
		// Channel full, skip update to stay async
		log.Println("Send buffer full, skipping update")
	}
	return nil
}

func (p *ProcessChannel) Receive() (<-chan []byte, error) {
	go func() {
		for {
			_, message, err := p.Conn.ReadMessage()
			if err != nil {
				p.Disconnect()
				return
			}
			select {
			case p.receive <- message:
			default:
				log.Println("Receive buffer full, skipping message")
			}
		}
	}()
	return p.receive, nil
}

func (p *ProcessChannel) StartMonitoring() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Initial fetch and send
	p.updateProcesses()

	for {
		select {
		case <-ticker.C:
			p.updateProcesses()
		case <-p.stop:
			return
		}
	}
}

func (p *ProcessChannel) updateProcesses() {
	processList, err := processes.GetProcesses()
	if err != nil {
		log.Println("Error getting processes:", err)
		return
	}
	p.Send(processList)
}
