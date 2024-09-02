package casbin

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements Cache interface using Redis as the backend.
type RedisCache struct {
	client *redis.Client
	prefix string
	expire time.Duration
}

type opt func(*RedisCache)

func WithPrefix(p string) opt {
	return func(rc *RedisCache) {
		rc.prefix = p
	}
}

func WithExpire(t time.Duration) opt {
	return func(rc *RedisCache) {
		rc.expire = t
	}
}

// NewRedisCache creates a new RedisCache with given Redis options and a prefix for all keys.
func NewRedisCache(client *redis.Client, opts ...opt) *RedisCache {

	c := &RedisCache{
		client: client,
		prefix: "casbin",
		expire: 10 * time.Minute,
	}

	for _, f := range opts {
		f(c)
	}
	return c
}

// setKey adds prefix to the key to namespace it.
func (c *RedisCache) setKey(key string) string {
	return c.prefix + ":" + key
}

// Set puts key and value into cache.
func (c *RedisCache) Set(key string, value bool, extra ...interface{}) error {
	var expiration time.Duration = c.expire
	if len(extra) > 0 {
		if exp, ok := extra[0].(time.Duration); ok {
			expiration = exp
		}
	}
	return c.client.Set(context.TODO(), c.setKey(key), value, expiration).Err()
}

// Get returns result for key.
func (c *RedisCache) Get(key string) (bool, error) {
	key = c.setKey(key)
	return c.client.Get(context.TODO(), key).Bool()
}

// Delete will remove the specific key in cache.
func (c *RedisCache) Delete(key string) error {
	key = c.setKey(key)
	return c.client.Del(context.TODO(), key).Err()
}

// Clear deletes all the items stored in cache that match the prefix.
func (c *RedisCache) Clear() error {
	pattern := c.prefix + ":*"
	ctx := context.TODO()
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}
