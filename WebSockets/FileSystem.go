package WebSockets

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	Mgr "github.com/ahmedfargh/server-manager/Managers"
	"github.com/gorilla/websocket"
)

type FileChannelClient struct {
	Client *websocket.Conn
	Send   chan []byte
}
type FileChannel struct {
	Clients    []*FileChannelClient
	Register   chan *FileChannelClient
	Mu         sync.RWMutex
	UnRegister chan *FileChannelClient
}

var FileChannelInstance = NewFileChannel()

func init() {
	go FileChannelInstance.RegisterClient()
	go FileChannelInstance.UnRegisterClient()
	// go FileChannelInstance.WritePump()
}
func NewFileChannel() *FileChannel {
	file_channel := &FileChannel{
		Clients:    make([]*FileChannelClient, 0),
		Register:   make(chan *FileChannelClient, 10),
		UnRegister: make(chan *FileChannelClient, 10),
		Mu:         sync.RWMutex{},
	}
	go file_channel.RegisterClient()
	go file_channel.UnRegisterClient()
	return file_channel
}
func (fc *FileChannelClient) ReadFromClient() {
	for {
		_, message, err := fc.Client.ReadMessage()
		if err != nil {
			return
		}
		var message_body map[string]interface{}
		err = json.Unmarshal(message, &message_body)
		if err != nil {
			fc.Send <- []byte("Invalid message format")
			continue
		}
		if message_body["action"] == "disconnect" {
			FileChannelInstance.UnRegister <- fc
			return
		}
		if message_body["action"] == "copy" {
			file_mgr := Mgr.FileManager{}
			var channel = make(chan Mgr.FileSystemEvent)
			go file_mgr.CopyDirectory(message_body["src"].(string), message_body["dst"].(string), channel)
			fault := 0
			for {
				select {
				case data := <-channel:
					json_data, err := json.Marshal(data)
					if err != nil {
						continue
					}
					fc.Send <- json_data
					fault = 0

				default:
					fc.Send <- []byte("no signal from copying routine")
					fault++
					if fault > 5 {
						return
					}
					time.Sleep(1 * time.Second)

				}
			}
		}
	}
}
func (fc *FileChannel) Connect(w http.ResponseWriter, r *http.Request) error {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	client := &FileChannelClient{
		Client: conn,
		Send:   make(chan []byte, 10),
	}
	fc.Register <- client
	go client.SendToClient()
	go client.ReadFromClient()
	return nil
}
func (fc *FileChannel) WritePump() {
	for {
		fc.Mu.RLock()
		for _, client := range fc.Clients {
			select {
			case message := <-client.Send:
				message_body, err := client.Client.NextWriter(websocket.TextMessage)
				if err != nil {
					continue
				}

				message_body.Write(message)
				if err := message_body.Close(); err != nil {
					continue
				}
			default:
			}
		}
		fc.Mu.RUnlock()
	}
}
func (client *FileChannelClient) SendToClient() {
	for {
		select {
		case message := <-client.Send:
			message_body, err := client.Client.NextWriter(websocket.TextMessage)
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
func (fc *FileChannel) RegisterClient() {
	for {
		select {
		case client := <-fc.Register:
			fc.Mu.Lock()
			fc.Clients = append(fc.Clients, client)
			fc.Mu.Unlock()
		}
	}
}
func (fc *FileChannel) UnRegisterClient() {
	for {
		select {
		case client := <-fc.UnRegister:
			fc.Mu.Lock()
			for i, c := range fc.Clients {
				if c == client {
					fc.Clients = append(fc.Clients[:i], fc.Clients[i+1:]...)
					break
				}
			}
			fc.Mu.Unlock()
		}
	}
}
