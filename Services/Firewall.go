package Services

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	Config "github.com/ahmedfargh/server-manager/Config"
	CRUD "github.com/ahmedfargh/server-manager/Database/CRUD"
	Models "github.com/ahmedfargh/server-manager/Database/Models"
	Repository "github.com/ahmedfargh/server-manager/Database/Repository"
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

// ValidateIPOrCIDR validates whether input is a valid IPv4/IPv6 address or CIDR subnet
func ValidateIPOrCIDR(input string) (string, bool, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", false, fmt.Errorf("ip address cannot be empty")
	}

	// Check if it's a CIDR
	if strings.Contains(trimmed, "/") {
		ip, _, err := net.ParseCIDR(trimmed)
		if err != nil {
			return "", false, fmt.Errorf("invalid CIDR subnet format: %w", err)
		}
		isIPv6 := ip.To4() == nil
		return trimmed, isIPv6, nil
	}

	// Check if it's a single IP
	parsedIP := net.ParseIP(trimmed)
	if parsedIP == nil {
		return "", false, fmt.Errorf("invalid IP address format: %s", trimmed)
	}
	isIPv6 := parsedIP.To4() == nil
	return trimmed, isIPv6, nil
}

func NewAuditLogCRUD() *CRUD.AuditLogCRUD {
	return CRUD.NewAuditLogCRUD(Repository.NewAuditRepository(Config.DB))
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

func (f *Firewall) Enable(UserId int) (string, error) {
	fmt.Printf("Enabling firewall on platform: %s\n", f.Platform)

	switch f.Platform {
	case "debian", "ubuntu":
		message, err := f.ubuntuFirewall.Enable()
		f.recordLog(message, uint(UserId))
		return message, err

	case "arch":
		message, err := f.archFirewall.Enable()
		f.recordLog(message, uint(UserId))
		return message, err
	case "redhat":
		message, err := f.redHatFirewall.Enable()
		f.recordLog(message, uint(UserId))
		return message, err
	default:
		return "UNSUPPORTED PLATFORM: " + f.Platform, fmt.Errorf("unsupported platform: %s", f.Platform)
	}
}
func (f *Firewall) recordLog(message string, UserId uint) {
	if Config.DB == nil {
		return
	}
	audit := NewAuditLogCRUD()
	var userIDPtr *uint
	if UserId != 0 {
		userIDPtr = &UserId
	}
	audit_model := Models.AuditLog{
		ServiceType: "firewall",
		UserID:      userIDPtr,
		Action:      message,
		ServiceID:   "firewall",
	}
	audit.CreateAudit(&audit_model)
}

func (f *Firewall) Disable(UserId int) (string, error) {
	fmt.Printf("Disabling firewall on platform: %s\n", f.Platform)
	switch f.Platform {
	case "debian", "ubuntu":
		message, err := f.ubuntuFirewall.Disable()
		f.recordLog(message, uint(UserId))
		return message, err
	case "arch":
		message, err := f.archFirewall.Disable()
		f.recordLog(message, uint(UserId))
		return message, err
	case "redhat":
		message, err := f.redHatFirewall.Disable()
		f.recordLog(message, uint(UserId))
		return message, err
	default:
		return "UNSUPPORTED PLATFORM: " + f.Platform, fmt.Errorf("unsupported platform: %s", f.Platform)
	}
}

func (f *Firewall) Status(UserId int) (string, error) {
	fmt.Printf("Getting firewall status on platform: %s\n", f.Platform)

	switch f.Platform {
	case "debian", "ubuntu":
		message, err := f.ubuntuFirewall.Status()
		f.recordLog(message, uint(UserId))
		return message, err
	case "arch":
		message, err := f.archFirewall.Status()
		f.recordLog(message, uint(UserId))
		return message, err
	case "redhat":
		message, err := f.redHatFirewall.Status()
		f.recordLog(message, uint(UserId))
		return message, err
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

func (f *Firewall) BlockIP(ip string, UserId int) (string, error) {
	cleanedIP, isIPv6, err := ValidateIPOrCIDR(ip)
	if err != nil {
		return "", err
	}
	fmt.Printf("Blocking IP %s on platform: %s\n", cleanedIP, f.Platform)

	var message string
	switch f.Platform {
	case "debian", "ubuntu":
		message, err = f.ubuntuFirewall.BlockIP(cleanedIP)
	case "arch":
		message, err = f.archFirewall.BlockIP(cleanedIP, isIPv6)
	case "redhat":
		message, err = f.redHatFirewall.BlockIP(cleanedIP, isIPv6)
	default:
		return "UNSUPPORTED PLATFORM: " + f.Platform, fmt.Errorf("unsupported platform: %s", f.Platform)
	}

	if err == nil {
		f.recordLog(fmt.Sprintf("Blocked IP %s: %s", cleanedIP, message), uint(UserId))
	}
	return message, err
}

func (f *Firewall) UnblockIP(ip string, UserId int) (string, error) {
	cleanedIP, isIPv6, err := ValidateIPOrCIDR(ip)
	if err != nil {
		return "", err
	}
	fmt.Printf("Unblocking IP %s on platform: %s\n", cleanedIP, f.Platform)

	var message string
	switch f.Platform {
	case "debian", "ubuntu":
		message, err = f.ubuntuFirewall.UnblockIP(cleanedIP)
	case "arch":
		message, err = f.archFirewall.UnblockIP(cleanedIP, isIPv6)
	case "redhat":
		message, err = f.redHatFirewall.UnblockIP(cleanedIP, isIPv6)
	default:
		return "UNSUPPORTED PLATFORM: " + f.Platform, fmt.Errorf("unsupported platform: %s", f.Platform)
	}

	if err == nil {
		f.recordLog(fmt.Sprintf("Unblocked IP %s: %s", cleanedIP, message), uint(UserId))
	}
	return message, err
}

