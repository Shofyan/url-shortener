package concurrency_test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Shofyan/url-shortener/internal/domain/entity"
	"github.com/Shofyan/url-shortener/internal/domain/valueobject"
	"github.com/Shofyan/url-shortener/internal/infrastructure/database/postgres"
)

// TestConcurrentClickCounting tests thread-safe visit count increment under high concurrency.
func TestConcurrentClickCounting(t *testing.T) {
	// Skip this test if not running integration tests
	if testing.Short() {
		t.Skip("Skipping concurrency test in short mode")
	}

	// Setup test database connection
	// Note: This would require a test database - in practice you'd use testcontainers or similar
	// For now, this demonstrates the test structure
	db, err := setupTestDB()
	if err != nil {
		// print error for debugging
		t.Logf("Error setting up test database: %v", err)
		t.Skip("Database not available for concurrency test")
	}

	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("Failed to close database: %v", err)
		}
	}()

	repo := postgres.NewURLRepository(db)
	ctx := context.Background()

	// Setup test data
	shortKey, err := valueobject.NewShortKey("test123")
	require.NoError(t, err)

	// Insert initial URL for testing
	err = insertTestURL(ctx, repo, shortKey)
	require.NoError(t, err)

	// Get initial visit count
	initialURL, err := repo.FindByShortKey(ctx, shortKey)
	require.NoError(t, err)

	initialCount := initialURL.VisitCount

	// Test concurrent increments
	concurrentRequests := 100

	var wg sync.WaitGroup

	errorCount := 0

	var errorMutex sync.Mutex

	// Launch concurrent goroutines
	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)

		go func(requestID int) {
			defer wg.Done()

			if err := repo.IncrementVisitCount(ctx, shortKey); err != nil {
				errorMutex.Lock()
				errorCount++
				errorMutex.Unlock()
				t.Logf("Request %d failed: %v", requestID, err)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify results
	finalURL, err := repo.FindByShortKey(ctx, shortKey)
	require.NoError(t, err)

	expectedCount := initialCount + int64(concurrentRequests)
	assert.Equal(t, expectedCount, finalURL.VisitCount,
		"Visit count should be exactly %d after %d concurrent increments",
		expectedCount, concurrentRequests)

	assert.Equal(t, 0, errorCount, "No errors should occur during concurrent increments")

	// Verify last_accessed_at was updated (just ensure it's not nil)
	assert.NotNil(t, finalURL.LastAccessedAt, "LastAccessedAt should be set after increment")

	t.Logf("Concurrency test completed successfully:")
	t.Logf("  - Concurrent requests: %d", concurrentRequests)
	t.Logf("  - Initial count: %d", initialCount)
	t.Logf("  - Final count: %d", finalURL.VisitCount)
	t.Logf("  - Expected count: %d", expectedCount)
	t.Logf("  - Errors: %d", errorCount)
}

// TestConcurrentClickCountingWithRetries tests resilience under extreme load.
func TestConcurrentClickCountingWithRetries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping extreme concurrency test in short mode")
	}

	db, err := setupTestDB()
	if err != nil {
		t.Skip("Database not available for concurrency test")
	}

	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("Failed to close database: %v", err)
		}
	}()

	repo := postgres.NewURLRepository(db)
	ctx := context.Background()

	shortKey, err := valueobject.NewShortKey("extreme123")
	require.NoError(t, err)

	err = insertTestURL(ctx, repo, shortKey)
	require.NoError(t, err)

	// Test with more concurrent requests but reasonable limits
	concurrentRequests := 200
	maxRetries := 3

	var wg sync.WaitGroup

	successCount := int64(0)

	var successMutex sync.Mutex

	start := time.Now()

	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)

		go func(requestID int) {
			defer wg.Done()

			// Retry logic for database contention
			for retry := 0; retry < maxRetries; retry++ {
				if err := repo.IncrementVisitCount(ctx, shortKey); err != nil {
					if retry == maxRetries-1 {
						t.Logf("Request %d failed after %d retries: %v", requestID, maxRetries, err)
						return
					}
					// Brief backoff before retry
					time.Sleep(time.Millisecond * time.Duration(retry+1))

					continue
				}

				successMutex.Lock()
				successCount++
				successMutex.Unlock()

				break
			}
		}(i)
	}

	wg.Wait()

	duration := time.Since(start)

	// Verify that most requests succeeded (lower threshold for extreme load)
	successRate := float64(successCount) / float64(concurrentRequests) * 100
	assert.Greater(t, successRate, 80.0,
		"Success rate should be above 80%% under extreme load")

	t.Logf("Extreme concurrency test results:")
	t.Logf("  - Total requests: %d", concurrentRequests)
	t.Logf("  - Successful requests: %d", successCount)
	t.Logf("  - Success rate: %.2f%%", successRate)
	t.Logf("  - Duration: %v", duration)
	t.Logf("  - Requests per second: %.2f", float64(concurrentRequests)/duration.Seconds())
}

