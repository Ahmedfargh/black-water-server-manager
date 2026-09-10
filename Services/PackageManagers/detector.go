package PackageManagers

import (
	"bufio"
	"os"
	"strings"
)

// SystemDetector detects the host operating system details
type SystemDetector interface {
	DetectOS() HostOSInfo
	ResolvePrimaryManager(osInfo HostOSInfo) string
}

type LinuxSystemDetector struct {
	osReleasePath string
}

func NewLinuxSystemDetector() *LinuxSystemDetector {
	return &LinuxSystemDetector{
		osReleasePath: "/etc/os-release",
	}
}

// DetectOS reads and parses /etc/os-release
func (d *LinuxSystemDetector) DetectOS() HostOSInfo {
	info := HostOSInfo{
		ID:         "linux",
		Name:       "Linux",
		PrettyName: "Linux System",
	}

	file, err := os.Open(d.osReleasePath)
	if err != nil {
		return info
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)

		switch key {
		case "ID":
			info.ID = strings.ToLower(val)
		case "NAME":
			info.Name = val
		case "PRETTY_NAME":
			info.PrettyName = val
		case "VERSION_ID":
			info.VersionID = val
		case "ID_LIKE":
			info.IDLike = strings.ToLower(val)
		}
	}
	return info
}

// ResolvePrimaryManager maps distribution characteristics to its standard native package manager
func (d *LinuxSystemDetector) ResolvePrimaryManager(osInfo HostOSInfo) string {
	id := strings.ToLower(osInfo.ID)
	idLike := strings.ToLower(osInfo.IDLike)

	// Direct ID match
	switch id {
	case "arch", "manjaro", "endeavouros", "artix", "garuda":
		return "pacman"
	case "ubuntu", "debian", "linuxmint", "pop", "elementary", "kali", "raspbian":
		return "apt"
	case "fedora", "rhel", "centos", "rocky", "almalinux", "oracle":
		return "dnf"
	case "opensuse", "opensuse-leap", "opensuse-tumbleweed", "sles":
		return "zypper"
	case "alpine":
		return "apk"
	}

	// ID_LIKE fallback
	if strings.Contains(idLike, "arch") {
		return "pacman"
	}
	if strings.Contains(idLike, "debian") || strings.Contains(idLike, "ubuntu") {
		return "apt"
	}
	if strings.Contains(idLike, "rhel") || strings.Contains(idLike, "fedora") || strings.Contains(idLike, "centos") {
		return "dnf"
	}
	if strings.Contains(idLike, "suse") {
		return "zypper"
	}

	return ""
}
