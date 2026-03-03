package processes

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	//"github.com/shirou/gopsutil/v3/process"
)

type processes struct {
	PID    int32
	Name   string
	Status string
}

func GetProcesses() ([]processes, error) {
	var processList []processes
	files, err := os.ReadDir("/proc")
	if err != nil {
		return processList, fmt.Errorf("failed to read /proc directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			var proc processes
			pid := file.Name()
			pidInt, err := strconv.Atoi(pid)
			if err != nil {
				continue
			}
			proc.PID = int32(pidInt)
			cmdlinePath := fmt.Sprintf("/proc/%s/cmdline", pid)
			cmdlineBytes, err := os.ReadFile(cmdlinePath)
			if err != nil {
				continue
			}
			proc.Name = string(cmdlineBytes)
			statusPath := fmt.Sprintf("/proc/%s/status", pid)
			statusBytes, err := os.ReadFile(statusPath)
			if err != nil {
				continue
			}
			statusLines := strings.Split(strings.TrimSpace(string(statusBytes)), "\n")
			for _, line := range statusLines {
				if strings.HasPrefix(line, "State:") {
					proc.Status = strings.TrimSpace(strings.TrimPrefix(line, "State:"))
					break
				}
			}
			processList = append(processList, proc)
		}
	}
	return processList, nil
}
func GetProcessByPID(pid int32) (processes, error) {
	var proc processes
	cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)
	cmdlineBytes, err := os.ReadFile(cmdlinePath)
	if err != nil {
		return proc, fmt.Errorf("failed to read cmdline file: %w", err)
	}
	proc.Name = string(cmdlineBytes)
	statusPath := fmt.Sprintf("/proc/%d/status", pid)
	statusBytes, err := os.ReadFile(statusPath)
	if err != nil {
		return proc, fmt.Errorf("failed to read status file: %w", err)
	}
	statusLines := strings.Split(strings.TrimSpace(string(statusBytes)), "\n")
	for _, line := range statusLines {
		if strings.HasPrefix(line, "State:") {
			proc.Status = strings.TrimSpace(strings.TrimPrefix(line, "State:"))
			break
		}
	}
	proc.PID = pid
	return proc, nil
}
