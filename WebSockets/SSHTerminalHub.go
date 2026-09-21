package WebSockets

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	Config "github.com/ahmedfargh/server-manager/Config"
	crud_service "github.com/ahmedfargh/server-manager/Database/CRUD"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	Repository "github.com/ahmedfargh/server-manager/Database/Repository"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

type SSHWSMessage struct {
	Type string `json:"type"` // "input", "resize", "ping"
	Data string `json:"data"` // text input
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

type RemoteSSHSession struct {
	SessionID     string
	UserID        uint
	Username      string
	ClientIP      string
	RemoteHost    string
	RemotePort    int
	RemoteUser    string
	Conn          *websocket.Conn
	SSHClient     *ssh.Client
	SSHSession    *ssh.Session
	StdinWriter   io.WriteCloser
	audit_service crud_service.AuditLogCRUD
	closeOnce     sync.Once
	isClosed      bool
	mu            sync.Mutex
}

type sshTerminalPool struct {
	mu       sync.RWMutex
	Sessions map[string]*RemoteSSHSession
}

var SSHTerminalPool = &sshTerminalPool{
	Sessions: make(map[string]*RemoteSSHSession),
}

func (p *sshTerminalPool) AddSession(session *RemoteSSHSession) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Sessions[session.SessionID] = session
}

func (p *sshTerminalPool) RemoveSession(sessionID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if session, exists := p.Sessions[sessionID]; exists {
		session.Close()
		delete(p.Sessions, sessionID)
	}
}

func (s *RemoteSSHSession) Close() {
	s.closeOnce.Do(func() {
		s.mu.Lock()
		s.isClosed = true
		s.mu.Unlock()

		if s.StdinWriter != nil {
			_ = s.StdinWriter.Close()
		}
		if s.SSHSession != nil {
			_ = s.SSHSession.Close()
		}
		if s.SSHClient != nil {
			_ = s.SSHClient.Close()
		}
		if s.Conn != nil {
			_ = s.Conn.Close()
		}

		// Log audit event
		auditModel := models.AuditLog{
			UserID:      &s.UserID,
			ServiceType: "SSH",
			ServiceID:   fmt.Sprintf("%s:%d", s.RemoteHost, s.RemotePort),
			Action:      "SSH_REMOTE_TERMINAL_DISCONNECT",
			Results:     fmt.Sprintf("Closed remote SSH session to %s@%s:%d (IP: %s)", s.RemoteUser, s.RemoteHost, s.RemotePort, s.ClientIP),
		}
		_ = s.audit_service.CreateAudit(&auditModel)
	})
}

