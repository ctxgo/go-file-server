package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedis redis模式
func NewRedis(options *redis.Options) (*Redis, error) {

	r := &Redis{
		client: redis.NewClient(options),
	}
	err := r.connect()
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Redis cache implement
type Redis struct {
	client *redis.Client
}

func (*Redis) String() string {
	return "redis"
}

// connect connect test
func (r *Redis) connect() error {
	var err error
	_, err = r.client.Ping(context.TODO()).Result()
	return err
}

// Set value with key and expire time
func (r *Redis) Set(key string, val any, expire time.Duration) error {
	return r.client.Set(context.TODO(), key, val, expire).Err()
}

// Get from key
func (r *Redis) Get(key string) (string, error) {
	s, err := r.client.Get(context.TODO(), key).Result()
	if err != nil && err == redis.Nil {
		return "", NewKeyNotFoundError(key)
	}
	return s, err
}

// Del delete key in redis
func (r *Redis) Del(key string) error {
	return r.client.Del(context.TODO(), key).Err()
}

// HashGet from key
func (r *Redis) HashGet(hk, key string) (string, error) {
	return r.client.HGet(context.TODO(), hk, key).Result()
}

// HashDel delete key in specify redis's hashtable
func (r *Redis) HashDel(hk, key string) error {
	return r.client.HDel(context.TODO(), hk, key).Err()
}

// Increase
func (r *Redis) Increase(key string) error {
	return r.client.Incr(context.TODO(), key).Err()
}

func (r *Redis) Decrease(key string) error {
	return r.client.Decr(context.TODO(), key).Err()
}

// Set ttl
func (r *Redis) Expire(key string, dur time.Duration) error {
	return r.client.Expire(context.TODO(), key, dur).Err()
}

// GetClient 暴露原生client
func (r *Redis) GetClient() any {
	return r.client
}
