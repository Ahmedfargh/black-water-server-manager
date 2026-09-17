package PackageManagers

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// BaseAdapter provides common helpers for command execution
type BaseAdapter struct {
	Name           string
	Category       ManagerCategory
	ExecutableName string
	ExecutablePath string
	DefaultTimeout time.Duration
}

// NewBaseAdapter initializes a BaseAdapter with safe defaults
func NewBaseAdapter(name string, category ManagerCategory, executableName string) BaseAdapter {
	path, _ := exec.LookPath(executableName)
	return BaseAdapter{
		Name:           name,
		Category:       category,
		ExecutableName: executableName,
		ExecutablePath: path,
		DefaultTimeout: 10 * time.Second,
	}
}

func (b *BaseAdapter) GetName() string {
	return b.Name
}

func (b *BaseAdapter) GetCategory() ManagerCategory {
	return b.Category
}

func (b *BaseAdapter) IsAvailable() bool {
	return b.ExecutablePath != ""
}

func (b *BaseAdapter) GetExecutablePath() string {
	return b.ExecutablePath
}

// RunCommand executes a command safely with timeout context
func (b *BaseAdapter) RunCommand(ctx context.Context, args ...string) (string, error) {
	return b.RunCommandWithTimeout(ctx, b.DefaultTimeout, nil, args...)
}

// RunCommandWithTimeout executes a command with custom timeout and optional environment variables
func (b *BaseAdapter) RunCommandWithTimeout(ctx context.Context, timeout time.Duration, extraEnv []string, args ...string) (string, error) {
	return b.runExecutableWithTimeout(ctx, b.ExecutablePath, timeout, extraEnv, args...)
}

// runExecutableWithTimeout handles execution with stdin pipe and timeout
func (b *BaseAdapter) runExecutableWithTimeout(ctx context.Context, execPath string, timeout time.Duration, extraEnv []string, args ...string) (string, error) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, execPath, args...)
	if len(extraEnv) > 0 {
		cmd.Env = append(cmd.Environ(), extraEnv...)
	}

	// Always provide automated 'yes' stdin to prevent pacman/apt/dnf from failing on interactive prompts (Error reading fd 0)
	cmd.Stdin = strings.NewReader("y\ny\ny\ny\ny\ny\ny\ny\ny\ny\n")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	outStr := stdout.String()
	errStr := stderr.String()

	combined := strings.TrimSpace(outStr)
	if combined == "" && errStr != "" {
		combined = strings.TrimSpace(errStr)
	} else if errStr != "" && outStr != "" {
		combined = combined + "\n" + strings.TrimSpace(errStr)
	}

	return combined, err
}

// ValidatePackageName ensures package names contain safe alphanumeric characters only to prevent command injection
func ValidatePackageName(packageName string) error {
	trimmed := strings.TrimSpace(packageName)
	if trimmed == "" {
		return fmt.Errorf("package name cannot be empty")
	}
	if len(trimmed) > 128 {
		return fmt.Errorf("package name exceeds maximum length of 128 characters")
	}
	re := regexp.MustCompile(`^[a-zA-Z0-9_\-\.\:\+@\/]+$`)
	if !re.MatchString(trimmed) {
		return fmt.Errorf("package name '%s' contains invalid characters", trimmed)
	}
	return nil
}

// ExecuteMutation executes a mutation command with timing, sudo detection, and structured OperationResult return
func (b *BaseAdapter) ExecuteMutation(ctx context.Context, action string, targetPkg string, timeout time.Duration, extraEnv []string, args ...string) (*OperationResult, error) {
	start := time.Now()
	
	execPath := b.ExecutablePath
	execArgs := args
	usedSudo := false

	// If not root, check if passwordless sudo is available
	if os.Geteuid() != 0 && b.Category == CategoryNative {
		sudoPath, err := exec.LookPath("sudo")
		if err == nil {
			// Test if sudo -n (non-interactive) works
			sudoCheck := exec.CommandContext(ctx, sudoPath, "-n", "true")
			if sudoCheck.Run() == nil {
				execPath = sudoPath
				execArgs = append([]string{"-n", b.ExecutablePath}, args...)
				usedSudo = true
			}
		}
	}

	fullCmd := b.ExecutableName + " " + strings.Join(args, " ")
	if usedSudo {
		fullCmd = "sudo -n " + fullCmd
	}

	output, err := b.runExecutableWithTimeout(ctx, execPath, timeout, extraEnv, execArgs...)
	elapsed := time.Since(start).Milliseconds()

	// Append permission advice if command failed due to root permissions
	if err != nil && (strings.Contains(output, "you cannot perform this operation unless you are root") || 
		strings.Contains(output, "Permission denied") || 
		strings.Contains(output, "are you root?")) {
		output = output + "\n\n⚠️ [PRIVILEGE NOTICE] Root privileges required for system package operations.\nOptions:\n1. Run the Blackwater backend process with root/sudo.\n2. Or grant passwordless sudo for package manager in /etc/sudoers (e.g. `%wheel ALL=(ALL) NOPASSWD: ALL` or `%wheel ALL=(ALL) NOPASSWD: /usr/bin/" + b.ExecutableName + "`)"
	}

	res := &OperationResult{
		Success:         err == nil,
		Action:          action,
		Manager:         b.Name,
		TargetPackage:   targetPkg,
		Command:         fullCmd,
		Output:          output,
		ExecutionTimeMs: elapsed,
		ExecutedAt:      time.Now(),
	}

	if err != nil {
		res.Message = fmt.Sprintf("Failed to execute %s on %s: %v", action, b.Name, err)
		return res, err
	}

	res.Message = fmt.Sprintf("Successfully executed %s on %s", action, b.Name)
	return res, nil
}

// PaginateSlice utility helper for package arrays
func PaginateSlice[T any](items []T, page int, limit int) ([]T, int) {
	total := len(items)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}

	start := (page - 1) * limit
	if start >= total {
		return []T{}, total
	}

	end := start + limit
	if end > total {
		end = total
	}

	return items[start:end], total
}
