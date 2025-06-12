// Package cache provides interfaces and implementations for caching.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrCacheMiss is returned when a key is not found in the cache.
var ErrCacheMiss = errors.New("cache: key not found")

// RedisCacheInterface defines methods for interacting with a Redis cache.
type RedisCacheInterface interface {
	Get(ctx context.Context, key string, data interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// redisCache implements RedisCacheInterface using a redis.Client.
type redisCache struct {
	client *redis.Client
}

// InitializeRedisCache creates a new Redis client based on the config,
// pings to check the connection, and returns a RedisCacheInterface.
func InitializeRedisCache(cfg RedisConfig) (RedisCacheInterface, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Ping the Redis server to ensure connectivity.
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return &redisCache{client: rdb}, nil
}

// Get retrieves an item from Redis. If the key is not found, it returns ErrCacheMiss.
// The data argument should be a pointer to the variable where the unmarshaled data will be stored.
func (rc *redisCache) Get(ctx context.Context, key string, data interface{}) error {
	val, err := rc.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return err
	}

	return json.Unmarshal([]byte(val), data)
}

// Set adds an item to Redis with a specified TTL.
// The value is marshaled to JSON before storing.
func (rc *redisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return rc.client.Set(ctx, key, jsonData, ttl).Err()
}

// Delete removes an item from Redis.
func (rc *redisCache) Delete(ctx context.Context, key string) error {
	return rc.client.Del(ctx, key).Err()
}
