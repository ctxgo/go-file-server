package cache

import (
	"errors"
	"go-file-server/pkgs/utils/str"
	"strconv"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

// NewMemory memory模式
func NewMemory() *Memory {
	return &Memory{
		Cache: cache.New(0, 0),
	}
}

type Memory struct {
	*cache.Cache
	sync.RWMutex
}

func (*Memory) String() string {
	return "memory"
}

func (m *Memory) Connect() {
}

func (m *Memory) GetClient() any {
	return m
}

func (m *Memory) Set(key string, val any, expire time.Duration) error {
	s, err := str.ConvertToString(val)
	if err != nil {
		return err
	}
	m.Cache.Set(key, s, expire)
	return nil
}

func (m *Memory) Get(key string) (string, error) {
	v, ok := m.Cache.Get(key)
	if !ok {
		return "", NewKeyNotFoundError(key)
	}
	if v, ok := v.(string); ok {
		return v, nil
	}
	return "", errors.New(key + " type error")
}

func (m *Memory) Del(key string) error {
	m.Cache.Delete(key)
	return nil
}

func (m *Memory) HashGet(hk, key string) (string, error) {
	return m.Get(hk + key)
}

func (m *Memory) HashDel(hk, key string) error {
	return m.Del(hk + key)
}

func (m *Memory) Increase(key string) error {
	return m.changeValue(key, 1)
}

func (m *Memory) Decrease(key string) error {
	return m.changeValue(key, -1)
}

func (m *Memory) changeValue(key string, delta int) error {
	m.Lock()
	defer m.Unlock()

	v, ok := m.Cache.Get(key)
	if !ok {
		m.Cache.Set(key, strconv.Itoa(delta), cache.NoExpiration)
		return nil
	}

	currentVal, err := strconv.Atoi(v.(string))
	if err != nil {
		return err
	}

	newVal := currentVal + delta
	m.Cache.Set(key, strconv.Itoa(newVal), cache.NoExpiration)
	return nil
}

func (m *Memory) Expire(key string, dur time.Duration) error {
	m.RLock()
	defer m.RUnlock()
	v, err := m.Get(key)
	if err != nil {
		return err
	}
	return m.Set(key, v, dur)
}
