package Services

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	Ubuntu "github.com/ahmedfargh/server-manager/Services/Ubuntu"
)

type Firewall struct {
	ubuntuFirewall *Ubuntu.UbuntuFireWall
	Platform       string
}

func NewFirewall() *Firewall {
	Firewall := &Firewall{
		ubuntuFirewall: Ubuntu.NewUbuntuFireWall(),
		Platform:       GetOs(),
	}
	return Firewall
}
func GetOs() string {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return "unknown"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var osInfo string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID_LIKE=") || strings.HasPrefix(line, "ID=") {
			osInfo = strings.Split(line, "=")[1]
		}
	}
	return osInfo
}
func (f *Firewall) Enable() (string, error) {
	fmt.Printf("Enabling firewall on platform: %s\n", f.Platform)
	switch f.Platform {
	case "debian", "ubuntu":
		return f.ubuntuFirewall.Enable()
	default:
		return "UNSUPPORTED PLATFORM" + f.Platform, fmt.Errorf("unsupported platform: %s", f.Platform)
	}
}
func (f *Firewall) Disable() (string, error) {
	fmt.Printf("Disabling firewall on platform: %s\n", f.Platform)
	switch f.Platform {
	case "debian", "ubuntu":
		return f.ubuntuFirewall.Disable()
	default:
		return "UNSUPPORTED PLATFORM" + f.Platform, fmt.Errorf("unsupported platform: %s", f.Platform)
	}
}

func (f *Firewall) Status() (string, error) {
	fmt.Printf("Getting firewall status on platform: %s\n", f.Platform)
	switch f.Platform {
	case "debian", "ubuntu":
		return f.ubuntuFirewall.Status()
	default:
		return "UNSUPPORTED PLATFORM" + f.Platform, fmt.Errorf("unsupported platform: %s", f.Platform)
	}
}
func (f *Firewall) Rules() (string, error) {
	fmt.Printf("Getting firewall rules on platform: %s\n", f.Platform)
	switch f.Platform {
	case "debian", "ubuntu":
		return f.ubuntuFirewall.Rules()
	default:
		return "UNSUPPORTED PLATFORM" + f.Platform, fmt.Errorf("unsupported platform: %s", f.Platform)
	}
}
func (f *Firewall) ListRules() (string, error) {
	fmt.Printf("Listing firewall rules on platform: %s\n", f.Platform)
	switch f.Platform {
	case "debian", "ubuntu":
		return f.ubuntuFirewall.ListRules()
	default:
		return "UNSUPPORTED PLATFORM" + f.Platform, fmt.Errorf("unsupported platform: %s", f.Platform)
	}
}
