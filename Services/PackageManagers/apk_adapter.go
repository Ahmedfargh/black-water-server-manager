package PackageManagers

import (
	"bufio"
	"context"
	"regexp"
	"strings"
	"time"
)

// ApkAdapter implements PackageManagerAdapter for Alpine Linux
type ApkAdapter struct {
	BaseAdapter
}

func NewApkAdapter() *ApkAdapter {
	return &ApkAdapter{
		BaseAdapter: NewBaseAdapter("apk", CategoryNative, "apk"),
	}
}

func (a *ApkAdapter) GetVersion(ctx context.Context) (string, error) {
	out, err := a.RunCommand(ctx, "--version")
	if err != nil && out == "" {
		return "", err
	}
	re := regexp.MustCompile(`apk-tools\s+([0-9.]+)`)
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

func (a *ApkAdapter) GetInstalledCount(ctx context.Context) (int, error) {
	out, err := a.RunCommand(ctx, "info")
	if err != nil && out == "" {
		return 0, err
	}
	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return 0, nil
	}
	return strings.Count(trimmed, "\n") + 1, nil
}

func (a *ApkAdapter) GetPendingUpdates(ctx context.Context) ([]PackageUpdate, error) {
	out, _ := a.RunCommand(ctx, "version", "-l", "<")
	if strings.TrimSpace(out) == "" {
		return []PackageUpdate{}, nil
	}

	var updates []PackageUpdate
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "Installed:") {
			continue
		}
		// Format: pkg-name-1.0.0 < 1.0.1
		parts := strings.Fields(line)
		if len(parts) >= 3 && parts[1] == "<" {
			updates = append(updates, PackageUpdate{
				Name:           parts[0],
				CurrentVersion: "installed",
				NewVersion:     parts[2],
			})
		}
	}
	return updates, nil
}

func (a *ApkAdapter) GetInstalledPackages(ctx context.Context, query string, page int, limit int) ([]PackageItem, int, error) {
	out, err := a.RunCommand(ctx, "info", "-v")
	if err != nil && out == "" {
		return nil, 0, err
	}

	var allItems []PackageItem
	q := strings.ToLower(strings.TrimSpace(query))
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Format: pkg-name-version
		lastDash := strings.LastIndex(line, "-")
		if lastDash != -1 {
			name := line[:lastDash]
			version := line[lastDash+1:]
			if q != "" && !strings.Contains(strings.ToLower(name), q) {
				continue
			}
			allItems = append(allItems, PackageItem{
				Name:    name,
				Version: version,
				Source:  "apk",
			})
		}
	}

	items, total := PaginateSlice(allItems, page, limit)
	return items, total, nil
}

func (a *ApkAdapter) CleanCache(ctx context.Context) (*OperationResult, error) {
	return a.ExecuteMutation(ctx, "clean_cache", "", 2*time.Minute, nil, "cache", "clean")
}

func (a *ApkAdapter) RefreshRepositories(ctx context.Context) (*OperationResult, error) {
	return a.ExecuteMutation(ctx, "refresh_repositories", "", 3*time.Minute, nil, "update")
}

func (a *ApkAdapter) UpgradeSystem(ctx context.Context) (*OperationResult, error) {
	return a.ExecuteMutation(ctx, "upgrade_system", "", 10*time.Minute, nil, "upgrade")
}

func (a *ApkAdapter) InstallPackage(ctx context.Context, packageName string) (*OperationResult, error) {
	if err := ValidatePackageName(packageName); err != nil {
		return nil, err
	}
	return a.ExecuteMutation(ctx, "install_package", packageName, 5*time.Minute, nil, "add", packageName)
}

func (a *ApkAdapter) RemovePackage(ctx context.Context, packageName string, purge bool) (*OperationResult, error) {
	if err := ValidatePackageName(packageName); err != nil {
		return nil, err
	}
	if purge {
		return a.ExecuteMutation(ctx, "remove_package", packageName, 5*time.Minute, nil, "del", "--purge", packageName)
	}
	return a.ExecuteMutation(ctx, "remove_package", packageName, 5*time.Minute, nil, "del", packageName)
}

func (a *ApkAdapter) UpgradePackage(ctx context.Context, packageName string) (*OperationResult, error) {
	if err := ValidatePackageName(packageName); err != nil {
		return nil, err
	}
	return a.ExecuteMutation(ctx, "upgrade_package", packageName, 5*time.Minute, nil, "add", "-u", packageName)
}
