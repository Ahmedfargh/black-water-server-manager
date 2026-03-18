package RedHat

import (
	"os/exec"
)

type RedHatFireWall struct {
}

func NewRedHatFireWall() *RedHatFireWall {
	return &RedHatFireWall{}
}

func (f *RedHatFireWall) Command(args ...string) (string, error) {
	sudoArgs := append([]string{"firewall-cmd"}, args...)
	cmd := exec.Command("sudo", sudoArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}
	return string(output), nil
}

func (f *RedHatFireWall) Enable() (string, error) {
	// Start and enable the service via systemctl
	cmd := exec.Command("systemctl", "start", "firewalld")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}
	exec.Command("systemctl", "enable", "firewalld").Run()
	return "Firewalld started and enabled successfully", nil
}

func (f *RedHatFireWall) Disable() (string, error) {
	// Stop and disable the service via systemctl
	cmd := exec.Command("sudo", "systemctl", "stop", "firewalld")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}
	exec.Command("sudo", "systemctl", "disable", "firewalld").Run()
	return "Firewalld stopped and disabled successfully", nil
}

func (f *RedHatFireWall) Status() (string, error) {
	return f.Command("--state")
}

func (f *RedHatFireWall) Rules() (string, error) {
	return f.Command("--list-all")
}

func (f *RedHatFireWall) ListRules() (string, error) {
	return f.Command("--list-all")
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
