package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache is the interface for cache operations
type Cache interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// redisCache is the Redis implementation of Cache
type redisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(client *redis.Client) Cache {
	return &redisCache{client: client}
}

func (c *redisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *redisCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *redisCache) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// NoOpCache is a no-op cache for when Redis is not available
type NoOpCache struct{}

func NewNoOpCache() Cache { return &NoOpCache{} }
func (c *NoOpCache) Set(_ context.Context, _ string, _ interface{}, _ time.Duration) error {
	return nil
}
func (c *NoOpCache) Get(_ context.Context, _ string) (string, error)  { return "", nil }
func (c *NoOpCache) Del(_ context.Context, _ string) error            { return nil }
func (c *NoOpCache) Exists(_ context.Context, _ string) (bool, error) { return false, nil }
