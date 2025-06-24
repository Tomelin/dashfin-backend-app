package cache

import (
	"context"
	"errors" // Added import for errors.New
	"time"
)

// CacheService defines the interface for a cache.
type CacheService interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Ping(ctx context.Context) error
}

type CacheConfig struct {
	Address  string      `mapstructure:"address" json:"address"`
	Password string      `mapstructure:"password" json:"password"`
	DB       int         `mapstructure:"db" json:"db"`
	Port     interface{} `mapstructure:"port" json:"port"`
	Username string      `mapstructure:"username" json:"username"`
}

// ErrNotFound is returned when an item is not found in the cache.
var ErrNotFound = errors.New("cache: item not found")