// setupTestDB creates a test database connection.
func setupTestDB() (*sql.DB, error) {
	// In a real implementation, you would:
	// 1. Use testcontainers to spin up a PostgreSQL instance
	// 2. Run migrations to create the schema
	// 3. Return the connection
	//
	// For this example, we'll try to connect to a local test database
	//       - DATABASE_USER=${POSTGRES_USER:-postgres}
	//   - DATABASE_PASSWORD=${POSTGRES_PASSWORD:-postgres}
	//   - DATABASE_DBNAME=${POSTGRES_DB:-urlshortener}
	// Get database connection parameters from environment variables with defaults
	dbUser := getEnv("POSTGRES_USER", "postgres")
	dbPassword := getEnv("POSTGRES_PASSWORD", "postgres")
	dbName := getEnv("POSTGRES_DB", "urlshortener")
	dbHost := getEnv("POSTGRES_HOST", "localhost")
	dbPort := getEnv("POSTGRES_PORT", "5432")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Configure connection pool for concurrency tests
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(time.Minute * 5)

	if err := db.Ping(); err != nil {
		_ = db.Close() // Explicitly ignore error during cleanup
		return nil, err
	}

	return db, nil
}

// insertTestURL inserts a test URL into the database using the repository.
func insertTestURL(ctx context.Context, repo *postgres.URLRepository, shortKey *valueobject.ShortKey) error {
	longURL, err := valueobject.NewLongURL("https://example.com/concurrent-test")
	if err != nil {
		return err
	}

	// Create URL entity with a random ID to avoid conflicts
	url := entity.NewURL(shortKey, longURL)
	randomID, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	url.ID = randomID.Int64() + 1000000 // Generate a random ID in range 1M-2M
	url.SetExpiration(time.Hour)

	// Try to save, ignore conflicts (for test setup)
	if err := repo.Save(ctx, url); err != nil {
		// Check if it's a duplicate key error, which is acceptable for tests
		if !isDuplicateKeyError(err) {
			return err
		}
	}

	return nil
}

// getEnv gets an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

// isDuplicateKeyError checks if the error is a duplicate key constraint violation.
func isDuplicateKeyError(err error) bool {
	return err != nil &&
		(strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "UNIQUE constraint") ||
			strings.Contains(err.Error(), "already exists"))
}

// BenchmarkIncrementVisitCount benchmarks the performance of concurrent visit counting.
func BenchmarkIncrementVisitCount(b *testing.B) {
	db, err := setupTestDB()
	if err != nil {
		b.Skip("Database not available for benchmark")
	}

	defer func() {
		if err := db.Close(); err != nil {
			b.Logf("Failed to close database: %v", err)
		}
	}()

	repo := postgres.NewURLRepository(db)
	ctx := context.Background()

	shortKey, _ := valueobject.NewShortKey("bench123")
	_ = insertTestURL(ctx, repo, shortKey)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = repo.IncrementVisitCount(ctx, shortKey)
		}
	})
}
