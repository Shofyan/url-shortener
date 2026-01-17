package repository

import (
	"context"
	"time"

	"github.com/Shofyan/url-shortener/internal/domain/entity"
	"github.com/Shofyan/url-shortener/internal/domain/valueobject"
)

// URLRepository defines the interface for URL persistence.
type URLRepository interface {
	// Save saves a new URL mapping
	Save(ctx context.Context, url *entity.URL) error

	// FindByShortKey retrieves a URL by its short key
	FindByShortKey(ctx context.Context, shortKey *valueobject.ShortKey) (*entity.URL, error)

	// FindByLongURL retrieves a URL by its long URL
	FindByLongURL(ctx context.Context, longURL *valueobject.LongURL) (*entity.URL, error)

	// Update updates an existing URL
	Update(ctx context.Context, url *entity.URL) error

	// Delete deletes a URL by its short key
	Delete(ctx context.Context, shortKey *valueobject.ShortKey) error

	// ExistsByShortKey checks if a short key already exists
	ExistsByShortKey(ctx context.Context, shortKey *valueobject.ShortKey) (bool, error)

	// IncrementVisitCount atomically increments visit count and updates last_accessed_at
	IncrementVisitCount(ctx context.Context, shortKey *valueobject.ShortKey) error

	// FindExpiredURLs returns URLs that expired before the given timestamp
	// Limited to maxResults for batch processing
	FindExpiredURLs(ctx context.Context, before time.Time, maxResults int) ([]*entity.URL, error)

	// DeleteExpiredBatch deletes multiple URLs by their short keys in a single transaction
	DeleteExpiredBatch(ctx context.Context, shortKeys []*valueobject.ShortKey) error

	// GetExpiredCount returns the total count of expired URLs for monitoring
	GetExpiredCount(ctx context.Context, before time.Time) (int64, error)
}
