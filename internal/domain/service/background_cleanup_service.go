package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Shofyan/url-shortener/internal/domain/repository"
	"github.com/Shofyan/url-shortener/internal/domain/valueobject"
)

// BackgroundURLCleanupService implements the URLCleanupService interface.
// This is the "Reaper" service that handles asynchronous cleanup of expired URLs.
type BackgroundURLCleanupService struct {
	urlRepo    repository.URLRepository
	cacheRepo  repository.CacheRepository
	config     *CleanupConfig
	stats      *CleanupStats
	statsMutex sync.RWMutex
	ticker     *time.Ticker
	done       chan bool
	running    int32 // atomic flag
	wg         sync.WaitGroup
}

// NewBackgroundURLCleanupService creates a new background cleanup service.
func NewBackgroundURLCleanupService(
	urlRepo repository.URLRepository,
	cacheRepo repository.CacheRepository,
	config *CleanupConfig,
) *BackgroundURLCleanupService {
	if config == nil {
		config = DefaultCleanupConfig()
	}

	return &BackgroundURLCleanupService{
		urlRepo:   urlRepo,
		cacheRepo: cacheRepo,
		config:    config,
		stats: &CleanupStats{
			LastCleanupTime: time.Time{},
		},
		done: make(chan bool),
	}
}

// StartCleanup starts the background cleanup process.
func (s *BackgroundURLCleanupService) StartCleanup(ctx context.Context) error {
	if !s.config.Enabled {
		log.Printf("[Cleanup] Cleanup service is disabled")
		return nil
	}

	if !atomic.CompareAndSwapInt32(&s.running, 0, 1) {
		return fmt.Errorf("cleanup service is already running")
	}

	log.Printf("[Cleanup] Starting URL cleanup service with interval: %v, batch size: %d",
		s.config.CleanupInterval, s.config.BatchSize)

	s.ticker = time.NewTicker(s.config.CleanupInterval)

	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		s.cleanupLoop(ctx)
	}()

	// Update stats
	s.statsMutex.Lock()
	s.stats.IsRunning = true
	s.statsMutex.Unlock()

	return nil
}

// StopCleanup stops the background cleanup process gracefully.
func (s *BackgroundURLCleanupService) StopCleanup() error {
	if !atomic.CompareAndSwapInt32(&s.running, 1, 0) {
		return fmt.Errorf("cleanup service is not running")
	}

	log.Printf("[Cleanup] Stopping URL cleanup service...")

	if s.ticker != nil {
		s.ticker.Stop()
	}

	close(s.done)
	s.wg.Wait()

	// Update stats
	s.statsMutex.Lock()
	s.stats.IsRunning = false
	s.statsMutex.Unlock()

	log.Printf("[Cleanup] URL cleanup service stopped")

	return nil
}

// cleanupLoop runs the periodic cleanup operations.
func (s *BackgroundURLCleanupService) cleanupLoop(ctx context.Context) {
	for {
		select {
		case <-s.done:
			return
		case <-s.ticker.C:
			if atomic.LoadInt32(&s.running) == 0 {
				return
			}

			// Create context with timeout for each cleanup operation
			cleanupCtx, cancel := context.WithTimeout(ctx, s.config.MaxCleanupDuration)

			cleaned, err := s.CleanupExpiredBatch(cleanupCtx, s.config.BatchSize)

			if err != nil {
				log.Printf("[Cleanup] Cleanup batch failed: %v", err)
			} else if cleaned > 0 {
				log.Printf("[Cleanup] Cleaned up %d expired URLs", cleaned)
			}

			cancel()
		}
	}
}

// CleanupExpiredBatch performs a single cleanup batch operation.
func (s *BackgroundURLCleanupService) CleanupExpiredBatch(ctx context.Context, batchSize int) (int, error) {
	start := time.Now()

	// Calculate cutoff time with buffer to avoid clock skew issues
	cutoffTime := time.Now().Add(-s.config.BufferTime)

	log.Printf("[Cleanup] Starting batch cleanup for URLs expired before %v", cutoffTime)

	// Find expired URLs
	expiredURLs, err := s.urlRepo.FindExpiredURLs(ctx, cutoffTime, batchSize)
	if err != nil {
		duration := time.Since(start)
		s.updateStats(0, err, duration)

		return 0, fmt.Errorf("failed to find expired URLs: %w", err)
	}

	if len(expiredURLs) == 0 {
		duration := time.Since(start)
		s.updateStats(0, nil, duration)

		return 0, nil // No expired URLs to clean
	}

	// Extract short keys for batch deletion
	shortKeys := make([]*valueobject.ShortKey, len(expiredURLs))
	cacheKeys := make([]string, len(expiredURLs))

	for i, url := range expiredURLs {
		shortKeys[i] = url.ShortKey
		cacheKeys[i] = url.ShortKey.Value()
	}

	log.Printf("[Cleanup] Found %d expired URLs to delete", len(expiredURLs))

	// Delete from database in batch
	if err := s.urlRepo.DeleteExpiredBatch(ctx, shortKeys); err != nil {
		duration := time.Since(start)
		s.updateStats(0, err, duration)

		return 0, fmt.Errorf("failed to delete expired URLs from database: %w", err)
	}

	// Clean up cache entries (best effort - don't fail if cache cleanup fails)
	s.cleanupCacheEntries(ctx, cacheKeys)

	log.Printf("[Cleanup] Successfully deleted %d expired URLs", len(expiredURLs))

	duration := time.Since(start)
	s.updateStats(len(expiredURLs), nil, duration)

	return len(expiredURLs), nil
}

// cleanupCacheEntries removes expired URLs from cache.
func (s *BackgroundURLCleanupService) cleanupCacheEntries(ctx context.Context, cacheKeys []string) {
	for _, key := range cacheKeys {
		if err := s.cacheRepo.Delete(ctx, key); err != nil {
			// Log but don't fail - cache cleanup is best effort
			log.Printf("[Cleanup] Warning: Failed to delete cache key %s: %v", key, err)
		}
	}
}

// updateStats updates the cleanup statistics.
func (s *BackgroundURLCleanupService) updateStats(cleaned int, err error, duration time.Duration) {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()

	s.stats.LastCleanupTime = time.Now()
	s.stats.LastBatchSize = cleaned
	s.stats.TotalCleaned += int64(cleaned)

	if err != nil {
		s.stats.FailedRuns++
	} else {
		s.stats.SuccessfulRuns++
	}

	// Update average cleanup time using exponential moving average
	if s.stats.AverageCleanupMs == 0 {
		s.stats.AverageCleanupMs = float64(duration.Nanoseconds()) / 1e6
	} else {
		// EMA with alpha = 0.1
		alpha := 0.1
		currentMs := float64(duration.Nanoseconds()) / 1e6
		s.stats.AverageCleanupMs = alpha*currentMs + (1-alpha)*s.stats.AverageCleanupMs
	}
}

// GetCleanupStats returns statistics about cleanup operations.
func (s *BackgroundURLCleanupService) GetCleanupStats() *CleanupStats {
	s.statsMutex.RLock()
	defer s.statsMutex.RUnlock()

	// Return a copy to avoid race conditions
	statsCopy := *s.stats

	return &statsCopy
}
