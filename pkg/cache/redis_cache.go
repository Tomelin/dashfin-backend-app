package cache

import (
	"context"
	"time"
	// "errors" // Not needed here as ErrNotFound is used from this package (cache.ErrNotFound)

	"github.com/go-redis/redis/v8"
)

type redisCacheService struct {
	client *redis.Client
}

// NewRedisCacheService creates a new CacheService using Redis.
// It returns the CacheService interface, not the concrete type.
func NewRedisCacheService(client *redis.Client) CacheService {
	return &redisCacheService{client: client}
}

func (r *redisCacheService) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrNotFound // Use the exported ErrNotFound from this package
	} else if err != nil {
		return "", err
	}
	return val, nil
}

func (r *redisCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *redisCacheService) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
