package Arch

import "os/exec"

type ArchFireWall struct {
}

func NewArchFireWall() *ArchFireWall {
	return &ArchFireWall{}
}

func (f *ArchFireWall) UFWAction(action string) (string, error) {
	cmd := exec.Command("sudo", "ufw", action)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (f *ArchFireWall) Enable() (string, error) {
	return f.UFWAction("enable")
}

func (f *ArchFireWall) Disable() (string, error) {
	return f.UFWAction("disable")
}

func (f *ArchFireWall) Status() (string, error) {
	return f.UFWAction("status")
}

func (f *ArchFireWall) Rules() (string, error) {
	return f.UFWAction("numbered")
}

func (f *ArchFireWall) ListRules() (string, error) {
	return f.UFWAction("list")
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
