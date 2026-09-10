package PackageManagers

import "sync"

// ManagerRegistry maintains registered package manager adapters (Open/Closed Principle)
type ManagerRegistry struct {
	mu       sync.RWMutex
	adapters map[string]PackageManagerAdapter
}

func NewManagerRegistry() *ManagerRegistry {
	r := &ManagerRegistry{
		adapters: make(map[string]PackageManagerAdapter),
	}
	r.registerDefaults()
	return r
}

// Register adds or replaces an adapter in the registry
func (r *ManagerRegistry) Register(adapter PackageManagerAdapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[adapter.GetName()] = adapter
}

// Get retrieves an adapter by name
func (r *ManagerRegistry) Get(name string) (PackageManagerAdapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	adapter, ok := r.adapters[name]
	return adapter, ok
}

// GetAll returns a list of all registered adapters
func (r *ManagerRegistry) GetAll() []PackageManagerAdapter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]PackageManagerAdapter, 0, len(r.adapters))
	for _, a := range r.adapters {
		list = append(list, a)
	}
	return list
}

// registerDefaults seeds standard package managers
func (r *ManagerRegistry) registerDefaults() {
	r.Register(NewPacmanAdapter())
	r.Register(NewAptAdapter())
	r.Register(NewDnfAdapter())
	r.Register(NewSnapAdapter())
	r.Register(NewFlatpakAdapter())
	r.Register(NewZypperAdapter())
	r.Register(NewApkAdapter())
}
