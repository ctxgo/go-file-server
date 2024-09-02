package system

import (
	"go-file-server/pkgs/cache"
	"sync"
)

type SystemApi struct {
	mutex sync.RWMutex
	cache cache.AdapterCache
}

func NewSystemApi(
	cache cache.AdapterCache,
) *SystemApi {
	return &SystemApi{
		cache: cache,
	}
}
