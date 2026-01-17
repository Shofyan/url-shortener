// Package service contains domain services for business logic.
package service

import (
	"context"
	"time"
)

// URLCleanupService defines the interface for URL cleanup operations.
// This service implements the "Reaper" pattern for asynchronous expired URL cleanup.
type URLCleanupService interface {
	// StartCleanup starts the background cleanup process.
	// It runs periodic cleanup jobs to remove expired URLs in batches.
	StartCleanup(ctx context.Context) error

	// StopCleanup stops the background cleanup process gracefully.
	StopCleanup() error

	// CleanupExpiredBatch performs a single cleanup batch operation.
	// Returns the number of records cleaned up and any error.
	CleanupExpiredBatch(ctx context.Context, batchSize int) (int, error)

	// GetCleanupStats returns statistics about cleanup operations.
	GetCleanupStats() *CleanupStats
}

// CleanupStats contains statistics about cleanup operations.
type CleanupStats struct {
	LastCleanupTime  time.Time `json:"last_cleanup_time"`
	TotalCleaned     int64     `json:"total_cleaned"`
	LastBatchSize    int       `json:"last_batch_size"`
	SuccessfulRuns   int64     `json:"successful_runs"`
	FailedRuns       int64     `json:"failed_runs"`
	AverageCleanupMs float64   `json:"average_cleanup_ms"`
	IsRunning        bool      `json:"is_running"`
}

// CleanupConfig contains configuration for the cleanup service.
type CleanupConfig struct {
	// Interval between cleanup runs
	CleanupInterval time.Duration `json:"cleanup_interval"`

	// Maximum number of records to delete in one batch
	BatchSize int `json:"batch_size"`

	// Buffer time before deleting expired records (prevents clock skew issues)
	BufferTime time.Duration `json:"buffer_time"`

	// Maximum duration for a single cleanup operation
	MaxCleanupDuration time.Duration `json:"max_cleanup_duration"`

	// Enable cleanup service (allows disabling in production if needed)
	Enabled bool `json:"enabled"`
}

// DefaultCleanupConfig returns sensible defaults for cleanup configuration.
func DefaultCleanupConfig() *CleanupConfig {
	return &CleanupConfig{
		CleanupInterval:    15 * time.Minute, // Run every 15 minutes
		BatchSize:          1000,             // Process 1000 records per batch
		BufferTime:         1 * time.Hour,    // Delete only after 1 hour past expiration
		MaxCleanupDuration: 5 * time.Minute,  // Maximum 5 minutes per cleanup run
		Enabled:            true,             // Enabled by default
	}
}
