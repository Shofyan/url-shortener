package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/Shofyan/url-shortener/internal/domain/repository"
)

var (
	// ErrCacheMiss is returned when a key is not found in the cache.
	ErrCacheMiss = errors.New("cache miss")
)

// CacheRepository implements the CacheRepository interface for Redis.
type CacheRepository struct {
	client *redis.Client
}

// NewCacheRepository creates a new Redis cache repository.
func NewCacheRepository(client *redis.Client) *CacheRepository {
	return &CacheRepository{
		client: client,
	}
}

// Set stores a key-value pair with TTL.
func (r *CacheRepository) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Get retrieves a value by key.
func (r *CacheRepository) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrCacheMiss
	}

	if err != nil {
		return "", err
	}

	return val, nil
}

// Delete removes a key from cache.
func (r *CacheRepository) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Exists checks if a key exists in cache.
func (r *CacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// SetCacheEntry stores a structured cache entry with metadata.
func (r *CacheRepository) SetCacheEntry(ctx context.Context, key string, entry *repository.CacheEntry, ttl time.Duration) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCacheEntry retrieves a structured cache entry.
func (r *CacheRepository) GetCacheEntry(ctx context.Context, key string) (*repository.CacheEntry, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, ErrCacheMiss
	}

	if err != nil {
		return nil, err
	}

	var entry repository.CacheEntry
	if err := json.Unmarshal([]byte(val), &entry); err != nil {
		// If JSON parsing fails, it might be a legacy simple string cache entry
		// Fall back to creating a simple cache entry
		return &repository.CacheEntry{
			LongURL:     val,
			CreatedAt:   time.Now(),
			IsTombstone: false,
		}, nil
	}

	return &entry, nil
}

// SetTombstone stores a tombstone marker for an expired/deleted URL.
func (r *CacheRepository) SetTombstone(ctx context.Context, key, reason string, ttl time.Duration) error {
	tombstone := &repository.CacheEntry{
		IsTombstone: true,
		Reason:      reason,
		CreatedAt:   time.Now(),
	}

	return r.SetCacheEntry(ctx, key, tombstone, ttl)
}

// NewRedisClient creates a new Redis client.
func NewRedisClient(addr, password string, db, poolSize, minIdleConns int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
