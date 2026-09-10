package PackageManagers

import (
	"bufio"
	"context"
	"regexp"
	"strings"
)

// PacmanAdapter implements PackageManagerAdapter for Arch Linux / Manjaro
type PacmanAdapter struct {
	BaseAdapter
}

func NewPacmanAdapter() *PacmanAdapter {
	return &PacmanAdapter{
		BaseAdapter: NewBaseAdapter("pacman", CategoryNative, "pacman"),
	}
}

func (p *PacmanAdapter) GetVersion(ctx context.Context) (string, error) {
	out, err := p.RunCommand(ctx, "-V")
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`Pacman v([0-9.]+)`)
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

func (p *PacmanAdapter) GetInstalledCount(ctx context.Context) (int, error) {
	out, err := p.RunCommand(ctx, "-Q")
	if err != nil {
		return 0, err
	}
	if out == "" {
		return 0, nil
	}
	return strings.Count(out, "\n") + 1, nil
}

func (p *PacmanAdapter) GetPendingUpdates(ctx context.Context) ([]PackageUpdate, error) {
	out, _ := p.RunCommand(ctx, "-Qu")
	if strings.TrimSpace(out) == "" {
		return []PackageUpdate{}, nil
	}

	var updates []PackageUpdate
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Format: package_name current_version -> new_version
		parts := strings.Fields(line)
		if len(parts) >= 4 && parts[2] == "->" {
			updates = append(updates, PackageUpdate{
				Name:           parts[0],
				CurrentVersion: parts[1],
				NewVersion:     parts[3],
				Repository:     "official",
			})
		} else if len(parts) >= 2 {
			updates = append(updates, PackageUpdate{
				Name:           parts[0],
				CurrentVersion: parts[1],
				NewVersion:     "available",
			})
		}
	}
	return updates, nil
}

func (p *PacmanAdapter) GetInstalledPackages(ctx context.Context, query string, page int, limit int) ([]PackageItem, int, error) {
	out, err := p.RunCommand(ctx, "-Q")
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
				Source:  "pacman",
			})
		}
	}

	items, total := PaginateSlice(allItems, page, limit)
	return items, total, nil
}
