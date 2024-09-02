package bimap

import "sync"

// BiMap is a generic bidirectional map.
type BiMap[k comparable, v comparable] struct {
	lock sync.RWMutex
	kmap map[k]v
	vmap map[v]k
}

// NewBiMap initializes a new instance of BiMap.
func NewBiMap[k comparable, v comparable]() *BiMap[k, v] {
	return &BiMap[k, v]{
		kmap: make(map[k]v),
		vmap: make(map[v]k),
	}
}

// Insert adds or updates key-value pairs in both directions.
func (b *BiMap[k, v]) Insert(key k, value v) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if oldValue, ok := b.kmap[key]; ok {
		if oldValue == value {
			return
		}
		delete(b.vmap, oldValue)
	}

	b.kmap[key] = value
	b.vmap[value] = key
}

// GetByKey retrieves a value by key.
func (b *BiMap[k, v]) GetKey(key k) (v, bool) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	value, ok := b.kmap[key]
	return value, ok
}

// GetByValue retrieves a key by value.
func (b *BiMap[k, v]) GetValue(value v) (k, bool) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	key, ok := b.vmap[value]
	return key, ok
}
func (b *BiMap[k, v]) DelKey(key k) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if value, ok := b.kmap[key]; ok {
		delete(b.kmap, key)
		delete(b.vmap, value)
	}
}

// RemoveByValue 通过值删除映射
func (b *BiMap[k, v]) DelValue(value v) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if key, ok := b.vmap[value]; ok {
		delete(b.kmap, key)
		delete(b.vmap, value)
	}
}
