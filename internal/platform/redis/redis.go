package redis

import (
	"context"
	"fmt"
	"log"

	"fixio/pkg/cache"
	"fixio/pkg/config"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates a new Redis client connection
func NewRedisClient(env *config.Environment) (cache.Cache, *redis.Client) {
	if env.RedisHost == "" {
		log.Println("⚠️  Redis not configured, using NoOp cache")
		return cache.NewNoOpCache(), nil
	}

	addr := fmt.Sprintf("%s:%s", env.RedisHost, env.RedisPort)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: env.RedisPassword,
		DB:       env.RedisDB,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("⚠️  Redis connection failed: %v, using NoOp cache", err)
		return cache.NewNoOpCache(), nil
	}

	log.Println("✅ Redis connected successfully")
	return cache.NewRedisCache(client), client
}
