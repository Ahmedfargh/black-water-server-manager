package systemd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var validUnitPattern = regexp.MustCompile(`^[a-zA-Z0-9_\-\.@:]+$`)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

// IsSystemdAvailable checks if systemctl binary exists and systemd is running
func (s *Service) IsSystemdAvailable() bool {
	_, err := exec.LookPath("systemctl")
	if err != nil {
		return false
	}
	// Check if systemd is booted (PID 1 is systemd or /run/systemd/system exists)
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		return true
	}
	// Test basic command
	cmd := exec.Command("systemctl", "is-system-running")
	_ = cmd.Run()
	return cmd.ProcessState != nil
}

// ListUnits returns active and inactive units on the system
func (s *Service) ListUnits(ctx context.Context, unitType string, search string) ([]UnitInfo, error) {
	if !s.IsSystemdAvailable() {
		return nil, errors.New("systemd is not available or running on this host")
	}

	args := []string{"list-units", "--all", "--no-pager", "--plain"}
	if unitType != "" && unitType != "all" {
		args = append(args, "--type="+unitType)
	}

	// Try JSON output first (systemd v246+)
	jsonArgs := append(args, "-o", "json")
	cmdJSON := exec.CommandContext(ctx, "systemctl", jsonArgs...)
	outJSON, err := cmdJSON.Output()
	if err == nil && len(outJSON) > 0 {
		var rawUnits []struct {
			Unit        string `json:"unit"`
			Load        string `json:"load"`
			Active      string `json:"active"`
			Sub         string `json:"sub"`
			Description string `json:"description"`
		}
		if err := json.Unmarshal(outJSON, &rawUnits); err == nil {
			var units []UnitInfo
			searchLower := strings.ToLower(search)
			for _, u := range rawUnits {
				uType := extractUnitType(u.Unit)
				if searchLower != "" {
					if !strings.Contains(strings.ToLower(u.Unit), searchLower) &&
						!strings.Contains(strings.ToLower(u.Description), searchLower) {
						continue
					}
				}
				units = append(units, UnitInfo{
					Unit:        u.Unit,
					Load:        u.Load,
					Active:      u.Active,
					Sub:         u.Sub,
					Description: u.Description,
					Type:        uType,
				})
			}
			return units, nil
		}
	}

	// Fallback to plain tabular output
	cmdPlain := exec.CommandContext(ctx, "systemctl", args...)
	outPlain, err := cmdPlain.Output()
	if err != nil && len(outPlain) == 0 {
		return nil, fmt.Errorf("systemctl execution failed: %w", err)
	}

	return s.parseTabularUnits(outPlain, search)
}

func extractUnitType(unit string) string {
	idx := strings.LastIndex(unit, ".")
	if idx != -1 && idx < len(unit)-1 {
		return unit[idx+1:]
	}
	return "service"
}

func (s *Service) parseTabularUnits(output []byte, search string) ([]UnitInfo, error) {
	var units []UnitInfo
	scanner := bufio.NewScanner(bytes.NewReader(output))
	searchLower := strings.ToLower(search)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "UNIT ") || strings.HasPrefix(line, "LOAD ") || strings.HasPrefix(line, "To show all") || strings.Contains(line, "loaded units listed") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		// Some lines might start with a dot or bullet symbol (e.g. "● nginx.service ...")
		startIdx := 0
		if fields[0] == "●" || fields[0] == "*" {
			startIdx = 1
		}

		if len(fields) <= startIdx+3 {
			continue
		}

		unitName := fields[startIdx]
		// Validate that it looks like a unit name (contains a dot suffix)
		if !strings.Contains(unitName, ".") {
			continue
		}

		load := fields[startIdx+1]
		active := fields[startIdx+2]
		sub := fields[startIdx+3]

		description := ""
		if len(fields) > startIdx+4 {
			description = strings.Join(fields[startIdx+4:], " ")
		}

		if searchLower != "" {
			if !strings.Contains(strings.ToLower(unitName), searchLower) &&
				!strings.Contains(strings.ToLower(description), searchLower) {
				continue
			}
		}

		units = append(units, UnitInfo{
			Unit:        unitName,
			Load:        load,
			Active:      active,
			Sub:         sub,
			Description: description,
			Type:        extractUnitType(unitName),
		})
	}

	return units, nil
}

// GetUnitStatus retrieves verbose status information for a unit
func (s *Service) GetUnitStatus(ctx context.Context, unitName string) (string, error) {
	if !validUnitPattern.MatchString(unitName) {
		return "", errors.New("invalid unit name format")
	}

	cmd := exec.CommandContext(ctx, "systemctl", "status", unitName, "--no-pager", "-l")
	out, _ := cmd.CombinedOutput()
	return string(out), nil
}

// ExecuteUnitAction executes lifecycle commands on a unit (start, stop, restart, enable, disable, reload)
func (s *Service) ExecuteUnitAction(ctx context.Context, unitName string, action string) (UnitActionResponse, error) {
	start := time.Now()
	res := UnitActionResponse{
		Unit:   unitName,
		Action: action,
	}

	if !validUnitPattern.MatchString(unitName) {
		res.Success = false
		res.Output = "Invalid unit name characters"
		res.DurationMs = time.Since(start).Milliseconds()
		return res, errors.New(res.Output)
	}

	validActions := map[string]bool{
		"start":   true,
		"stop":    true,
		"restart": true,
		"reload":  true,
		"enable":  true,
		"disable": true,
	}

	if !validActions[action] {
		res.Success = false
		res.Output = fmt.Sprintf("Unsupported action: %s", action)
		res.DurationMs = time.Since(start).Milliseconds()
		return res, errors.New(res.Output)
	}

	// Prepare execution command with sudo if non-root
	var cmdName string
	var args []string

	if os.Geteuid() != 0 {
		if _, err := exec.LookPath("sudo"); err == nil {
			cmdName = "sudo"
			args = []string{"-n", "systemctl", action, unitName}
		} else {
			cmdName = "systemctl"
			args = []string{action, unitName}
		}
	} else {
		cmdName = "systemctl"
		args = []string{action, unitName}
	}

	execCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, cmdName, args...)
	out, err := cmd.CombinedOutput()

	res.DurationMs = time.Since(start).Milliseconds()
	res.Output = strings.TrimSpace(string(out))

	if err != nil {
		res.Success = false
		if strings.Contains(res.Output, "sudo: a password is required") || strings.Contains(res.Output, "interactive authentication required") {
			res.Output = "Root privileges or passwordless sudo required for systemctl operations."
		}
		if res.Output == "" {
			res.Output = err.Error()
		}
		return res, fmt.Errorf("systemctl %s failed: %s", action, res.Output)
	}

	res.Success = true
	if res.Output == "" {
		res.Output = fmt.Sprintf("Successfully executed '%s' on %s", action, unitName)
	}
	return res, nil
}

// StreamUnitLogs opens a log stream via journalctl
func (s *Service) StreamUnitLogs(ctx context.Context, unitName string, lines int) (io.ReadCloser, error) {
	if !validUnitPattern.MatchString(unitName) {
		return nil, errors.New("invalid unit name")
	}

	if lines <= 0 {
		lines = 50
	}
	if lines > 500 {
		lines = 500
	}

	args := []string{"-u", unitName, "-f", "-n", fmt.Sprintf("%d", lines), "--no-pager"}
	cmd := exec.CommandContext(ctx, "journalctl", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start journalctl: %w", err)
	}

	return stdout, nil
}
