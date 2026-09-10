package PackageManagers

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PackageManagerService orchestrates detection, caching, and querying
type PackageManagerService struct {
	detector SystemDetector
	registry *ManagerRegistry

	cacheMu     sync.RWMutex
	cachedAt    time.Time
	cachedTTL   time.Duration
	cachedStats *PackagesOverview
}

func NewPackageManagerService(detector SystemDetector, registry *ManagerRegistry) *PackageManagerService {
	if detector == nil {
		detector = NewLinuxSystemDetector()
	}
	if registry == nil {
		registry = NewManagerRegistry()
	}
	return &PackageManagerService{
		detector:  detector,
		registry:  registry,
		cachedTTL: 30 * time.Second,
	}
}

// GetSystemOverview returns the detected host OS and available package managers with metrics
func (s *PackageManagerService) GetSystemOverview(ctx context.Context, forceRefresh bool) (*PackagesOverview, error) {
	s.cacheMu.RLock()
	if !forceRefresh && s.cachedStats != nil && time.Since(s.cachedAt) < s.cachedTTL {
		defer s.cacheMu.RUnlock()
		return s.cachedStats, nil
	}
	s.cacheMu.RUnlock()

	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	// Double-check locking
	if !forceRefresh && s.cachedStats != nil && time.Since(s.cachedAt) < s.cachedTTL {
		return s.cachedStats, nil
	}

	osInfo := s.detector.DetectOS()
	primary := s.detector.ResolvePrimaryManager(osInfo)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var detectedManagers []ManagerMetadata

	adapters := s.registry.GetAll()
	for _, adapter := range adapters {
		if !adapter.IsAvailable() {
			continue
		}

		wg.Add(1)
		go func(ad PackageManagerAdapter) {
			defer wg.Done()
			ver, _ := ad.GetVersion(ctx)
			count, _ := ad.GetInstalledCount(ctx)

			meta := ManagerMetadata{
				Name:           ad.GetName(),
				Category:       ad.GetCategory(),
				IsPrimary:      ad.GetName() == primary,
				Version:        ver,
				ExecutablePath: ad.GetExecutablePath(),
				TotalPackages:  count,
				Available:      true,
			}

			mu.Lock()
			detectedManagers = append(detectedManagers, meta)
			mu.Unlock()
		}(adapter)
	}

	wg.Wait()

	overview := &PackagesOverview{
		OS:             osInfo,
		PrimaryManager: primary,
		Managers:       detectedManagers,
		GeneratedAt:    time.Now(),
	}

	s.cachedStats = overview
	s.cachedAt = time.Now()

	return overview, nil
}

// GetManagerUpdates fetches pending updates for a specific manager
func (s *PackageManagerService) GetManagerUpdates(ctx context.Context, name string) ([]PackageUpdate, error) {
	adapter, ok := s.registry.Get(name)
	if !ok || !adapter.IsAvailable() {
		return nil, fmt.Errorf("package manager '%s' not found or unavailable on host", name)
	}
	return adapter.GetPendingUpdates(ctx)
}

// GetInstalledPackages fetches paginated and filtered packages for a specific manager
func (s *PackageManagerService) GetInstalledPackages(ctx context.Context, name string, query string, page int, limit int) ([]PackageItem, int, error) {
	adapter, ok := s.registry.Get(name)
	if !ok || !adapter.IsAvailable() {
		return nil, 0, fmt.Errorf("package manager '%s' not found or unavailable on host", name)
	}
	return adapter.GetInstalledPackages(ctx, query, page, limit)
}
