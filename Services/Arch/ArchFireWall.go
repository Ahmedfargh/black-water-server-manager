package Arch

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/ahmedfargh/server-manager/Services/RedHat"
)

type ArchFireWall struct {
	redHatFirewall *RedHat.RedHatFireWall
}

func NewArchFireWall() *ArchFireWall {
	return &ArchFireWall{
		redHatFirewall: RedHat.NewRedHatFireWall(),
	}
}

// hasBinary checks if a given binary is in PATH
func (f *ArchFireWall) hasBinary(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}

func (f *ArchFireWall) UFWAction(args ...string) (string, error) {
	cmd := exec.Command("ufw", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback to sudo if permission denied
		cmdSudo := exec.Command("sudo", append([]string{"ufw"}, args...)...)
		outSudo, errSudo := cmdSudo.CombinedOutput()
		if errSudo == nil {
			return string(outSudo), nil
		}
		return string(output), err
	}
	return string(output), nil
}

func (f *ArchFireWall) Enable() (string, error) {
	if f.hasBinary("ufw") {
		return f.UFWAction("enable")
	}
	if f.hasBinary("firewall-cmd") {
		return f.redHatFirewall.Enable()
	}
	return "No supported firewall tool found (ufw / firewalld)", fmt.Errorf("no firewall daemon available on Arch")
}

func (f *ArchFireWall) Disable() (string, error) {
	if f.hasBinary("ufw") {
		return f.UFWAction("disable")
	}
	if f.hasBinary("firewall-cmd") {
		return f.redHatFirewall.Disable()
	}
	return "No supported firewall tool found (ufw / firewalld)", fmt.Errorf("no firewall daemon available on Arch")
}

func (f *ArchFireWall) Status() (string, error) {
	if f.hasBinary("ufw") {
		return f.UFWAction("status")
	}
	if f.hasBinary("firewall-cmd") {
		return f.redHatFirewall.Status()
	}
	if f.hasBinary("iptables") {
		cmd := exec.Command("iptables", "-L", "-n")
		_, err := cmd.CombinedOutput()
		if err == nil {
			return "running (iptables)", nil
		}
	}
	return "inactive (no firewall daemon installed)", nil
}

func (f *ArchFireWall) Rules() (string, error) {
	if f.hasBinary("ufw") {
		return f.UFWAction("status", "numbered")
	}
	if f.hasBinary("firewall-cmd") {
		return f.redHatFirewall.Rules()
	}
	if f.hasBinary("iptables") {
		cmd := exec.Command("iptables", "-S")
		out, err := cmd.CombinedOutput()
		if err == nil {
			return strings.TrimSpace(string(out)), nil
		}
	}
	return "No active rules found", nil
}

func (f *ArchFireWall) ListRules() (string, error) {
	if f.hasBinary("ufw") {
		return f.UFWAction("status")
	}
	if f.hasBinary("firewall-cmd") {
		return f.redHatFirewall.ListRules()
	}
	if f.hasBinary("iptables") {
		cmd := exec.Command("iptables", "-L", "-v", "-n")
		out, err := cmd.CombinedOutput()
		if err == nil {
			return strings.TrimSpace(string(out)), nil
		}
	}
	return "No active rules found", nil
}

func (f *ArchFireWall) AddRule() bool {
	return true
}

func (f *ArchFireWall) DeleteRule() bool {
	return true
}

func (f *ArchFireWall) UpdateRule() bool {
	return true
}

func (f *ArchFireWall) ClearRules() bool {
	return true
}
