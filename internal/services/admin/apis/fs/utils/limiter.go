package utils

import (
	"fmt"
	"go-file-server/pkgs/utils/limiter"
	"time"

	"github.com/patrickmn/go-cache"
)

type LimiterManager struct {
	limiters *cache.Cache
}

func NewLimiterManager(defaultExpiration, cleanupInterval time.Duration) *LimiterManager {
	return &LimiterManager{
		limiters: cache.New(defaultExpiration, cleanupInterval),
	}
}

func (m *LimiterManager) GetLimiter(key string, rateLimitBytes uint64) *limiter.Limiter {
	key = fmt.Sprintf("%s-%d", key, rateLimitBytes)
	if lim, found := m.limiters.Get(key); found {
		return lim.(*limiter.Limiter)
	}

	newLimiter := m.createLimiter(rateLimitBytes)
	if added := m.limiters.Add(key, newLimiter, cache.DefaultExpiration); added == nil {
		return newLimiter
	}

	lim, _ := m.limiters.Get(key)
	return lim.(*limiter.Limiter)
}

func (m *LimiterManager) createLimiter(rateLimitBytes uint64) *limiter.Limiter {
	return limiter.NewLimiter(rateLimitBytes, rateLimitBytes)
}

type IdManager struct {
	id *cache.Cache
}

func NewIdManager(defaultExpiration, cleanupInterval time.Duration) *IdManager {
	return &IdManager{
		id: cache.New(defaultExpiration, cleanupInterval),
	}
}

func (m *IdManager) GetID(key string) (string, bool) {
	id, ok := m.id.Get(key)
	if ok {
		return id.(string), true
	}
	return "", false
}

func (m *IdManager) GetOrCreateID(key string, nid string) string {
	if added := m.id.Add(key, nid, cache.DefaultExpiration); added == nil {
		return nid
	}

	id, _ := m.id.Get(key)
	return id.(string)
}
