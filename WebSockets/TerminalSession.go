package WebSockets

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	Config "github.com/ahmedfargh/server-manager/Config"
	crud_service "github.com/ahmedfargh/server-manager/Database/CRUD"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	Repository "github.com/ahmedfargh/server-manager/Database/Repository"
	websocket "github.com/gorilla/websocket"
	json "github.com/json-iterator/go"
)

const (
	MaxCommandLength = 2048
	MaxOutputBytes   = 256 * 1024 // 256 KB output buffer cap
	CommandTimeout   = 30 * time.Second
)

// Dangerous command patterns that pose catastrophic risk to the host
var dangerousCommandPatterns = []*regexp.Regexp{
	// Root deletion patterns (rm -rf /, rm -r -f /, rm -rf /*, --no-preserve-root)
	regexp.MustCompile(`(?i)\brm\s+.*(-[a-zA-Z]*r[a-zA-Z]*f[a-zA-Z]*|-[a-zA-Z]*r[a-zA-Z]*\s+-[a-zA-Z]*f[a-zA-Z]*|-[a-zA-Z]*f[a-zA-Z]*\s+-[a-zA-Z]*r[a-zA-Z]*).*(/|/\*|/\.\*)(\s|$)`),
	// Fork bombs
	regexp.MustCompile(`(?i):\(\)\s*\{\s*:\|:&\s*\};:`),
	regexp.MustCompile(`(?i)bomb\(\)\s*\{\s*bomb\|bomb&\s*\};`),
	// Raw drive destruction
	regexp.MustCompile(`(?i)\bdd\s+.*of=/dev/(sd[a-z]|nvme[0-9]|hd[a-z]|vd[a-z]|mmcblk[0-9])`),
	regexp.MustCompile(`(?i)>\s*/dev/(sd[a-z]|nvme[0-9]|hd[a-z]|vd[a-z]|mmcblk[0-9])`),
	regexp.MustCompile(`(?i)\bmkfs(\.[a-z0-9]+)?\s+.*(/dev/(sd[a-z]|nvme[0-9]|hd[a-z]|vd[a-z]|mmcblk[0-9]))`),
	regexp.MustCompile(`(?i)\b(wipefs|fdisk|parted)\s+.*(/dev/(sd[a-z]|nvme[0-9]|hd[a-z]|vd[a-z]|mmcblk[0-9]))`),
	// Kernel panic triggers
	regexp.MustCompile(`(?i)echo\s+[cbseio]\s*>\s*/proc/sysrq-trigger`),
}

// SanitizeTerminalCommand validates and filters incoming command strings
func SanitizeTerminalCommand(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("empty command")
	}

	// Remove dangerous null bytes and non-printable control characters (allow tab \t and newline \n)
	var sanitized strings.Builder
	for _, r := range trimmed {
		if r == 0 {
			continue // Drop NULL bytes
		}
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			continue // Drop unprintable control characters
		}
		sanitized.WriteRune(r)
	}

	result := sanitized.String()
	if len(result) > MaxCommandLength {
		return "", fmt.Errorf("command exceeds maximum allowed limit of %d characters", MaxCommandLength)
	}

	// Check against catastrophic patterns
	for _, pattern := range dangerousCommandPatterns {
		if pattern.MatchString(result) {
			return "", fmt.Errorf("🛡️ [SECURITY ENGINE] Command blocked: Destructive or forbidden system pattern detected")
		}
	}

	return result, nil
}

type TerminalSession struct {
	SessionID      int32  `json:"session_id"`
	Username       string `json:"username"`
	ClientIP       string `json:"client_ip"`
	Conn           *websocket.Conn
	ExecuteCommand chan string
	SendResult     chan string
	audit_service  crud_service.AuditLogCRUD
	closeOnce      sync.Once
	isClosed       bool
	mu             sync.Mutex
}

type terminalPool struct {
	mu       sync.RWMutex
	Sessions map[int32]*TerminalSession
}

var TerminalPool = &terminalPool{
	Sessions: make(map[int32]*TerminalSession),
}

func (ts *terminalPool) ConnectSession(sessionID int32, clientIP string, username string, conn *websocket.Conn) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Terminate any existing session for this user to prevent zombie connections and resource leaks
	if existing, exists := ts.Sessions[sessionID]; exists {
		existing.Close()
	}

	newSession := &TerminalSession{
		SessionID:      sessionID,
		Username:       username,
		ClientIP:       clientIP,
		Conn:           conn,
		ExecuteCommand: make(chan string, 10),
		SendResult:     make(chan string, 10),
		audit_service:  crud_service.AuditLogCRUD{Repo: Repository.NewAuditRepository(Config.DB)},
	}
	ts.Sessions[sessionID] = newSession

	go newSession.WritePump()
	go newSession.ReadPump()
	go newSession.RunCommands()
}

func (ts *terminalPool) RemoveSession(sessionID int32) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if session, exists := ts.Sessions[sessionID]; exists {
		session.Close()
		delete(ts.Sessions, sessionID)
	}
}

