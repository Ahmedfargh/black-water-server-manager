package PackageManagers

import (
	"bufio"
	"context"
	"regexp"
	"strings"
)

// ZypperAdapter implements PackageManagerAdapter for openSUSE / SLES
type ZypperAdapter struct {
	BaseAdapter
}

func NewZypperAdapter() *ZypperAdapter {
	return &ZypperAdapter{
		BaseAdapter: NewBaseAdapter("zypper", CategoryNative, "zypper"),
	}
}

func (z *ZypperAdapter) GetVersion(ctx context.Context) (string, error) {
	out, err := z.RunCommand(ctx, "--version")
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

func (z *ZypperAdapter) GetInstalledCount(ctx context.Context) (int, error) {
	out, err := z.RunCommand(ctx, "packages", "--installed-only")
	if err != nil && out == "" {
		return 0, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) <= 2 {
		return 0, nil
	}
	return len(lines) - 2, nil
}

func (z *ZypperAdapter) GetPendingUpdates(ctx context.Context) ([]PackageUpdate, error) {
	out, _ := z.RunCommand(ctx, "list-updates")
	if strings.TrimSpace(out) == "" {
		return []PackageUpdate{}, nil
	}

	var updates []PackageUpdate
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "v |") || strings.HasPrefix(line, "--") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 5 {
			updates = append(updates, PackageUpdate{
				Name:           strings.TrimSpace(parts[2]),
				CurrentVersion: strings.TrimSpace(parts[3]),
				NewVersion:     strings.TrimSpace(parts[4]),
			})
		}
	}
	return updates, nil
}

func (z *ZypperAdapter) GetInstalledPackages(ctx context.Context, query string, page int, limit int) ([]PackageItem, int, error) {
	out, err := z.RunCommand(ctx, "search", "--installed-only")
	if err != nil && out == "" {
		return nil, 0, err
	}

	var allItems []PackageItem
	q := strings.ToLower(strings.TrimSpace(query))
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "S |") || strings.HasPrefix(line, "--") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			name := strings.TrimSpace(parts[1])
			version := strings.TrimSpace(parts[3])
			if q != "" && !strings.Contains(strings.ToLower(name), q) {
				continue
			}
			allItems = append(allItems, PackageItem{
				Name:    name,
				Version: version,
				Source:  "zypper",
			})
		}
	}

	items, total := PaginateSlice(allItems, page, limit)
	return items, total, nil
}
