package PackageManagers

import (
	"bufio"
	"context"
	"regexp"
	"strings"
	"time"
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

func (p *PacmanAdapter) CleanCache(ctx context.Context) (*OperationResult, error) {
	return p.ExecuteMutation(ctx, "clean_cache", "", 2*time.Minute, nil, "-Sc", "--noconfirm")
}

func (p *PacmanAdapter) RefreshRepositories(ctx context.Context) (*OperationResult, error) {
	return p.ExecuteMutation(ctx, "refresh_repositories", "", 3*time.Minute, nil, "-Sy", "--noconfirm")
}

func (p *PacmanAdapter) UpgradeSystem(ctx context.Context) (*OperationResult, error) {
	return p.ExecuteMutation(ctx, "upgrade_system", "", 10*time.Minute, nil, "-Su", "--noconfirm")
}

func (p *PacmanAdapter) InstallPackage(ctx context.Context, packageName string) (*OperationResult, error) {
	if err := ValidatePackageName(packageName); err != nil {
		return nil, err
	}
	return p.ExecuteMutation(ctx, "install_package", packageName, 5*time.Minute, nil, "-S", "--noconfirm", packageName)
}

func (p *PacmanAdapter) RemovePackage(ctx context.Context, packageName string, purge bool) (*OperationResult, error) {
	if err := ValidatePackageName(packageName); err != nil {
		return nil, err
	}
	subFlag := "-R"
	if purge {
		subFlag = "-Rns"
	}
	return p.ExecuteMutation(ctx, "remove_package", packageName, 5*time.Minute, nil, subFlag, "--noconfirm", packageName)
}

func (p *PacmanAdapter) UpgradePackage(ctx context.Context, packageName string) (*OperationResult, error) {
	if err := ValidatePackageName(packageName); err != nil {
		return nil, err
	}
	return p.ExecuteMutation(ctx, "upgrade_package", packageName, 5*time.Minute, nil, "-S", "--noconfirm", packageName)
}
