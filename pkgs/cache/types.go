package cache

import (
	"fmt"
	"time"
)

type AdapterCache interface {
	String() string
	Get(key string) (string, error)
	Set(key string, val any, expire time.Duration) error
	Del(key string) error
	HashGet(hk, key string) (string, error)
	HashDel(hk, key string) error
	Increase(key string) error
	Decrease(key string) error
	Expire(key string, dur time.Duration) error
	GetClient() any
}

type KeyNotFoundError struct {
	Key string
}

func (e *KeyNotFoundError) Error() string {
	return fmt.Sprintf("key not found: %s", e.Key)
}

func NewKeyNotFoundError(key string) *KeyNotFoundError {
	return &KeyNotFoundError{Key: key}
}

func IsKeyNotFoundError(err error) bool {
	_, ok := err.(*KeyNotFoundError)
	return ok
}
