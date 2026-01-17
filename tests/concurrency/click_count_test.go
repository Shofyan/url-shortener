package concurrency_test

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
		t.Skip("Database not available for concurrency test")
	}
	defer db.Close()

	repo := postgres.NewURLRepository(db)
	ctx := context.Background()

	// Setup test data
	shortKey, err := valueobject.NewShortKey("test123")
	require.NoError(t, err)

	// Insert initial URL for testing
	err = insertTestURL(db, shortKey)
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

	// Verify last_accessed_at was updated
	assert.NotNil(t, finalURL.LastAccessedAt, "LastAccessedAt should be set after increment")
	assert.True(t, finalURL.LastAccessedAt.After(initialURL.CreatedAt),
		"LastAccessedAt should be after creation time")

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
	defer db.Close()

	repo := postgres.NewURLRepository(db)
	ctx := context.Background()

	shortKey, err := valueobject.NewShortKey("extreme123")
	require.NoError(t, err)

	err = insertTestURL(db, shortKey)
	require.NoError(t, err)

	// Test with even more concurrent requests
	concurrentRequests := 500
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

	// Verify that most requests succeeded
	successRate := float64(successCount) / float64(concurrentRequests) * 100
	assert.Greater(t, successRate, 95.0,
		"Success rate should be above 95%% under extreme load")

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
	dsn := "postgres://test:test@localhost:5432/urlshortener_test?sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// insertTestURL inserts a test URL into the database.
func insertTestURL(db *sql.DB, shortKey *valueobject.ShortKey) error {
	query := `
		INSERT INTO urls (id, short_key, long_url, created_at, expires_at, visit_count, last_accessed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (short_key) DO NOTHING
	`

	now := time.Now()
	expiresAt := now.Add(time.Hour)

	_, err := db.Exec(query,
		12345,
		shortKey.Value(),
		"https://example.com/concurrent-test",
		now,
		expiresAt,
		0,
		nil,
	)

	return err
}

// BenchmarkIncrementVisitCount benchmarks the performance of concurrent visit counting.
func BenchmarkIncrementVisitCount(b *testing.B) {
	db, err := setupTestDB()
	if err != nil {
		b.Skip("Database not available for benchmark")
	}
	defer db.Close()

	repo := postgres.NewURLRepository(db)
	ctx := context.Background()

	shortKey, _ := valueobject.NewShortKey("bench123")
	_ = insertTestURL(db, shortKey)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = repo.IncrementVisitCount(ctx, shortKey)
		}
	})
}
