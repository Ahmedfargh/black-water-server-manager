package PackageManagers

import (
	"bytes"
	"context"
	"os/exec"
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
	timeout := b.DefaultTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, b.ExecutablePath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := strings.TrimSpace(stdout.String())
	if err != nil && output == "" {
		return strings.TrimSpace(stderr.String()), err
	}
	return output, err
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
