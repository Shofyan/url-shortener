package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Shofyan/url-shortener/internal/application/dto"
	"github.com/Shofyan/url-shortener/internal/application/usecase"
	"github.com/Shofyan/url-shortener/internal/domain/entity"
	"github.com/Shofyan/url-shortener/internal/domain/repository"
	"github.com/Shofyan/url-shortener/internal/domain/service"
	"github.com/Shofyan/url-shortener/internal/domain/valueobject"
)

// MockURLRepository is a mock implementation of URLRepository.
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

// MockCacheRepository is a mock implementation of CacheRepository.
type MockCacheRepository struct {
	mock.Mock
}

func (m *MockCacheRepository) Set(ctx context.Context, key, value string, ttl time.Duration) error {
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

// MockGeneratorService is a mock implementation of GeneratorService.
type MockIDGenerator struct {
	mock.Mock
}

func (m *MockIDGenerator) Generate() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

type MockShortKeyGenerator struct {
	mock.Mock
}

func (m *MockShortKeyGenerator) GenerateFromID(id int64) (*valueobject.ShortKey, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*valueobject.ShortKey), args.Error(1)
}

func (m *MockShortKeyGenerator) DecodeToID(shortKey *valueobject.ShortKey) (int64, error) {
	args := m.Called(shortKey)
	return args.Get(0).(int64), args.Error(1)
}

func TestShortenURL_Success(t *testing.T) {
	// Setup
	mockURLRepo := new(MockURLRepository)
	mockCacheRepo := new(MockCacheRepository)
	mockIDGen := new(MockIDGenerator)
	mockShortKeyGen := new(MockShortKeyGenerator)
	genService := service.NewGeneratorService(mockIDGen, mockShortKeyGen)

	uc := usecase.NewShortenURLUseCase(
		mockURLRepo,
		mockCacheRepo,
		genService,
		"http://localhost:8080",
		time.Hour,
	)

	// Test data
	req := &dto.ShortenURLRequest{
		LongURL:    "https://example.com",
		TTLSeconds: 3600, // 1 hour
	}

	expectedID := int64(12345)
	shortKey, _ := valueobject.NewShortKey("abc123")
	longURL, _ := valueobject.NewLongURL("https://example.com")

	// Mock expectations
	mockURLRepo.On("FindByLongURL", mock.Anything, longURL).Return(nil, usecase.ErrURLNotFound)
	mockIDGen.On("Generate").Return(expectedID, nil)
	mockShortKeyGen.On("GenerateFromID", expectedID).Return(shortKey, nil)
	mockURLRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.URL")).Return(nil)
	mockCacheRepo.On("SetCacheEntry", mock.Anything, shortKey.Value(), mock.AnythingOfType("*repository.CacheEntry"), mock.AnythingOfType("time.Duration")).Return(nil)

	// Execute
	resp, err := uc.Shorten(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "abc123", resp.ShortKey)
	assert.Equal(t, "https://example.com", resp.LongURL)
	assert.Equal(t, "http://localhost:8080/abc123", resp.ShortURL)

	// Verify mocks
	mockURLRepo.AssertExpectations(t)
	mockCacheRepo.AssertExpectations(t)
	mockIDGen.AssertExpectations(t)
	mockShortKeyGen.AssertExpectations(t)
}

func TestShortenURL_DefaultTTL(t *testing.T) {
	// Setup
	mockURLRepo := new(MockURLRepository)
	mockCacheRepo := new(MockCacheRepository)
	mockIDGen := new(MockIDGenerator)
	mockShortKeyGen := new(MockShortKeyGenerator)
	genService := service.NewGeneratorService(mockIDGen, mockShortKeyGen)

	uc := usecase.NewShortenURLUseCase(
		mockURLRepo,
		mockCacheRepo,
		genService,
		"http://localhost:8080",
		time.Hour,
	)

	// Test data - no TTL specified
	req := &dto.ShortenURLRequest{
		LongURL: "https://example.com",
		// TTLSeconds: 0, // Should use default 24 hours
	}

	expectedID := int64(12345)
	shortKey, _ := valueobject.NewShortKey("abc123")
	longURL, _ := valueobject.NewLongURL("https://example.com")

	// Mock expectations
	mockURLRepo.On("FindByLongURL", mock.Anything, longURL).Return(nil, usecase.ErrURLNotFound)
	mockIDGen.On("Generate").Return(expectedID, nil)
	mockShortKeyGen.On("GenerateFromID", expectedID).Return(shortKey, nil)
	mockURLRepo.On("Save", mock.Anything, mock.MatchedBy(func(url *entity.URL) bool {
		// Verify that expiration is set to 24 hours from now
		if url.ExpiresAt == nil {
			return false
		}
		expectedExpiration := time.Now().Add(24 * time.Hour)
		timeDiff := url.ExpiresAt.Sub(expectedExpiration).Abs()
		return timeDiff < time.Minute // Allow 1 minute tolerance
	})).Return(nil)
	mockCacheRepo.On("SetCacheEntry", mock.Anything, shortKey.Value(), mock.AnythingOfType("*repository.CacheEntry"), mock.AnythingOfType("time.Duration")).Return(nil)

	// Execute
	resp, err := uc.Shorten(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.ExpiresAt)

	// Verify mocks
	mockURLRepo.AssertExpectations(t)
	mockCacheRepo.AssertExpectations(t)
	mockIDGen.AssertExpectations(t)
	mockShortKeyGen.AssertExpectations(t)
}

