package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Shofyan/url-shortener/internal/domain/entity"
	"github.com/Shofyan/url-shortener/internal/domain/repository"
	"github.com/Shofyan/url-shortener/internal/domain/service"
	"github.com/Shofyan/url-shortener/internal/domain/valueobject"
)

// MockURLRepository for cleanup service testing.
type MockURLRepository struct {
	mock.Mock
}

func (m *MockURLRepository) Save(ctx context.Context, url *entity.URL) error {
	args := m.Called(ctx, url)
	return args.Error(0)
}

func (m *MockURLRepository) FindByShortKey(ctx context.Context, shortKey *valueobject.ShortKey) (*entity.URL, error) {
	args := m.Called(ctx, shortKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*entity.URL), args.Error(1)
}

func (m *MockURLRepository) FindByLongURL(ctx context.Context, longURL *valueobject.LongURL) (*entity.URL, error) {
	args := m.Called(ctx, longURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*entity.URL), args.Error(1)
}

func (m *MockURLRepository) Update(ctx context.Context, url *entity.URL) error {
	args := m.Called(ctx, url)
	return args.Error(0)
}

func (m *MockURLRepository) Delete(ctx context.Context, shortKey *valueobject.ShortKey) error {
	args := m.Called(ctx, shortKey)
	return args.Error(0)
}

func (m *MockURLRepository) ExistsByShortKey(ctx context.Context, shortKey *valueobject.ShortKey) (bool, error) {
	args := m.Called(ctx, shortKey)
	return args.Bool(0), args.Error(1)
}

func (m *MockURLRepository) IncrementVisitCount(ctx context.Context, shortKey *valueobject.ShortKey) error {
	args := m.Called(ctx, shortKey)
	return args.Error(0)
}

func (m *MockURLRepository) FindExpiredURLs(ctx context.Context, before time.Time, maxResults int) ([]*entity.URL, error) {
	args := m.Called(ctx, before, maxResults)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*entity.URL), args.Error(1)
}

func (m *MockURLRepository) DeleteExpiredBatch(ctx context.Context, shortKeys []*valueobject.ShortKey) error {
	args := m.Called(ctx, shortKeys)
	return args.Error(0)
}

func (m *MockURLRepository) GetExpiredCount(ctx context.Context, before time.Time) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

// MockCacheRepository for cleanup service testing.
type MockCacheRepository struct {
	mock.Mock
}

func (m *MockCacheRepository) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheRepository) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheRepository) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockCacheRepository) SetCacheEntry(ctx context.Context, key string, entry *repository.CacheEntry, ttl time.Duration) error {
	args := m.Called(ctx, key, entry, ttl)
	return args.Error(0)
}

func (m *MockCacheRepository) GetCacheEntry(ctx context.Context, key string) (*repository.CacheEntry, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*repository.CacheEntry), args.Error(1)
}

func (m *MockCacheRepository) SetTombstone(ctx context.Context, key string, reason string, ttl time.Duration) error {
	args := m.Called(ctx, key, reason, ttl)
	return args.Error(0)
}

// TestBackgroundURLCleanupService_CleanupExpiredBatch tests the cleanup batch functionality.
func TestBackgroundURLCleanupService_CleanupExpiredBatch(t *testing.T) {
	tests := []struct {
		name            string
		setupMocks      func(*MockURLRepository, *MockCacheRepository)
		batchSize       int
		expectedCleaned int
		expectError     bool
	}{
		{
			name:            "No expired URLs - should return 0",
			batchSize:       100,
			expectedCleaned: 0,
			expectError:     false,
			setupMocks: func(urlRepo *MockURLRepository, _ *MockCacheRepository) {
				urlRepo.On("FindExpiredURLs", mock.Anything, mock.AnythingOfType("time.Time"), 100).
					Return([]*entity.URL{}, nil)
			},
		},
		{
			name:            "Multiple expired URLs - should clean all",
			batchSize:       100,
			expectedCleaned: 3,
			expectError:     false,
			setupMocks: func(urlRepo *MockURLRepository, cacheRepo *MockCacheRepository) {
				// Create expired URLs
				shortKey1, _ := valueobject.NewShortKey("expired1")
				shortKey2, _ := valueobject.NewShortKey("expired2")
				shortKey3, _ := valueobject.NewShortKey("expired3")
				longURL, _ := valueobject.NewLongURL("https://example.com")

				url1 := entity.NewURL(shortKey1, longURL)
				url2 := entity.NewURL(shortKey2, longURL)
				url3 := entity.NewURL(shortKey3, longURL)

				expiredURLs := []*entity.URL{url1, url2, url3}

				urlRepo.On("FindExpiredURLs", mock.Anything, mock.AnythingOfType("time.Time"), 100).
					Return(expiredURLs, nil)

				// Expect batch delete with the short keys
				urlRepo.On("DeleteExpiredBatch", mock.Anything, mock.AnythingOfType("[]*valueobject.ShortKey")).
					Return(nil)

				// Expect cache cleanup (best effort)
				cacheRepo.On("Delete", mock.Anything, "expired1").Return(nil)
				cacheRepo.On("Delete", mock.Anything, "expired2").Return(nil)
				cacheRepo.On("Delete", mock.Anything, "expired3").Return(nil)
			},
		},
		{
			name:            "Database error during batch delete - should return error",
			batchSize:       100,
			expectedCleaned: 0,
			expectError:     true,
			setupMocks: func(urlRepo *MockURLRepository, _ *MockCacheRepository) {
				shortKey1, _ := valueobject.NewShortKey("expired1")
				longURL, _ := valueobject.NewLongURL("https://example.com")
				url1 := entity.NewURL(shortKey1, longURL)
				expiredURLs := []*entity.URL{url1}

				urlRepo.On("FindExpiredURLs", mock.Anything, mock.AnythingOfType("time.Time"), 100).
					Return(expiredURLs, nil)

				// Simulate database error during batch delete
				urlRepo.On("DeleteExpiredBatch", mock.Anything, mock.AnythingOfType("[]*valueobject.ShortKey")).
					Return(assert.AnError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			urlRepo := &MockURLRepository{}
			cacheRepo := &MockCacheRepository{}

			tt.setupMocks(urlRepo, cacheRepo)

			// Create cleanup service with test config
			config := &service.CleanupConfig{
				Enabled:            true,
				CleanupInterval:    1 * time.Minute,
				BatchSize:          100,
				BufferTime:         1 * time.Hour,
				MaxCleanupDuration: 5 * time.Minute,
			}

			cleanupService := service.NewBackgroundURLCleanupService(urlRepo, cacheRepo, config)

			// Execute test
			ctx := context.Background()
			cleaned, err := cleanupService.CleanupExpiredBatch(ctx, tt.batchSize)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedCleaned, cleaned)

			// Verify mock expectations
			urlRepo.AssertExpectations(t)
			cacheRepo.AssertExpectations(t)
		})
	}
}

