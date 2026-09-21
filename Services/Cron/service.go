package cron

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Service struct {
	mu sync.Mutex
}

func NewService() *Service {
	return &Service{}
}

// GenerateJobID creates a deterministic hash ID from the raw line content
func GenerateJobID(line string) string {
	hasher := sha256.New()
	hasher.Write([]byte(strings.TrimSpace(line)))
	return hex.EncodeToString(hasher.Sum(nil))[:16]
}

// HumanizeExpression converts a standard 5-part cron or macro into English
func HumanizeExpression(expr string) string {
	expr = strings.TrimSpace(expr)
	switch expr {
	case "@reboot":
		return "Run once at system startup"
	case "@hourly", "0 * * * *":
		return "Every hour at minute 0"
	case "@daily", "@midnight", "0 0 * * *":
		return "Every day at midnight (00:00)"
	case "@weekly", "0 0 * * 0":
		return "Every Sunday at midnight"
	case "@monthly", "0 0 1 * *":
		return "First day of every month at midnight"
	case "@yearly", "@annually", "0 0 1 1 *":
		return "First day of every year at midnight"
	}

	parts := strings.Fields(expr)
	if len(parts) != 5 {
		return "Custom Schedule (" + expr + ")"
	}

	min, hour, dom, mon, dow := parts[0], parts[1], parts[2], parts[3], parts[4]

	// Handle */X minute patterns
	if strings.HasPrefix(min, "*/") && hour == "*" && dom == "*" && mon == "*" && dow == "*" {
		return fmt.Sprintf("Every %s minutes", strings.TrimPrefix(min, "*/"))
	}

	// Handle daily at HH:MM
	if !strings.Contains(min, "*") && !strings.Contains(hour, "*") && dom == "*" && mon == "*" && dow == "*" {
		return fmt.Sprintf("Daily at %02s:%02s", hour, min)
	}

	// Handle hourly at :MM
	if !strings.Contains(min, "*") && hour == "*" && dom == "*" && mon == "*" && dow == "*" {
		return fmt.Sprintf("Every hour at minute %s", min)
	}

	return fmt.Sprintf("Schedule: %s %s (Day %s, Month %s, WDay %s)", min, hour, dom, mon, dow)
}

// readCrontabRaw executes `crontab -l` and returns the raw output lines
func (s *Service) readCrontabRaw() ([]string, error) {
	cmd := exec.Command("crontab", "-l")
	out, err := cmd.CombinedOutput()
	if err != nil {
		outStr := string(out)
		// If no crontab exists for user, return empty slice without error
		if strings.Contains(outStr, "no crontab for") || strings.Contains(outStr, "crontab: no crontab") {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read crontab: %s", strings.TrimSpace(outStr))
	}

	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, nil
}

// writeCrontabRaw writes lines back to crontab
func (s *Service) writeCrontabRaw(lines []string) error {
	tmpFile, err := os.CreateTemp("", "blackwater-crontab-*")
	if err != nil {
		return fmt.Errorf("failed to create temp crontab file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	content := strings.Join(lines, "\n")
	if len(lines) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write to temp crontab file: %w", err)
	}
	tmpFile.Close()

	cmd := exec.Command("crontab", tmpFile.Name())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("crontab install error: %s", strings.TrimSpace(string(out)))
	}

	return nil
}

// ListJobs returns all parsed cron jobs
func (s *Service) ListJobs() ([]CronJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	lines, err := s.readCrontabRaw()
	if err != nil {
		return nil, err
	}

	var jobs []CronJob
	var currentComment string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			currentComment = ""
			continue
		}

		// Check if pure comment line
		if strings.HasPrefix(trimmed, "#") {
			// Check if this is a disabled cron job: "# * * * * * command" or "# @daily command"
			candidate := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
			if isCronExpressionLine(candidate) {
				job := parseCronLine(candidate, false, currentComment, line)
				jobs = append(jobs, job)
				currentComment = ""
				continue
			}

			// Regular commentary
			commentText := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
			if currentComment == "" {
				currentComment = commentText
			} else {
				currentComment += " " + commentText
			}
			continue
		}

		// Active cron job line
		if isCronExpressionLine(trimmed) {
			job := parseCronLine(trimmed, true, currentComment, line)
			jobs = append(jobs, job)
			currentComment = ""
		}
	}

	return jobs, nil
}

func isCronExpressionLine(line string) bool {
	if strings.HasPrefix(line, "@") {
		return true
	}
	fields := strings.Fields(line)
	return len(fields) >= 6
}