func (ts *TerminalSession) WritePump() {
	defer ts.Close()
	for result := range ts.SendResult {
		ts.mu.Lock()
		if ts.isClosed || ts.Conn == nil {
			ts.mu.Unlock()
			return
		}
		err := ts.Conn.WriteMessage(websocket.TextMessage, []byte(result))
		ts.mu.Unlock()
		if err != nil {
			return
		}
	}
}

func (ts *TerminalSession) ReadPump() {
	defer func() {
		ts.Close()
		TerminalPool.RemoveSession(ts.SessionID)
	}()

	for {
		_, message, err := ts.Conn.ReadMessage()
		if err != nil {
			return
		}

		ts.mu.Lock()
		closed := ts.isClosed
		ts.mu.Unlock()
		if closed {
			return
		}

		ts.ExecuteCommand <- string(message)
	}
}

func (ts *TerminalSession) RunCommands() {
	defer ts.Close()

	for cmdPayload := range ts.ExecuteCommand {
		var commandStructure map[string]any
		err := json.Unmarshal([]byte(cmdPayload), &commandStructure)
		if err != nil {
			ts.safeSend("❌ Invalid command payload format")
			continue
		}

		rawCommand, ok := commandStructure["command"].(string)
		if !ok || strings.TrimSpace(rawCommand) == "" {
			ts.safeSend("❌ Command field is missing or empty")
			continue
		}

		// 1. Sanitize and validate command
		cleanCmd, sanitizeErr := SanitizeTerminalCommand(rawCommand)
		if sanitizeErr != nil {
			// Record blocked security audit event
			userID := uint(ts.SessionID)
			auditLog := models.AuditLog{
				UserID:      &userID,
				ServiceType: "terminal_security_block",
				ServiceID:   ts.ClientIP,
				Action:      rawCommand,
				Results:     sanitizeErr.Error(),
			}
			_ = ts.audit_service.CreateAudit(&auditLog)

			ts.safeSend(fmt.Sprintf("%v", sanitizeErr))
			continue
		}

		// 2. Execute with context timeout and process group isolation
		ctx, cancel := context.WithTimeout(context.Background(), CommandTimeout)
		execCmd := exec.CommandContext(ctx, "sh", "-c", cleanCmd)

		// Set process group so we can terminate child trees cleanly on timeout
		execCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		var stdout, stderr bytes.Buffer
		execCmd.Stdout = &stdout
		execCmd.Stderr = &stderr

		startTime := time.Now()
		execErr := execCmd.Start()

		var result string
		if execErr != nil {
			result = fmt.Sprintf("Failed to spawn process: %v", execErr)
		} else {
			done := make(chan error, 1)
			go func() {
				done <- execCmd.Wait()
			}()

			select {
			case <-ctx.Done():
				// Kill the entire process group
				if execCmd.Process != nil && execCmd.Process.Pid > 0 {
					_ = syscall.Kill(-execCmd.Process.Pid, syscall.SIGKILL)
				}
				result = fmt.Sprintf("⏱️ [TIMEOUT] Command execution timed out after %v", CommandTimeout)
			case waitErr := <-done:
				outBytes := stdout.Bytes()
				errBytes := stderr.Bytes()

				// Cap output size to prevent DoS/memory exhaustion
				if len(outBytes) > MaxOutputBytes {
					outBytes = append(outBytes[:MaxOutputBytes], []byte("\n... [OUTPUT TRUNCATED: Exceeded 256 KB limit]")...)
				}
				if len(errBytes) > MaxOutputBytes {
					errBytes = append(errBytes[:MaxOutputBytes], []byte("\n... [STDERR TRUNCATED]")...)
				}

				result = string(outBytes)
				if len(errBytes) > 0 {
					if len(result) > 0 && !strings.HasSuffix(result, "\n") {
						result += "\n"
					}
					result += string(errBytes)
				}

				if waitErr != nil && result == "" {
					result = fmt.Sprintf("Error: %v", waitErr)
				}
			}
		}
		cancel()

		elapsed := time.Since(startTime).Milliseconds()

		// 3. Persist audit trail with client IP attribution
		userID := uint(ts.SessionID)
		auditLog := models.AuditLog{
			UserID:      &userID,
			ServiceType: "terminal_session",
			ServiceID:   ts.ClientIP,
			Action:      cleanCmd,
			Results:     fmt.Sprintf("[Duration: %dms] %s", elapsed, result),
		}
		_ = ts.audit_service.CreateAudit(&auditLog)

		ts.safeSend(result)
	}
}

func (ts *TerminalSession) safeSend(msg string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.isClosed {
		return
	}
	select {
	case ts.SendResult <- msg:
	default:
		// Drop if buffer full to prevent blocking
	}
}

func (ts *TerminalSession) Close() {
	ts.closeOnce.Do(func() {
		ts.mu.Lock()
		ts.isClosed = true
		if ts.Conn != nil {
			_ = ts.Conn.Close()
		}
		ts.mu.Unlock()
	})
}
