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

	// SetCacheEntry stores a structured cache entry with metadata
	SetCacheEntry(ctx context.Context, key string, entry *CacheEntry, ttl time.Duration) error

	// GetCacheEntry retrieves a structured cache entry
	GetCacheEntry(ctx context.Context, key string) (*CacheEntry, error)

	// SetTombstone stores a tombstone marker for an expired/deleted URL
	SetTombstone(ctx context.Context, key string, reason string, ttl time.Duration) error
}

// CacheEntry represents a structured cache entry with metadata.
type CacheEntry struct {
	LongURL     string     `json:"long_url"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	IsTombstone bool       `json:"is_tombstone"`
	Reason      string     `json:"reason,omitempty"` // For tombstones: "expired", "deleted", etc.
}

// IsExpired checks if the cache entry is logically expired.
func (e *CacheEntry) IsExpired() bool {
	if e.ExpiresAt == nil {
		return false
	}

	return time.Now().After(*e.ExpiresAt)
}
