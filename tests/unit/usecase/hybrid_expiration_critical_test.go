// Package usecase contains unit tests for URL shortener use cases,
// including critical tests for hybrid expiration strategy.
package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Shofyan/url-shortener/internal/application/usecase"
	"github.com/Shofyan/url-shortener/internal/domain/entity"
	"github.com/Shofyan/url-shortener/internal/domain/valueobject"
)

// TestHybridExpirationStrategy_NoSynchronousDeletes is the CRITICAL test to ensure
// the read path never performs synchronous deletes - this is the core principle of the hybrid strategy.
func TestHybridExpirationStrategy_NoSynchronousDeletes(t *testing.T) {
	urlRepo := &MockURLRepository{}
	cacheRepo := &MockCacheRepository{}

	// Cache miss
	cacheRepo.On("GetCacheEntry", mock.Anything, "expired").Return(nil, assert.AnError)

	// Database returns expired URL
	shortKey, _ := valueobject.NewShortKey("expired")
	longURL, _ := valueobject.NewLongURL("https://example.com")
	expiredURL := entity.NewURL(shortKey, longURL)
	expiredTime := time.Now().Add(-1 * time.Hour)
	expiredURL.ExpiresAt = &expiredTime

	urlRepo.On("FindByShortKey", mock.Anything, mock.AnythingOfType("*valueobject.ShortKey")).Return(expiredURL, nil)
	cacheRepo.On("SetTombstone", mock.Anything, "expired", "expired", time.Hour).Return(nil)

	useCase := usecase.NewShortenURLUseCase(
		urlRepo,
		cacheRepo,
		nil, // Generator service not needed for GetLongURL tests
		"http://localhost:8080",
		24*time.Hour,
	)

	// Execute
	ctx := context.Background()
	_, err := useCase.GetLongURL(ctx, "expired")

	// Verify error
	require.Equal(t, usecase.ErrURLExpired, err)

	// CRITICAL: Verify no synchronous delete was called
	urlRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
	cacheRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)

	// Verify tombstone was cached
	cacheRepo.AssertCalled(t, "SetTombstone", mock.Anything, "expired", "expired", time.Hour)

	urlRepo.AssertExpectations(t)
	cacheRepo.AssertExpectations(t)
}