// HandleSSHTerminal connects a WebSocket client to a remote SSH server
func HandleSSHTerminal(w http.ResponseWriter, r *http.Request, userID uint, username string, clientIP string) {
	// 1. Parse connection parameters from query or header
	host := r.URL.Query().Get("host")
	portStr := r.URL.Query().Get("port")
	remoteUser := r.URL.Query().Get("user")
	authType := r.URL.Query().Get("auth_type") // "password" or "key"
	password := r.URL.Query().Get("password")
	privateKey := r.URL.Query().Get("private_key")

	if host == "" {
		http.Error(w, "Query parameter 'host' is required", http.StatusBadRequest)
		return
	}
	if remoteUser == "" {
		remoteUser = "root"
	}
	port := 22
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil && p > 0 && p <= 65535 {
			port = p
		}
	}

	// 2. Prepare SSH Auth Methods
	var authMethods []ssh.AuthMethod
	if authType == "key" || privateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			http.Error(w, "Failed to parse SSH private key: "+err.Error(), http.StatusBadRequest)
			return
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	} else if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	} else {
		http.Error(w, "Either 'password' or 'private_key' must be provided", http.StatusBadRequest)
		return
	}

	sshConfig := &ssh.ClientConfig{
		User:            remoteUser,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // For dynamic client terminal connections
		Timeout:         15 * time.Second,
	}

	// 3. Dial SSH Server
	targetAddr := net.JoinHostPort(host, strconv.Itoa(port))
	sshClient, err := ssh.Dial("tcp", targetAddr, sshConfig)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to establish SSH connection to %s: %v", targetAddr, err), http.StatusBadGateway)
		return
	}

	// 4. Create SSH Session
	sshSession, err := sshClient.NewSession()
	if err != nil {
		_ = sshClient.Close()
		http.Error(w, "Failed to create SSH session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Request PTY
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	if err := sshSession.RequestPty("xterm-256color", 24, 80, modes); err != nil {
		_ = sshSession.Close()
		_ = sshClient.Close()
		http.Error(w, "Failed to allocate PTY: "+err.Error(), http.StatusInternalServerError)
		return
	}

	stdinPipe, err := sshSession.StdinPipe()
	if err != nil {
		_ = sshSession.Close()
		_ = sshClient.Close()
		http.Error(w, "Failed to create stdin pipe: "+err.Error(), http.StatusInternalServerError)
		return
	}

	stdoutPipe, err := sshSession.StdoutPipe()
	if err != nil {
		_ = sshSession.Close()
		_ = sshClient.Close()
		http.Error(w, "Failed to create stdout pipe: "+err.Error(), http.StatusInternalServerError)
		return
	}

	stderrPipe, err := sshSession.StderrPipe()
	if err != nil {
		_ = sshSession.Close()
		_ = sshClient.Close()
		http.Error(w, "Failed to create stderr pipe: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 6. Upgrade HTTP to WebSocket
	conn, err := DockerUpgrader.Upgrade(w, r, nil)
	if err != nil {
		_ = sshSession.Close()
		_ = sshClient.Close()
		return
	}

	// 7. Start Remote Shell
	if err := sshSession.Shell(); err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("\r\n[Blackwater Error] Failed to start shell: %v\r\n", err)))
		_ = conn.Close()
		_ = sshSession.Close()
		_ = sshClient.Close()
		return
	}

	sessionID := fmt.Sprintf("ssh_%d_%d", userID, time.Now().UnixNano())
	remoteSession := &RemoteSSHSession{
		SessionID:     sessionID,
		UserID:        userID,
		Username:      username,
		ClientIP:      clientIP,
		RemoteHost:    host,
		RemotePort:    port,
		RemoteUser:    remoteUser,
		Conn:          conn,
		SSHClient:     sshClient,
		SSHSession:    sshSession,
		StdinWriter:   stdinPipe,
		audit_service: crud_service.AuditLogCRUD{Repo: Repository.NewAuditRepository(Config.DB)},
	}

	SSHTerminalPool.AddSession(remoteSession)

	// Audit Log
	auditModel := models.AuditLog{
		UserID:      &userID,
		ServiceType: "SSH",
		ServiceID:   fmt.Sprintf("%s:%d", host, port),
		Action:      "SSH_REMOTE_TERMINAL_CONNECT",
		Results:     fmt.Sprintf("Established interactive SSH session to %s@%s:%d (IP: %s)", remoteUser, host, port, clientIP),
	}
	_ = remoteSession.audit_service.CreateAudit(&auditModel)

	// Pipe SSH stdout & stderr -> WebSocket
	go func() {
		defer remoteSession.Close()
		buf := make([]byte, 4096)
		for {
			n, rErr := stdoutPipe.Read(buf)
			if n > 0 {
				remoteSession.mu.Lock()
				if remoteSession.isClosed {
					remoteSession.mu.Unlock()
					return
				}
				wErr := conn.WriteMessage(websocket.TextMessage, buf[:n])
				remoteSession.mu.Unlock()
				if wErr != nil {
					return
				}
			}
			if rErr != nil {
				return
			}
		}
	}()

	go func() {
		defer remoteSession.Close()
		buf := make([]byte, 4096)
		for {
			n, rErr := stderrPipe.Read(buf)
			if n > 0 {
				remoteSession.mu.Lock()
				if remoteSession.isClosed {
					remoteSession.mu.Unlock()
					return
				}
				wErr := conn.WriteMessage(websocket.TextMessage, buf[:n])
				remoteSession.mu.Unlock()
				if wErr != nil {
					return
				}
			}
			if rErr != nil {
				return
			}
		}
	}()

	// Pipe WebSocket -> SSH stdin & Window Change
	go func() {
		defer func() {
			remoteSession.Close()
			SSHTerminalPool.RemoveSession(sessionID)
		}()

		for {
			msgType, msgBytes, rErr := conn.ReadMessage()
			if rErr != nil {
				return
			}

			if msgType == websocket.TextMessage || msgType == websocket.BinaryMessage {
				// Check if this is a structured JSON command (e.g. resize)
				var structuredMsg SSHWSMessage
				if err := json.Unmarshal(msgBytes, &structuredMsg); err == nil && structuredMsg.Type != "" {
					switch structuredMsg.Type {
					case "resize":
						if structuredMsg.Rows > 0 && structuredMsg.Cols > 0 {
							_ = sshSession.WindowChange(structuredMsg.Rows, structuredMsg.Cols)
						}
					case "input":
						_, _ = stdinPipe.Write([]byte(structuredMsg.Data))
					case "ping":
						remoteSession.mu.Lock()
						_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"pong"}`))
						remoteSession.mu.Unlock()
					}
				} else {
					// Raw direct terminal input
					_, _ = stdinPipe.Write(msgBytes)
				}
			}
		}
	}()
}
