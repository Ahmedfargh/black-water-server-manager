package Ubuntu

import "os/exec"

type UbuntuFireWall struct {
}

func NewUbuntuFireWall() *UbuntuFireWall {
	return &UbuntuFireWall{}
}
func (f *UbuntuFireWall) UFWAction(args ...string) (string, error) {
	cmd := exec.Command("ufw", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		cmdSudo := exec.Command("sudo", append([]string{"ufw"}, args...)...)
		outSudo, errSudo := cmdSudo.CombinedOutput()
		if errSudo == nil {
			return string(outSudo), nil
		}
		return "", err
	}
	return string(output), nil
}
func (f *UbuntuFireWall) Enable() (string, error) {
	return f.UFWAction("enable")
}
func (f *UbuntuFireWall) Disable() (string, error) {
	return f.UFWAction("disable")
}
func (f *UbuntuFireWall) Status() (string, error) {
	return f.UFWAction("status")
}
func (f *UbuntuFireWall) Rules() (string, error) {
	return f.UFWAction("status", "numbered")
}
func (f *UbuntuFireWall) ListRules() (string, error) {
	return f.UFWAction("status")
}
func (f *UbuntuFireWall) AddRule() bool {
	return true
}
func (f *UbuntuFireWall) DeleteRule() bool {
	return true
}
func (f *UbuntuFireWall) UpdateRule() bool {
	return true
}
func (f *UbuntuFireWall) ClearRules() bool {
	return true
}
