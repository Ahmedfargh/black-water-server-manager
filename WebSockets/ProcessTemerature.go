package WebSockets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/host"
)

type CpuTemperature struct {
	Conn        *websocket.Conn
	Temperature []host.TemperatureStat `json:"temperature"`
	mu          sync.Mutex
	Send        chan []byte
	Receive     chan []byte
	Stop        chan struct{}
}

func NewCpuChannel() *CpuTemperature {
	return &CpuTemperature{
		Send:    make(chan []byte, 100),
		Receive: make(chan []byte, 100),
		Stop:    make(chan struct{}),
	}
}

func (c *CpuTemperature) Connect(w http.ResponseWriter, r *http.Request) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	c.Conn = conn

	done := make(chan struct{})

	go c.StartMonitoring()

	go c.WritePump(done)

	<-done
	return nil
}

func (c *CpuTemperature) StartMonitoring() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Stop:
			return
		case <-ticker.C:
			temperatures, err := host.SensorsTemperatures()
			// core_sensors := []host.TemperatureStat{}
			// if err != nil {
			// 	continue
			// }
			// for _, temp := range temperatures {
			// 	key := strings.ToLower(temp.SensorKey)
			// 	if strings.Contains(key, "coretemp") {
			// 		core_sensors = append(core_sensors, temp)
			// 	} else {
			// 		fmt.Printf("Skipping non-core temperature sensor: %s\n", temp.SensorKey)
			// 		continue
			// 	}
			// }
			c.mu.Lock()
			c.Temperature = temperatures
			c.mu.Unlock()

			data, err := json.Marshal(c.Temperature)
			if err != nil {
				continue
			}
			c.Send <- data
		}
	}
}

func (c *CpuTemperature) WritePump(done chan struct{}) {
	defer func() {
		c.Conn.Close()
		close(done)
	}()

	for {
		select {
		case message, ok := <-c.Send:
			fmt.Println()
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				return
			}
		case <-c.Stop:
			return
		}
	}
}
func (c *CpuTemperature) ReadPump() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Stop <- struct{}{}
			}
			return
		}
		c.Receive <- message
	}
}
func (c *CpuTemperature) send(data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	select {
	case c.Send <- bytes:
	default:
		// Channel full, skip update to stay async
	}
	return nil
}
func (c *CpuTemperature) Disconnect() error {
	select {
	case <-c.Stop:
		return nil
	default:
		close(c.Stop)
	}
	return nil
}
