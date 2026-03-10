package WebSockets

import "net/http"

type WebSocketChannel interface {
	Connect(w http.ResponseWriter, r *http.Request) error
	Disconnect() error
	Send(data interface{}) error
	Receive() (<-chan []byte, error)
}
