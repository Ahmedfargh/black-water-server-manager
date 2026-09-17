package PackageManagers

import (
	"context"
)

// PackageInspector defines capabilities to inspect version and installed count (ISP)
type PackageInspector interface {
	GetVersion(ctx context.Context) (string, error)
	GetInstalledCount(ctx context.Context) (int, error)
}

// PackageQueryer defines capabilities to list and search installed packages (ISP)
type PackageQueryer interface {
	GetInstalledPackages(ctx context.Context, query string, page int, limit int) ([]PackageItem, int, error)
}

// PackageUpdateChecker defines capabilities to query pending updates (ISP)
type PackageUpdateChecker interface {
	GetPendingUpdates(ctx context.Context) ([]PackageUpdate, error)
}

// PackageLifecycleManager defines capabilities for package installation, removal, updates, and cache cleaning (ISP)
type PackageLifecycleManager interface {
	CleanCache(ctx context.Context) (*OperationResult, error)
	RefreshRepositories(ctx context.Context) (*OperationResult, error)
	UpgradeSystem(ctx context.Context) (*OperationResult, error)
	InstallPackage(ctx context.Context, packageName string) (*OperationResult, error)
	RemovePackage(ctx context.Context, packageName string, purge bool) (*OperationResult, error)
	UpgradePackage(ctx context.Context, packageName string) (*OperationResult, error)
}

// PackageManagerAdapter provides a unified contract for concrete adapters (LSP & DIP)
type PackageManagerAdapter interface {
	PackageInspector
	PackageQueryer
	PackageUpdateChecker
	PackageLifecycleManager

	GetName() string
	GetCategory() ManagerCategory
	IsAvailable() bool
	GetExecutablePath() string
}
