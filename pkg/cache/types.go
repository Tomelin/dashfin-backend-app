// Package cache defines common types and interfaces for caching mechanisms.
package cache

// RedisConfig holds the configuration for connecting to a Redis server.
type RedisConfig struct {
	Address           string `yaml:"address"`             // Redis server address (e.g., "localhost:6379")
	Password          string `yaml:"password"`            // Redis password (leave empty if none)
	DB                int    `yaml:"db"`                  // Redis database number (e.g., 0)
	DefaultTTLSeconds int    `yaml:"default_ttl_seconds"` // Default TTL for cache entries in seconds
}
