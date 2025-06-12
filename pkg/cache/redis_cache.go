package cache

import (
	"context"
	"fmt"
	"time"

	// "errors" // Not needed here as ErrNotFound is used from this package (cache.ErrNotFound)

	"github.com/go-redis/redis/v8"
)

type redisCacheService struct {
	client *redis.Client
}

type CacheClient *redis.Client

// NewRedisCacheService creates a new CacheService using Redis.
// It returns the CacheService interface, not the concrete type.
func NewRedisCacheService(cfg CacheConfig) (CacheService, error) {

	var client redisCacheService
	client.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Address, cfg.Port),
		Password: cfg.Password,
		Username: cfg.Username,
		DB:       cfg.DB,
	})

	err := client.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	return &redisCacheService{client: client.client}, nil
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

func (r *redisCacheService) Ping(ctx context.Context) error {
	_, err := r.client.Ping(ctx).Result()
	return err

}
