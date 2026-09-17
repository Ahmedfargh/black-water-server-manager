package PackageManagers

import (
	"bufio"
	"context"
	"regexp"
	"strings"
	"time"
)

// FlatpakAdapter implements PackageManagerAdapter for Flatpak packages
type FlatpakAdapter struct {
	BaseAdapter
}

func NewFlatpakAdapter() *FlatpakAdapter {
	return &FlatpakAdapter{
		BaseAdapter: NewBaseAdapter("flatpak", CategoryUniversal, "flatpak"),
	}
}

func (f *FlatpakAdapter) GetVersion(ctx context.Context) (string, error) {
	out, err := f.RunCommand(ctx, "--version")
	if err != nil && out == "" {
		return "", err
	}
	re := regexp.MustCompile(`Flatpak\s+([0-9.]+)`)
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

func (f *FlatpakAdapter) GetInstalledCount(ctx context.Context) (int, error) {
	out, err := f.RunCommand(ctx, "list", "--app")
	if err != nil && out == "" {
		return 0, err
	}
	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return 0, nil
	}
	return strings.Count(trimmed, "\n") + 1, nil
}

func (f *FlatpakAdapter) GetPendingUpdates(ctx context.Context) ([]PackageUpdate, error) {
	out, err := f.RunCommand(ctx, "remote-ls", "--updates")
	if err != nil && out == "" {
		return []PackageUpdate{}, nil
	}

	var updates []PackageUpdate
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			updates = append(updates, PackageUpdate{
				Name:           parts[0],
				CurrentVersion: "installed",
				NewVersion:     "available",
			})
		}
	}
	return updates, nil
}

func (f *FlatpakAdapter) GetInstalledPackages(ctx context.Context, query string, page int, limit int) ([]PackageItem, int, error) {
	out, err := f.RunCommand(ctx, "list", "--columns=name,application,version")
	if err != nil && out == "" {
		return nil, 0, err
	}

	var allItems []PackageItem
	q := strings.ToLower(strings.TrimSpace(query))
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		name := strings.TrimSpace(parts[0])
		appID := ""
		ver := "current"
		if len(parts) > 1 {
			appID = strings.TrimSpace(parts[1])
		}
		if len(parts) > 2 {
			ver = strings.TrimSpace(parts[2])
			if ver == "" {
				ver = "current"
			}
		}

		if q != "" && !strings.Contains(strings.ToLower(name), q) && !strings.Contains(strings.ToLower(appID), q) {
			continue
		}

		desc := appID
		allItems = append(allItems, PackageItem{
			Name:        name,
			Version:     ver,
			Description: desc,
			Source:      "flatpak",
		})
	}

	items, total := PaginateSlice(allItems, page, limit)
	return items, total, nil
}

func (f *FlatpakAdapter) CleanCache(ctx context.Context) (*OperationResult, error) {
	return f.ExecuteMutation(ctx, "clean_cache", "", 3*time.Minute, nil, "uninstall", "--unused", "-y")
}

func (f *FlatpakAdapter) RefreshRepositories(ctx context.Context) (*OperationResult, error) {
	return f.ExecuteMutation(ctx, "refresh_repositories", "", 3*time.Minute, nil, "update", "--appstream")
}

func (f *FlatpakAdapter) UpgradeSystem(ctx context.Context) (*OperationResult, error) {
	return f.ExecuteMutation(ctx, "upgrade_system", "", 10*time.Minute, nil, "update", "-y")
}

func (f *FlatpakAdapter) InstallPackage(ctx context.Context, packageName string) (*OperationResult, error) {
	if err := ValidatePackageName(packageName); err != nil {
		return nil, err
	}
	return f.ExecuteMutation(ctx, "install_package", packageName, 5*time.Minute, nil, "install", "-y", packageName)
}

func (f *FlatpakAdapter) RemovePackage(ctx context.Context, packageName string, purge bool) (*OperationResult, error) {
	if err := ValidatePackageName(packageName); err != nil {
		return nil, err
	}
	return f.ExecuteMutation(ctx, "remove_package", packageName, 5*time.Minute, nil, "uninstall", "-y", packageName)
}

func (f *FlatpakAdapter) UpgradePackage(ctx context.Context, packageName string) (*OperationResult, error) {
	if err := ValidatePackageName(packageName); err != nil {
		return nil, err
	}
	return f.ExecuteMutation(ctx, "upgrade_package", packageName, 5*time.Minute, nil, "update", "-y", packageName)
}