func TestGetLongURL_Success(t *testing.T) {
	// Setup
	mockURLRepo := new(MockURLRepository)
	mockCacheRepo := new(MockCacheRepository)
	mockIDGen := new(MockIDGenerator)
	mockShortKeyGen := new(MockShortKeyGenerator)
	genService := service.NewGeneratorService(mockIDGen, mockShortKeyGen)

	uc := usecase.NewShortenURLUseCase(
		mockURLRepo,
		mockCacheRepo,
		genService,
		"http://localhost:8080",
		time.Hour,
	)

	// Test data
	shortKey, _ := valueobject.NewShortKey("abc123")
	longURL, _ := valueobject.NewLongURL("https://example.com")

	url := &entity.URL{
		ID:         12345,
		ShortKey:   shortKey,
		LongURL:    longURL,
		CreatedAt:  time.Now(),
		VisitCount: 5,
	}
	expiresAt := time.Now().Add(time.Hour)
	url.ExpiresAt = &expiresAt

	// Mock expectations - cache miss, then database hit
	mockCacheRepo.On("GetCacheEntry", mock.Anything, "abc123").Return(nil, assert.AnError)
	mockURLRepo.On("FindByShortKey", mock.Anything, shortKey).Return(url, nil)
	mockURLRepo.On("IncrementVisitCount", mock.Anything, shortKey).Return(nil)
	mockCacheRepo.On("SetCacheEntry", mock.Anything, "abc123", mock.AnythingOfType("*repository.CacheEntry"), mock.AnythingOfType("time.Duration")).Return(nil)

	// Execute
	redirectURL, err := uc.GetLongURL(context.Background(), "abc123")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com", redirectURL)

	// Verify mocks
	mockURLRepo.AssertExpectations(t)
	mockCacheRepo.AssertExpectations(t)
}

func TestGetStats_Success(t *testing.T) {
	// Setup
	mockURLRepo := new(MockURLRepository)
	mockCacheRepo := new(MockCacheRepository)
	mockIDGen := new(MockIDGenerator)
	mockShortKeyGen := new(MockShortKeyGenerator)
	genService := service.NewGeneratorService(mockIDGen, mockShortKeyGen)

	uc := usecase.NewShortenURLUseCase(
		mockURLRepo,
		mockCacheRepo,
		genService,
		"http://localhost:8080",
		time.Hour,
	)

	// Test data
	shortKey, _ := valueobject.NewShortKey("abc123")
	longURL, _ := valueobject.NewLongURL("https://example.com")

	createdAt := time.Now().Add(-time.Hour)
	expiresAt := time.Now().Add(time.Hour)
	lastAccessedAt := time.Now().Add(-time.Minute)

	url := &entity.URL{
		ID:             12345,
		ShortKey:       shortKey,
		LongURL:        longURL,
		CreatedAt:      createdAt,
		ExpiresAt:      &expiresAt,
		VisitCount:     10,
		LastAccessedAt: &lastAccessedAt,
	}

	// Mock expectations
	mockURLRepo.On("FindByShortKey", mock.Anything, shortKey).Return(url, nil)

	// Execute
	resp, err := uc.GetStats(context.Background(), "abc123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "abc123", resp.ShortKey)
	assert.Equal(t, "https://example.com", resp.LongURL)
	assert.Equal(t, int64(10), resp.VisitCount)
	assert.Equal(t, createdAt.Format(time.RFC3339), resp.CreatedAt)
	assert.Equal(t, expiresAt.Format(time.RFC3339), resp.ExpiresAt)
	assert.Equal(t, lastAccessedAt.Format(time.RFC3339), resp.LastAccessedAt)

	// Verify mocks
	mockURLRepo.AssertExpectations(t)
}
