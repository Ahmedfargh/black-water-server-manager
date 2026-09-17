package PackageManagers

import "time"

// ManagerCategory differentiates between native OS package managers and universal formats
type ManagerCategory string

const (
	CategoryNative    ManagerCategory = "native"
	CategoryUniversal ManagerCategory = "universal"
)

// PackageItem represents an installed package
type PackageItem struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
	Source      string `json:"source,omitempty"`
}

// PackageUpdate represents an available update for an installed package
type PackageUpdate struct {
	Name           string `json:"name"`
	CurrentVersion string `json:"current_version"`
	NewVersion     string `json:"new_version"`
	Repository     string `json:"repository,omitempty"`
}

// ManagerMetadata contains information about a detected package manager
type ManagerMetadata struct {
	Name           string          `json:"name"`
	Category       ManagerCategory `json:"category"`
	IsPrimary      bool            `json:"is_primary"`
	Version        string          `json:"version"`
	ExecutablePath string          `json:"executable_path"`
	TotalPackages  int             `json:"total_packages"`
	PendingUpdates int             `json:"pending_updates"`
	Available      bool            `json:"available"`
}

// HostOSInfo encapsulates host distribution details
type HostOSInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PrettyName string `json:"pretty_name"`
	VersionID  string `json:"version_id,omitempty"`
	IDLike     string `json:"id_like,omitempty"`
}

// PackagesOverview encapsulates the overall system package status
type PackagesOverview struct {
	OS             HostOSInfo        `json:"os"`
	PrimaryManager string            `json:"primary_manager"`
	Managers       []ManagerMetadata `json:"managers"`
	GeneratedAt    time.Time         `json:"generated_at"`
}

// OperationResult standardizes output from package manager mutations
type OperationResult struct {
	Success         bool      `json:"success"`
	Action          string    `json:"action"`
	Manager         string    `json:"manager"`
	TargetPackage   string    `json:"target_package,omitempty"`
	Command         string    `json:"command"`
	Output          string    `json:"output"`
	ExecutionTimeMs int64     `json:"execution_time_ms"`
	ExecutedAt      time.Time `json:"executed_at"`
	Message         string    `json:"message"`
}

// PackageActionRequest encapsulates request payload for package operations
type PackageActionRequest struct {
	Package string `json:"package" binding:"required"`
	Purge   bool   `json:"purge"`
}
