package Services

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ahmedfargh/server-manager/Services/Arch"
	"github.com/ahmedfargh/server-manager/Services/RedHat"
	Ubuntu "github.com/ahmedfargh/server-manager/Services/Ubuntu"
)

type Firewall struct {
	ubuntuFirewall *Ubuntu.UbuntuFireWall
	archFirewall   *Arch.ArchFireWall
	redHatFirewall *RedHat.RedHatFireWall
	Platform       string
}

func NewFirewall() *Firewall {
	Firewall := &Firewall{
		ubuntuFirewall: Ubuntu.NewUbuntuFireWall(),
		archFirewall:   Arch.NewArchFireWall(),
		redHatFirewall: RedHat.NewRedHatFireWall(),
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
	var osID, osLike string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			osID = strings.Trim(strings.Split(line, "=")[1], "\"")
		}
		if strings.HasPrefix(line, "ID_LIKE=") {
			osLike = strings.Trim(strings.Split(line, "=")[1], "\"")
		}
	}

	if osID == "ubuntu" || strings.Contains(osLike, "ubuntu") || osID == "debian" || strings.Contains(osLike, "debian") {
		return "debian"
	}
	if osID == "arch" || strings.Contains(osLike, "arch") {
		return "arch"
	}
	if osID == "fedora" || osID == "centos" || osID == "rhel" || strings.Contains(osLike, "fedora") || strings.Contains(osLike, "rhel") || strings.Contains(osLike, "centos") {
		return "redhat"
	}

	return "unknown"
}

func (f *Firewall) Enable() (string, error) {
	fmt.Printf("Enabling firewall on platform: %s\n", f.Platform)
	switch f.Platform {
	case "debian", "ubuntu":
		return f.ubuntuFirewall.Enable()
	case "arch":
		return f.archFirewall.Enable()
	case "redhat":
		return f.redHatFirewall.Enable()
	default:
		return "UNSUPPORTED PLATFORM: " + f.Platform, fmt.Errorf("unsupported platform: %s", f.Platform)
	}
}

func (f *Firewall) Disable() (string, error) {
	fmt.Printf("Disabling firewall on platform: %s\n", f.Platform)
	switch f.Platform {
	case "debian", "ubuntu":
		return f.ubuntuFirewall.Disable()
	case "arch":
		return f.archFirewall.Disable()
	case "redhat":
		return f.redHatFirewall.Disable()
	default:
		return "UNSUPPORTED PLATFORM: " + f.Platform, fmt.Errorf("unsupported platform: %s", f.Platform)
	}
}

func (f *Firewall) Status() (string, error) {
	fmt.Printf("Getting firewall status on platform: %s\n", f.Platform)
	switch f.Platform {
	case "debian", "ubuntu":
		return f.ubuntuFirewall.Status()
	case "arch":
		return f.archFirewall.Status()
	case "redhat":
		return f.redHatFirewall.Status()
	default:
		return "UNSUPPORTED PLATFORM: " + f.Platform, fmt.Errorf("unsupported platform: %s", f.Platform)
	}
}

func (f *Firewall) Rules() (string, error) {
	fmt.Printf("Getting firewall rules on platform: %s\n", f.Platform)
	switch f.Platform {
	case "debian", "ubuntu":
		return f.ubuntuFirewall.Rules()
	case "arch":
		return f.archFirewall.Rules()
	case "redhat":
		return f.redHatFirewall.Rules()
	default:
		return "UNSUPPORTED PLATFORM: " + f.Platform, fmt.Errorf("unsupported platform: %s", f.Platform)
	}
}

func (f *Firewall) ListRules() (string, error) {
	fmt.Printf("Listing firewall rules on platform: %s\n", f.Platform)
	switch f.Platform {
	case "debian", "ubuntu":
		return f.ubuntuFirewall.ListRules()
	case "arch":
		return f.archFirewall.ListRules()
	case "redhat":
		return f.redHatFirewall.ListRules()
	default:
		return "UNSUPPORTED PLATFORM: " + f.Platform, fmt.Errorf("unsupported platform: %s", f.Platform)
	}
}
