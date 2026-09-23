package RedHat

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

type RedHatFireWall struct {
}

func NewRedHatFireWall() *RedHatFireWall {
	return &RedHatFireWall{}
}

func (f *RedHatFireWall) Command(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 1. Try directly without sudo (firewall-cmd state & rules reading does not need sudo)
	cmd := exec.CommandContext(ctx, "firewall-cmd", args...)
	output, err := cmd.CombinedOutput()
	if err == nil {
		return string(output), nil
	}

	// 2. Fallback to sudo if needed
	sudoArgs := append([]string{"-n", "firewall-cmd"}, args...)
	cmdSudo := exec.CommandContext(ctx, "sudo", sudoArgs...)
	outSudo, errSudo := cmdSudo.CombinedOutput()
	if errSudo == nil {
		return string(outSudo), nil
	}

	return string(output), err
}

func (f *RedHatFireWall) Enable() (string, error) {
	// Start and enable the service via systemctl
	cmd := exec.Command("systemctl", "start", "firewalld")
	output, err := cmd.CombinedOutput()
	if err != nil {
		cmdSudo := exec.Command("sudo", "-n", "systemctl", "start", "firewalld")
		outputSudo, errSudo := cmdSudo.CombinedOutput()
		if errSudo != nil {
			return string(output), err
		}
		output = outputSudo
	}
	exec.Command("systemctl", "enable", "firewalld").Run()
	return "Firewalld started and enabled successfully", nil
}

func (f *RedHatFireWall) Disable() (string, error) {
	// Stop and disable the service via systemctl
	cmd := exec.Command("systemctl", "stop", "firewalld")
	output, err := cmd.CombinedOutput()
	if err != nil {
		cmdSudo := exec.Command("sudo", "-n", "systemctl", "stop", "firewalld")
		outputSudo, errSudo := cmdSudo.CombinedOutput()
		if errSudo != nil {
			return string(output), err
		}
		output = outputSudo
	}
	exec.Command("systemctl", "disable", "firewalld").Run()
	return "Firewalld stopped and disabled successfully", nil
}

func (f *RedHatFireWall) Status() (string, error) {
	out, err := f.Command("--state")
	if err != nil {
		if strings.TrimSpace(out) != "" {
			return strings.TrimSpace(out), nil
		}
		return "not running", nil
	}
	return strings.TrimSpace(out), nil
}

func (f *RedHatFireWall) Rules() (string, error) {
	out, err := f.Command("--list-all")
	if err != nil {
		if strings.TrimSpace(out) != "" {
			return strings.TrimSpace(out), nil
		}
		return "Firewall is inactive or not running", nil
	}
	return strings.TrimSpace(out), nil
}

func (f *RedHatFireWall) ListRules() (string, error) {
	return f.Rules()
}

func (f *RedHatFireWall) BlockIP(ip string, isIPv6 bool) (string, error) {
	family := "ipv4"
	if isIPv6 {
		family = "ipv6"
	}
	richRule := "rule family='" + family + "' source address='" + ip + "' drop"
	
	// Add permanent and runtime rules
	f.Command("--add-rich-rule=" + richRule)
	out, err := f.Command("--permanent", "--add-rich-rule="+richRule)
	if err != nil {
		return out, err
	}
	f.Command("--reload")
	return "Blocked IP " + ip + " successfully", nil
}

func (f *RedHatFireWall) UnblockIP(ip string, isIPv6 bool) (string, error) {
	family := "ipv4"
	if isIPv6 {
		family = "ipv6"
	}
	richRule := "rule family='" + family + "' source address='" + ip + "' drop"

	f.Command("--remove-rich-rule=" + richRule)
	out, err := f.Command("--permanent", "--remove-rich-rule="+richRule)
	if err != nil {
		return out, err
	}
	f.Command("--reload")
	return "Unblocked IP " + ip + " successfully", nil
}

func (f *RedHatFireWall) AddRule() bool {
	return true
}
func (f *RedHatFireWall) DeleteRule() bool {
	return true
}
func (f *RedHatFireWall) UpdateRule() bool {
	return true
}
func (f *RedHatFireWall) ClearRules() bool {
	return true
}

