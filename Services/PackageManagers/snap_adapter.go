package PackageManagers

import (
	"bufio"
	"context"
	"regexp"
	"strings"
)

// SnapAdapter implements PackageManagerAdapter for Canonical Snap
type SnapAdapter struct {
	BaseAdapter
}

func NewSnapAdapter() *SnapAdapter {
	return &SnapAdapter{
		BaseAdapter: NewBaseAdapter("snap", CategoryUniversal, "snap"),
	}
}

func (s *SnapAdapter) GetVersion(ctx context.Context) (string, error) {
	out, err := s.RunCommand(ctx, "version")
	if err != nil && out == "" {
		return "", err
	}
	re := regexp.MustCompile(`snap\s+([0-9.]+)`)
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

func (s *SnapAdapter) GetInstalledCount(ctx context.Context) (int, error) {
	out, err := s.RunCommand(ctx, "list")
	if err != nil && out == "" {
		return 0, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) <= 1 {
		return 0, nil
	}
	return len(lines) - 1, nil // subtract header
}

func (s *SnapAdapter) GetPendingUpdates(ctx context.Context) ([]PackageUpdate, error) {
	out, _ := s.RunCommand(ctx, "refresh", "--list")
	if strings.TrimSpace(out) == "" || strings.Contains(out, "All snaps up to date") {
		return []PackageUpdate{}, nil
	}

	var updates []PackageUpdate
	scanner := bufio.NewScanner(strings.NewReader(out))
	headerSkipped := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if !headerSkipped {
			headerSkipped = true
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			updates = append(updates, PackageUpdate{
				Name:           parts[0],
				CurrentVersion: "installed",
				NewVersion:     parts[1],
			})
		}
	}
	return updates, nil
}

func (s *SnapAdapter) GetInstalledPackages(ctx context.Context, query string, page int, limit int) ([]PackageItem, int, error) {
	out, err := s.RunCommand(ctx, "list")
	if err != nil && out == "" {
		return nil, 0, err
	}

	var allItems []PackageItem
	q := strings.ToLower(strings.TrimSpace(query))
	scanner := bufio.NewScanner(strings.NewReader(out))
	headerSkipped := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if !headerSkipped {
			headerSkipped = true
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
				Source:  "snap",
			})
		}
	}

	items, total := PaginateSlice(allItems, page, limit)
	return items, total, nil
}
