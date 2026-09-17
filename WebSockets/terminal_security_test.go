package WebSockets

import (
	"strings"
	"testing"
)

func TestSanitizeTerminalCommand(t *testing.T) {
	// 1. Valid safe commands should pass cleanly
	validCommands := []string{
		"ls -la /var/log",
		"htop",
		"ps aux | grep nginx",
		"df -h",
		"free -m",
		"cat /etc/os-release",
		"uptime",
		"docker ps -a",
		"systemctl status nginx",
		"echo 'Hello World'",
		"find . -name '*.go'",
	}

	for _, cmd := range validCommands {
		clean, err := SanitizeTerminalCommand(cmd)
		if err != nil {
			t.Errorf("Expected valid command '%s' to pass, got error: %v", cmd, err)
		}
		if clean != cmd {
			t.Errorf("Expected sanitized output to match '%s', got '%s'", cmd, clean)
		}
	}

	// 2. Destructive & Dangerous commands that MUST be blocked
	dangerousCommands := []string{
		"rm -rf /",
		"rm -rf /*",
		"rm -rf /.*",
		"rm -r -f /",
		"rm --no-preserve-root -rf /",
		":(){ :|:& };:",
		"bomb() { bomb|bomb& }; bomb",
		"dd if=/dev/zero of=/dev/sda",
		"dd if=/dev/urandom of=/dev/nvme0n1",
		"> /dev/sda",
		"> /dev/nvme0n1",
		"mkfs.ext4 /dev/sda1",
		"wipefs -a /dev/sda",
		"fdisk /dev/sda",
		"echo c > /proc/sysrq-trigger",
		"echo b > /proc/sysrq-trigger",
	}

	for _, cmd := range dangerousCommands {
		_, err := SanitizeTerminalCommand(cmd)
		if err == nil {
			t.Errorf("Expected dangerous command '%s' to be BLOCKED, but it passed!", cmd)
		} else if !strings.Contains(err.Error(), "SECURITY ENGINE") {
			t.Errorf("Expected security block error for '%s', got: %v", cmd, err)
		}
	}

	// 3. Control characters & excessive payload tests
	nullByteCmd := "ls\x00 -la"
	clean, err := SanitizeTerminalCommand(nullByteCmd)
	if err != nil || strings.Contains(clean, "\x00") {
		t.Errorf("Expected null bytes to be stripped, got '%s', err: %v", clean, err)
	}

	oversizedCmd := strings.Repeat("a", MaxCommandLength+100)
	_, err = SanitizeTerminalCommand(oversizedCmd)
	if err == nil {
		t.Errorf("Expected oversized command to be rejected, but it passed")
	}

	emptyCmd := "   "
	_, err = SanitizeTerminalCommand(emptyCmd)
	if err == nil {
		t.Errorf("Expected empty command to be rejected, but it passed")
	}
}
