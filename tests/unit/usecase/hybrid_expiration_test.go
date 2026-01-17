package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Shofyan/url-shortener/internal/application/usecase"
	"github.com/Shofyan/url-shortener/internal/domain/entity"
	"github.com/Shofyan/url-shortener/internal/domain/repository"
	"github.com/Shofyan/url-shortener/internal/domain/valueobject"
)

// TestHybridExpirationStrategy tests the hybrid expiration strategy implementation.
func TestHybridExpirationStrategy(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*MockURLRepository, *MockCacheRepository)
		shortKey    string
		expectError error
		expectCalls func(*testing.T, *MockURLRepository, *MockCacheRepository)
	}{
		{
			name:     "Cache hit with valid URL - should not hit database",
			shortKey: "abc123",
			setupMocks: func(urlRepo *MockURLRepository, cacheRepo *MockCacheRepository) {
				validEntry := &repository.CacheEntry{
					LongURL:     "https://example.com",
					ExpiresAt:   nil, // Never expires
					CreatedAt:   time.Now(),
					IsTombstone: false,
				}

				cacheRepo.On("GetCacheEntry", mock.Anything, "abc123").Return(validEntry, nil)
				urlRepo.On("IncrementVisitCount", mock.Anything, mock.AnythingOfType("*valueobject.ShortKey")).Return(nil)
			},
			expectError: nil,
			expectCalls: func(t *testing.T, urlRepo *MockURLRepository, cacheRepo *MockCacheRepository) {
				// Should not call FindByShortKey since cache hit
				urlRepo.AssertNotCalled(t, "FindByShortKey", mock.Anything, mock.Anything)

				// Should call GetCacheEntry and IncrementVisitCount
				cacheRepo.AssertCalled(t, "GetCacheEntry", mock.Anything, "abc123")
				urlRepo.AssertCalled(t, "IncrementVisitCount", mock.Anything, mock.AnythingOfType("*valueobject.ShortKey"))
			},
		},
		{
			name:     "Cache miss with database expired URL - lazy validation without sync delete",
			shortKey: "expired123",
			setupMocks: func(urlRepo *MockURLRepository, cacheRepo *MockCacheRepository) {
				// Cache miss
				cacheRepo.On("GetCacheEntry", mock.Anything, "expired123").Return(nil, assert.AnError)

				// Database returns expired URL
				shortKey, _ := valueobject.NewShortKey("expired123")
				longURL, _ := valueobject.NewLongURL("https://example.com")
				expiredURL := entity.NewURL(shortKey, longURL)
				expiredTime := time.Now().Add(-1 * time.Hour)
				expiredURL.ExpiresAt = &expiredTime

				urlRepo.On("FindByShortKey", mock.Anything, mock.AnythingOfType("*valueobject.ShortKey")).Return(expiredURL, nil)
				cacheRepo.On("SetTombstone", mock.Anything, "expired123", "expired", time.Hour).Return(nil)
			},
			expectError: usecase.ErrURLExpired,
			expectCalls: func(t *testing.T, urlRepo *MockURLRepository, cacheRepo *MockCacheRepository) {
				// Should call database to fetch URL
				urlRepo.AssertCalled(t, "FindByShortKey", mock.Anything, mock.AnythingOfType("*valueobject.ShortKey"))

				// CRITICAL: Should NOT delete from database synchronously
				urlRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)

				// Should cache tombstone to prevent further DB hits
				cacheRepo.AssertCalled(t, "SetTombstone", mock.Anything, "expired123", "expired", time.Hour)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			urlRepo := &MockURLRepository{}
			cacheRepo := &MockCacheRepository{}

			tt.setupMocks(urlRepo, cacheRepo)

			// Create use case
			useCase := usecase.NewShortenURLUseCase(
				urlRepo,
				cacheRepo,
				nil, // Generator service not needed for GetLongURL tests
				"http://localhost:8080",
				24*time.Hour,
			)

			// Execute test
			ctx := context.Background()
			result, err := useCase.GetLongURL(ctx, tt.shortKey)

			// Verify results
			if tt.expectError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectError, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
			}

			// Verify mock calls
			if tt.expectCalls != nil {
				tt.expectCalls(t, urlRepo, cacheRepo)
			}

			// Verify all expectations met
			urlRepo.AssertExpectations(t)
			cacheRepo.AssertExpectations(t)
		})
	}
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
