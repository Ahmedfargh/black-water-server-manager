package PackageManagers

import (
	"context"
	"strings"
	"testing"
)

func TestLinuxSystemDetector(t *testing.T) {
	detector := NewLinuxSystemDetector()
	osInfo := detector.DetectOS()
	if osInfo.ID == "" {
		t.Errorf("Expected non-empty OS ID")
	}

	primary := detector.ResolvePrimaryManager(osInfo)
	t.Logf("Detected OS: %s (%s), Primary Manager: %s", osInfo.Name, osInfo.ID, primary)
}

func TestPackageManagerService(t *testing.T) {
	service := NewPackageManagerService(nil, nil)
	ctx := context.Background()

	overview, err := service.GetSystemOverview(ctx, true)
	if err != nil {
		t.Fatalf("Failed to get system overview: %v", err)
	}

	t.Logf("Detected %d available package managers", len(overview.Managers))
	for _, m := range overview.Managers {
		t.Logf("Manager: %s (version: %s, count: %d, primary: %v)", m.Name, m.Version, m.TotalPackages, m.IsPrimary)
	}
}

func TestValidatePackageName(t *testing.T) {
	valid := []string{"htop", "curl", "nginx-core", "libssl-dev", "python3.11", "app_1.0+b1", "@types/node", "org.mozilla.firefox"}
	for _, pkg := range valid {
		if err := ValidatePackageName(pkg); err != nil {
			t.Errorf("Expected valid for %s, got error: %v", pkg, err)
		}
	}

	invalid := []string{"", "htop; rm -rf /", "curl | bash", "pkg name", "pkg$(whoami)", "a`calc`", strings.Repeat("a", 130)}
	for _, pkg := range invalid {
		if err := ValidatePackageName(pkg); err == nil {
			t.Errorf("Expected invalid for %s, got no error", pkg)
		}
	}
}
