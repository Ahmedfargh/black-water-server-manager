package Arch

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ufw", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback to sudo if permission denied
		cmdSudo := exec.CommandContext(ctx, "sudo", append([]string{"-n", "ufw"}, args...)...)
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
		out, err := f.UFWAction("status")
		if err == nil {
			return out, nil
		}
	}
	if f.hasBinary("firewall-cmd") {
		out, err := f.redHatFirewall.Status()
		if err == nil && out != "not running" && out != "" {
			return out, nil
		}
	}
	if f.hasBinary("iptables") {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "iptables", "-L", "-n")
		_, err := cmd.CombinedOutput()
		if err == nil {
			return "running (iptables)", nil
		}
	}
	return "inactive (no firewall daemon running)", nil
}

func (f *ArchFireWall) Rules() (string, error) {
	if f.hasBinary("ufw") {
		out, err := f.UFWAction("status", "numbered")
		if err == nil {
			return out, nil
		}
	}
	if f.hasBinary("firewall-cmd") {
		out, err := f.redHatFirewall.Rules()
		if err == nil && !strings.Contains(out, "inactive") && !strings.Contains(out, "not running") {
			return out, nil
		}
	}
	if f.hasBinary("iptables") {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "iptables", "-S")
		out, err := cmd.CombinedOutput()
		if err == nil {
			return strings.TrimSpace(string(out)), nil
		}
	}
	return "No active rules found", nil
}

func (f *ArchFireWall) ListRules() (string, error) {
	return f.Rules()
}

func (f *ArchFireWall) BlockIP(ip string, isIPv6 bool) (string, error) {
	if f.hasBinary("ufw") {
		return f.UFWAction("deny", "from", ip)
	}
	if f.hasBinary("firewall-cmd") {
		return f.redHatFirewall.BlockIP(ip, isIPv6)
	}
	if f.hasBinary("iptables") {
		iptCmd := "iptables"
		if isIPv6 {
			iptCmd = "ip6tables"
		}
		cmd := exec.Command(iptCmd, "-I", "INPUT", "-s", ip, "-j", "DROP")
		out, err := cmd.CombinedOutput()
		if err != nil {
			cmdSudo := exec.Command("sudo", "-n", iptCmd, "-I", "INPUT", "-s", ip, "-j", "DROP")
			outSudo, errSudo := cmdSudo.CombinedOutput()
			if errSudo != nil {
				return string(out), err
			}
			out = outSudo
		}
		return "Blocked IP " + ip + " via iptables", nil
	}
	return "No supported firewall tool found", fmt.Errorf("no firewall tool available")
}

func (f *ArchFireWall) UnblockIP(ip string, isIPv6 bool) (string, error) {
	if f.hasBinary("ufw") {
		return f.UFWAction("delete", "deny", "from", ip)
	}
	if f.hasBinary("firewall-cmd") {
		return f.redHatFirewall.UnblockIP(ip, isIPv6)
	}
	if f.hasBinary("iptables") {
		iptCmd := "iptables"
		if isIPv6 {
			iptCmd = "ip6tables"
		}
		cmd := exec.Command(iptCmd, "-D", "INPUT", "-s", ip, "-j", "DROP")
		out, err := cmd.CombinedOutput()
		if err != nil {
			cmdSudo := exec.Command("sudo", "-n", iptCmd, "-D", "INPUT", "-s", ip, "-j", "DROP")
			outSudo, errSudo := cmdSudo.CombinedOutput()
			if errSudo != nil {
				return string(out), err
			}
			out = outSudo
		}
		return "Unblocked IP " + ip + " via iptables", nil
	}
	return "No supported firewall tool found", fmt.Errorf("no firewall tool available")
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

