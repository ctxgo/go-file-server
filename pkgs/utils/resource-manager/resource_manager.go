package resourcemanager

import (
	"sync"
)

// ResourceManager is a generic manager for creating, storing, and accessing instances of type T.
type ResourceManager[T any] struct {
	resources    map[string]T // Map of keys to instances of T
	sync.RWMutex              // Synchronization for concurrent access
}

// NewResourceManager creates a new ResourceManager with the given factory.
func NewResourceManager[T any]() *ResourceManager[T] {
	return &ResourceManager[T]{
		resources: make(map[string]T),
	}
}

// Get retrieves an instance by key. Returns the instance and a boolean indicating its existence.
func (rm *ResourceManager[T]) Get(k string) (T, bool) {
	rm.RLock()
	defer rm.RUnlock()
	return rm.get(k)
}

func (rm *ResourceManager[T]) Set(k string, v T) {
	rm.Lock()
	defer rm.Unlock()
	rm.resources[k] = v
}

// GetOrSet retrieves an instance by key, or sets and returns the provided value if key doesn't exist.
// Returns the value (existing or newly set) and a boolean indicating whether the key existed before.
func (rm *ResourceManager[T]) GetOrSet(k string, v T) (T, bool) {
	rm.Lock()
	defer rm.Unlock()

	if existing, exists := rm.resources[k]; exists {
		return existing, true
	}

	rm.resources[k] = v
	return v, false
}

// Del deletes an instance by key.
func (rm *ResourceManager[T]) Del(k string) {
	rm.Lock()
	defer rm.Unlock()
	delete(rm.resources, k)
}

// get retrieves an instance from the internal map without locking.
func (rm *ResourceManager[T]) get(k string) (T, bool) {
	v, ok := rm.resources[k]
	return v, ok
}
