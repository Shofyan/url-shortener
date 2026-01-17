package repository

import (
	"context"
	"time"
)

// CacheRepository defines the interface for caching operations.
type CacheRepository interface {
	// Set stores a key-value pair with TTL
	Set(ctx context.Context, key string, value string, ttl time.Duration) error

	// Get retrieves a value by key
	Get(ctx context.Context, key string) (string, error)

	// Delete removes a key from cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in cache
	Exists(ctx context.Context, key string) (bool, error)
}