func parseCronLine(line string, enabled bool, comment string, rawLine string) CronJob {
	fields := strings.Fields(line)
	var expr, cmd string

	if strings.HasPrefix(line, "@") && len(fields) >= 2 {
		expr = fields[0]
		cmd = strings.Join(fields[1:], " ")
	} else if len(fields) >= 6 {
		expr = strings.Join(fields[0:5], " ")
		cmd = strings.Join(fields[5:], " ")
	} else {
		expr = "custom"
		cmd = line
	}

	return CronJob{
		ID:            GenerateJobID(rawLine),
		Expression:    expr,
		ScheduleHuman: HumanizeExpression(expr),
		Command:       cmd,
		Comment:       comment,
		Enabled:       enabled,
		Raw:           rawLine,
	}
}

// AddJob appends a new task to crontab
func (s *Service) AddJob(req CronMutationRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lines, err := s.readCrontabRaw()
	if err != nil {
		return err
	}

	var newLines []string
	if req.Comment != "" {
		newLines = append(newLines, fmt.Sprintf("# %s", req.Comment))
	}

	taskLine := fmt.Sprintf("%s %s", strings.TrimSpace(req.Expression), strings.TrimSpace(req.Command))
	if !req.Enabled {
		taskLine = "# " + taskLine
	}
	newLines = append(newLines, taskLine)

	lines = append(lines, newLines...)
	return s.writeCrontabRaw(lines)
}

// UpdateJob updates an existing task matching the ID
func (s *Service) UpdateJob(id string, req CronMutationRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lines, err := s.readCrontabRaw()
	if err != nil {
		return err
	}

	found := false
	var updatedLines []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if GenerateJobID(line) == id {
			found = true
			if req.Comment != "" {
				// If previous line was a comment, replace or add
				if len(updatedLines) > 0 && strings.HasPrefix(updatedLines[len(updatedLines)-1], "#") && !isCronExpressionLine(strings.TrimPrefix(updatedLines[len(updatedLines)-1], "#")) {
					updatedLines[len(updatedLines)-1] = fmt.Sprintf("# %s", req.Comment)
				} else {
					updatedLines = append(updatedLines, fmt.Sprintf("# %s", req.Comment))
				}
			}

			taskLine := fmt.Sprintf("%s %s", strings.TrimSpace(req.Expression), strings.TrimSpace(req.Command))
			if !req.Enabled {
				taskLine = "# " + taskLine
			}
			updatedLines = append(updatedLines, taskLine)
		} else {
			updatedLines = append(updatedLines, line)
		}
	}

	if !found {
		return errors.New("cron job not found")
	}

	return s.writeCrontabRaw(updatedLines)
}

// DeleteJob removes a task by ID
func (s *Service) DeleteJob(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lines, err := s.readCrontabRaw()
	if err != nil {
		return err
	}

	found := false
	var updatedLines []string

	for _, line := range lines {
		if GenerateJobID(line) == id {
			found = true
			continue // skip deleting target line
		}
		updatedLines = append(updatedLines, line)
	}

	if !found {
		return errors.New("cron job not found")
	}

	return s.writeCrontabRaw(updatedLines)
}

// ToggleJob enables or disables a task
func (s *Service) ToggleJob(id string, enable bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lines, err := s.readCrontabRaw()
	if err != nil {
		return err
	}

	found := false
	var updatedLines []string

	for _, line := range lines {
		if GenerateJobID(line) == id {
			found = true
			trimmed := strings.TrimSpace(line)
			if enable {
				// Remove leading "# "
				line = strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
			} else {
				// Add leading "# "
				if !strings.HasPrefix(trimmed, "#") {
					line = "# " + trimmed
				}
			}
		}
		updatedLines = append(updatedLines, line)
	}

	if !found {
		return errors.New("cron job not found")
	}

	return s.writeCrontabRaw(updatedLines)
}

// RunJob manually executes a cron command and captures output
func (s *Service) RunJob(id string) (CronRunResponse, error) {
	jobs, err := s.ListJobs()
	if err != nil {
		return CronRunResponse{}, err
	}

	var target *CronJob
	for _, j := range jobs {
		if j.ID == id {
			target = &j
			break
		}
	}

	if target == nil {
		return CronRunResponse{}, errors.New("cron job not found")
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", target.Command)
	out, err := cmd.CombinedOutput()

	res := CronRunResponse{
		ID:         id,
		Command:    target.Command,
		Output:     string(out),
		DurationMs: time.Since(start).Milliseconds(),
		Success:    err == nil,
	}

	if err != nil {
		if res.Output == "" {
			res.Output = err.Error()
		}
		return res, fmt.Errorf("command execution error: %w", err)
	}

	if res.Output == "" {
		res.Output = "Command completed successfully with no stdout/stderr output."
	}

	return res, nil
}