// TestBackgroundURLCleanupService_BufferTime tests that cleanup respects buffer time.
func TestBackgroundURLCleanupService_BufferTime(t *testing.T) {
	urlRepo := &MockURLRepository{}
	cacheRepo := &MockCacheRepository{}

	// Create service with 2-hour buffer time
	config := &service.CleanupConfig{
		Enabled:            true,
		CleanupInterval:    1 * time.Minute,
		BatchSize:          100,
		BufferTime:         2 * time.Hour, // 2-hour buffer
		MaxCleanupDuration: 5 * time.Minute,
	}

	cleanupService := service.NewBackgroundURLCleanupService(urlRepo, cacheRepo, config)

	// Mock should be called with time that includes the buffer
	urlRepo.On("FindExpiredURLs", mock.Anything,
		mock.MatchedBy(func(cutoff time.Time) bool {
			// Cutoff should be approximately 2 hours ago (with some tolerance for test execution time)
			expectedCutoff := time.Now().Add(-2 * time.Hour)
			diff := expectedCutoff.Sub(cutoff).Abs()
			return diff < 10*time.Second // 10-second tolerance
		}),
		100).Return([]*entity.URL{}, nil)

	// Execute test
	ctx := context.Background()
	cleaned, err := cleanupService.CleanupExpiredBatch(ctx, 100)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, 0, cleaned)
	urlRepo.AssertExpectations(t)
}

// TestBackgroundURLCleanupService_Stats tests statistics tracking.
func TestBackgroundURLCleanupService_Stats(t *testing.T) {
	urlRepo := &MockURLRepository{}
	cacheRepo := &MockCacheRepository{}

	config := service.DefaultCleanupConfig()
	cleanupService := service.NewBackgroundURLCleanupService(urlRepo, cacheRepo, config)

	// Initial stats should be empty
	stats := cleanupService.GetCleanupStats()
	assert.Equal(t, int64(0), stats.TotalCleaned)
	assert.Equal(t, int64(0), stats.SuccessfulRuns)
	assert.Equal(t, int64(0), stats.FailedRuns)
	assert.False(t, stats.IsRunning)

	// Setup mock for successful cleanup
	shortKey1, _ := valueobject.NewShortKey("expired1")
	longURL, _ := valueobject.NewLongURL("https://example.com")
	url1 := entity.NewURL(shortKey1, longURL)
	expiredURLs := []*entity.URL{url1}

	urlRepo.On("FindExpiredURLs", mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("int")).
		Return(expiredURLs, nil)
	urlRepo.On("DeleteExpiredBatch", mock.Anything, mock.AnythingOfType("[]*valueobject.ShortKey")).
		Return(nil)
	cacheRepo.On("Delete", mock.Anything, "expired1").Return(nil)

	// Execute cleanup
	ctx := context.Background()
	cleaned, err := cleanupService.CleanupExpiredBatch(ctx, 100)

	// Verify results
	require.NoError(t, err)
	assert.Equal(t, 1, cleaned)

	// Verify stats updated
	stats = cleanupService.GetCleanupStats()
	assert.Equal(t, int64(1), stats.TotalCleaned)
	assert.Equal(t, int64(1), stats.SuccessfulRuns)
	assert.Equal(t, int64(0), stats.FailedRuns)
	assert.Equal(t, 1, stats.LastBatchSize)
	assert.GreaterOrEqual(t, stats.AverageCleanupMs, 0.0)
}

// TestBackgroundURLCleanupService_DisabledConfig tests disabled cleanup service.
func TestBackgroundURLCleanupService_DisabledConfig(t *testing.T) {
	urlRepo := &MockURLRepository{}
	cacheRepo := &MockCacheRepository{}

	// Create disabled config
	config := &service.CleanupConfig{
		Enabled: false,
	}

	cleanupService := service.NewBackgroundURLCleanupService(urlRepo, cacheRepo, config)

	// Start cleanup should not error but also not start anything
	ctx := context.Background()
	err := cleanupService.StartCleanup(ctx)
	assert.NoError(t, err)

	// Stats should show not running
	stats := cleanupService.GetCleanupStats()
	assert.False(t, stats.IsRunning)

	// No repository methods should be called
	urlRepo.AssertNotCalled(t, "FindExpiredURLs", mock.Anything, mock.Anything, mock.Anything)
}
