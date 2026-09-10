package PackageManagers

import (
	"bufio"
	"context"
	"regexp"
	"strings"
)

// AptAdapter implements PackageManagerAdapter for Debian/Ubuntu distributions
type AptAdapter struct {
	BaseAdapter
}

func NewAptAdapter() *AptAdapter {
	return &AptAdapter{
		BaseAdapter: NewBaseAdapter("apt", CategoryNative, "apt"),
	}
}

func (a *AptAdapter) GetVersion(ctx context.Context) (string, error) {
	out, err := a.RunCommand(ctx, "--version")
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`apt\s+([0-9.]+)`)
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

func (a *AptAdapter) GetInstalledCount(ctx context.Context) (int, error) {
	// dpkg-query is significantly faster and doesn't require root
	out, err := a.RunCommand(ctx, "list", "--installed")
	if err != nil {
		return 0, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) <= 1 {
		return 0, nil
	}
	// exclude header line "Listing..."
	return len(lines) - 1, nil
}

func (a *AptAdapter) GetPendingUpdates(ctx context.Context) ([]PackageUpdate, error) {
	out, err := a.RunCommand(ctx, "list", "--upgradable")
	if err != nil && out == "" {
		return []PackageUpdate{}, nil
	}

	var updates []PackageUpdate
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "Listing") {
			continue
		}
		// Format: package/suite new_version arch [upgradable from: old_version]
		parts := strings.Split(line, " ")
		if len(parts) >= 2 {
			nameParts := strings.Split(parts[0], "/")
			pkgName := nameParts[0]
			newVer := parts[1]
			oldVer := "installed"
			if idx := strings.Index(line, "[upgradable from: "); idx != -1 {
				oldVer = strings.TrimPrefix(line[idx:], "[upgradable from: ")
				oldVer = strings.TrimSuffix(oldVer, "]")
			}
			updates = append(updates, PackageUpdate{
				Name:           pkgName,
				CurrentVersion: oldVer,
				NewVersion:     newVer,
			})
		}
	}
	return updates, nil
}

func (a *AptAdapter) GetInstalledPackages(ctx context.Context, query string, page int, limit int) ([]PackageItem, int, error) {
	out, err := a.RunCommand(ctx, "list", "--installed")
	if err != nil && out == "" {
		return nil, 0, err
	}

	var allItems []PackageItem
	q := strings.ToLower(strings.TrimSpace(query))
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "Listing") {
			continue
		}
		// Format: package_name/suite version arch [installed]
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			name := strings.Split(parts[0], "/")[0]
			version := parts[1]
			if q != "" && !strings.Contains(strings.ToLower(name), q) {
				continue
			}
			allItems = append(allItems, PackageItem{
				Name:    name,
				Version: version,
				Source:  "apt",
			})
		}
	}

	items, total := PaginateSlice(allItems, page, limit)
	return items, total, nil
}
