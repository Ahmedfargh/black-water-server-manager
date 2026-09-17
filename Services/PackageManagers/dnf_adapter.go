package PackageManagers

import (
	"bufio"
	"context"
	"regexp"
	"strings"
	"time"
)

// DnfAdapter implements PackageManagerAdapter for Red Hat/Fedora/CentOS distributions
type DnfAdapter struct {
	BaseAdapter
}

func NewDnfAdapter() *DnfAdapter {
	execName := "dnf"
	b := NewBaseAdapter("dnf", CategoryNative, execName)
	if !b.IsAvailable() {
		// Fallback to yum if dnf is not present
		b = NewBaseAdapter("yum", CategoryNative, "yum")
	}
	return &DnfAdapter{BaseAdapter: b}
}

func (d *DnfAdapter) GetVersion(ctx context.Context) (string, error) {
	out, err := d.RunCommand(ctx, "--version")
	if err != nil && out == "" {
		return "", err
	}
	re := regexp.MustCompile(`([0-9.]+)`)
	matches := re.FindStringSubmatch(out)
	if len(matches) > 1 {
		return matches[1], nil
	}
	lines := strings.Split(out, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}
	return "Unknown", nil
}

func (d *DnfAdapter) GetInstalledCount(ctx context.Context) (int, error) {
	out, err := d.RunCommand(ctx, "list", "installed")
	if err != nil && out == "" {
		return 0, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) <= 1 {
		return 0, nil
	}
	return len(lines) - 1, nil
}

func (d *DnfAdapter) GetPendingUpdates(ctx context.Context) ([]PackageUpdate, error) {
	// dnf check-update returns exit code 100 when updates are found
	out, _ := d.RunCommand(ctx, "check-update")
	if strings.TrimSpace(out) == "" {
		return []PackageUpdate{}, nil
	}

	var updates []PackageUpdate
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "Last metadata") || strings.HasPrefix(line, "Security:") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			updates = append(updates, PackageUpdate{
				Name:       parts[0],
				NewVersion: parts[1],
				Repository: parts[2],
			})
		}
	}
	return updates, nil
}

func (d *DnfAdapter) GetInstalledPackages(ctx context.Context, query string, page int, limit int) ([]PackageItem, int, error) {
	out, err := d.RunCommand(ctx, "list", "installed")
	if err != nil && out == "" {
		return nil, 0, err
	}

	var allItems []PackageItem
	q := strings.ToLower(strings.TrimSpace(query))
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "Installed Packages") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			name := parts[0]
			version := parts[1]
			if q != "" && !strings.Contains(strings.ToLower(name), q) {
				continue
			}
			allItems = append(allItems, PackageItem{
				Name:    name,
				Version: version,
				Source:  d.Name,
			})
		}
	}

	items, total := PaginateSlice(allItems, page, limit)
	return items, total, nil
}

func (d *DnfAdapter) CleanCache(ctx context.Context) (*OperationResult, error) {
	return d.ExecuteMutation(ctx, "clean_cache", "", 2*time.Minute, nil, "clean", "all")
}

func (d *DnfAdapter) RefreshRepositories(ctx context.Context) (*OperationResult, error) {
	return d.ExecuteMutation(ctx, "refresh_repositories", "", 3*time.Minute, nil, "makecache")
}

func (d *DnfAdapter) UpgradeSystem(ctx context.Context) (*OperationResult, error) {
	return d.ExecuteMutation(ctx, "upgrade_system", "", 10*time.Minute, nil, "upgrade", "-y")
}

func (d *DnfAdapter) InstallPackage(ctx context.Context, packageName string) (*OperationResult, error) {
	if err := ValidatePackageName(packageName); err != nil {
		return nil, err
	}
	return d.ExecuteMutation(ctx, "install_package", packageName, 5*time.Minute, nil, "install", "-y", packageName)
}

func (d *DnfAdapter) RemovePackage(ctx context.Context, packageName string, purge bool) (*OperationResult, error) {
	if err := ValidatePackageName(packageName); err != nil {
		return nil, err
	}
	return d.ExecuteMutation(ctx, "remove_package", packageName, 5*time.Minute, nil, "remove", "-y", packageName)
}

func (d *DnfAdapter) UpgradePackage(ctx context.Context, packageName string) (*OperationResult, error) {
	if err := ValidatePackageName(packageName); err != nil {
		return nil, err
	}
	return d.ExecuteMutation(ctx, "upgrade_package", packageName, 5*time.Minute, nil, "upgrade", "-y", packageName)
}
