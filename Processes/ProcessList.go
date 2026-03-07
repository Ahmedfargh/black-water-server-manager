package processes

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	config "github.com/ahmedfargh/server-manager/Config"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	repository "github.com/ahmedfargh/server-manager/Database/Repository"
	// "github.com/shirou/gopsutil/v3/process"
)

type ProcessInfo struct {
	PID    int32
	Name   string
	Status string
}

func GetProcesses() ([]ProcessInfo, error) {
	var processList []ProcessInfo
	files, err := os.ReadDir("/proc")
	if err != nil {
		return processList, fmt.Errorf("failed to read /proc directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			var proc ProcessInfo
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
func GetProcessByPID(pid int32) (ProcessInfo, error) {
	var proc ProcessInfo
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
func saveProcessToDB(proc *models.Process) (*models.Process, error) {
	repo := repository.NewProcessRepository(config.DB)
	err := repo.CreateProcess(proc)
	if err != nil {
		return nil, err
	}
	return proc, nil
}
func GetPaginatedProcesses(page, pageSize int) ([]models.Process, int64, error) {
	repo := repository.NewProcessRepository(config.DB)
	return repo.GetPaginatedProcesses(page, pageSize)
}

func StartProcess(command string, args ...string) (*models.Process, error) {
	cmd := exec.Command(command, args...)
	err := cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start process: %w", err)
	}
	go func() {
		err := cmd.Wait()
		if err != nil {
			fmt.Printf("Process exited with error: %v\n", err)
		} else {
			fmt.Printf("Process exited successfully\n")
		}
	}()
	proc := models.Process{
		PID:     int32(cmd.Process.Pid),
		Name:    command,
		Status:  "Running",
		Command: command,
		Args:    strings.Join(args, " "),
	}
	saved, err := saveProcessToDB(&proc)
	if err != nil {
		return nil, err
	}
	return saved, nil
}
func KillProcess(pid int32) error {
	proc, err := os.FindProcess(int(pid))
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}
	err = proc.Kill()
	if err != nil {
		return fmt.Errorf("failed to kill process: %w", err)
	}
	return nil
}
