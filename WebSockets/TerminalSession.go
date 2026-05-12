package WebSockets

import (
	"os/exec"

	Config "github.com/ahmedfargh/server-manager/Config"
	crud_service "github.com/ahmedfargh/server-manager/Database/CRUD"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	Repository "github.com/ahmedfargh/server-manager/Database/Repository"
	websocket "github.com/gorilla/websocket"
	json "github.com/json-iterator/go"
)

type TerminalSession struct {
	SessionID      int32  `json:"session_id"`
	Command        string `json:"command"`
	Conn           *websocket.Conn
	ExecuteCommand chan string
	SendResult     chan string
	audit_service  crud_service.AuditLogCRUD
}
type terminalPool struct {
	Sessions map[int32]*TerminalSession
}

var TerminalPool = &terminalPool{}

func NewTerminalSession() *TerminalSession {
	return &TerminalSession{
		SessionID:      0,
		Command:        "",
		Conn:           nil,
		ExecuteCommand: make(chan string),
		SendResult:     make(chan string),
		audit_service:  crud_service.AuditLogCRUD{Repo: Repository.NewAuditRepository(Config.DB)},
	}
}
func (ts *terminalPool) ConnectSession(sessionID int32, conn *websocket.Conn) {
	if ts.Sessions == nil {
		ts.Sessions = make(map[int32]*TerminalSession)
	}

	// Terminate any existing session to prevent zombie connections and ensure proper reuse
	if existing, exists := ts.Sessions[sessionID]; exists {
		if existing.Conn != nil {
			existing.Conn.Close()
		}
	}

	newSession := &TerminalSession{
		SessionID:      sessionID,
		Conn:           conn,
		ExecuteCommand: make(chan string),
		SendResult:     make(chan string),
	}
	ts.Sessions[sessionID] = newSession

	go newSession.WritePump()
	go newSession.ReadPump()
	go newSession.RunCommands()
}

func (ts *TerminalSession) WritePump() {
	for {
		result, ok := <-ts.SendResult
		if !ok {
			ts.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		if err := ts.Conn.WriteMessage(websocket.TextMessage, []byte(result)); err != nil {
			return
		}
	}
}

func (ts *TerminalSession) ReadPump() {
	defer ts.Close()
	defer close(ts.ExecuteCommand)
	for {
		_, message, err := ts.Conn.ReadMessage()
		if err != nil {
			return
		}
		ts.ExecuteCommand <- string(message)
	}
}

func (ts *TerminalSession) RunCommands() {
	defer close(ts.SendResult)
	for {
		cmd, ok := <-ts.ExecuteCommand
		if !ok {
			return
		}

		var command_structure map[string]any
		err := json.Unmarshal([]byte(cmd), &command_structure)
		if err != nil {
			ts.SendResult <- "Invalid command format"
			continue
		}
		command, ok := command_structure["command"].(string)
		if !ok {
			ts.SendResult <- "Command field is missing or not a string"
			continue
		}
		execCmd := exec.Command("sh", "-c", command)
		output, err := execCmd.CombinedOutput()
		result := string(output)
		user_id := uint(ts.SessionID)
		audit_log := models.AuditLog{
			UserID:      &user_id,
			ServiceType: "terminal_session",
			Action:      string(cmd),
			Results:     result,
		}
		go ts.audit_service.CreateAudit(&audit_log)
		if err != nil {
			result += "\nError: " + err.Error()
		}
		ts.SendResult <- result
	}
}

func (ts *TerminalSession) Close() {
	if ts.Conn != nil {
		ts.Conn.Close()
	}
}
