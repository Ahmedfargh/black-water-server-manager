package Ubuntu

import (
	"os/exec"
)

type UbuntuSystemUpdate struct {
}

func NewUbuntuSystemUpdate() *UbuntuSystemUpdate {
	return &UbuntuSystemUpdate{}
}

func (u *UbuntuSystemUpdate) Update() (string, error) {
	cmd := exec.Command("apt-get", "update")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}
	return string(output), nil
}

func (u *UbuntuSystemUpdate) Upgrade() (string, error) {
	cmd := exec.Command("apt-get", "upgrade", "-y")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}
	return string(output), nil
}
