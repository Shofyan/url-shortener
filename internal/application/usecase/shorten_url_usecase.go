package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Shofyan/url-shortener/internal/application/dto"
	"github.com/Shofyan/url-shortener/internal/domain/entity"
	"github.com/Shofyan/url-shortener/internal/domain/repository"
	"github.com/Shofyan/url-shortener/internal/domain/service"
	"github.com/Shofyan/url-shortener/internal/domain/valueobject"
)

var (
	// ErrURLNotFound is returned when the requested URL is not found in the system.
	ErrURLNotFound = errors.New("URL not found")
	// ErrURLExpired is returned when the URL has exceeded its expiration time.
	ErrURLExpired      = errors.New("URL has expired")
	ErrCustomKeyExists = errors.New("custom short key already exists")
	ErrInternalError   = errors.New("internal server error")
)

// ShortenURLUseCase handles URL shortening business logic.
type ShortenURLUseCase struct {
	urlRepo    repository.URLRepository
	cacheRepo  repository.CacheRepository
	genService *service.GeneratorService
	baseURL    string
	defaultTTL time.Duration
}

// NewShortenURLUseCase creates a new ShortenURLUseCase.
func NewShortenURLUseCase(
	urlRepo repository.URLRepository,
	cacheRepo repository.CacheRepository,
	genService *service.GeneratorService,
	baseURL string,
	defaultTTL time.Duration,
) *ShortenURLUseCase {
	return &ShortenURLUseCase{
		urlRepo:    urlRepo,
		cacheRepo:  cacheRepo,
		genService: genService,
		baseURL:    baseURL,
		defaultTTL: defaultTTL,
	}
}

// Shorten creates a short URL from a long URL.
func (uc *ShortenURLUseCase) Shorten(ctx context.Context, req *dto.ShortenURLRequest) (*dto.ShortenURLResponse, error) {
	log.Printf("[Shorten] Starting URL shortening process for: %s", req.LongURL)

	longURL, err := uc.validateAndNormalizeLongURL(req.LongURL)
	if err != nil {
		return nil, err
	}

	// Check if URL already exists (only if no custom key is provided)
	if req.CustomKey == "" {
		if existingURL := uc.findExistingURL(ctx, longURL); existingURL != nil {
			return uc.buildResponse(existingURL), nil
		}
	}

	shortKey, id, err := uc.generateShortKey(ctx, req.CustomKey)
	if err != nil {
		return nil, err
	}

	url := uc.createAndConfigureURL(shortKey, longURL, id, int(req.TTLSeconds))

	if err := uc.urlRepo.Save(ctx, url); err != nil {
		log.Printf("[Shorten] Error saving URL to database: %v", err)
		return nil, fmt.Errorf("failed to save URL: %w", err)
	}

	log.Printf("[Shorten] URL saved successfully to database")

	uc.cacheURL(ctx, shortKey, longURL, url.ExpiresAt)

	log.Printf("[Shorten] URL shortening completed successfully. Short URL: %s/%s", uc.baseURL, shortKey.Value())

	return uc.buildResponse(url), nil
}

// validateAndNormalizeLongURL validates and normalizes the long URL.
func (uc *ShortenURLUseCase) validateAndNormalizeLongURL(rawURL string) (*valueobject.LongURL, error) {
	normalizedURL := valueobject.NormalizeURL(rawURL)
	log.Printf("[Shorten] Normalized URL: %s", normalizedURL)

	longURL, err := valueobject.NewLongURL(normalizedURL)
	if err != nil {
		log.Printf("[Shorten] Error creating long URL value object: %v", err)
		return nil, err
	}

	log.Printf("[Shorten] Long URL validation successful")

	return longURL, nil
}

// findExistingURL checks for existing non-expired URLs.
func (uc *ShortenURLUseCase) findExistingURL(ctx context.Context, longURL *valueobject.LongURL) *entity.URL {
	log.Printf("[Shorten] Checking if URL already exists in database")

	existingURL, err := uc.urlRepo.FindByLongURL(ctx, longURL)
	if err == nil && existingURL != nil && !existingURL.IsExpired() {
		log.Printf("[Shorten] Found existing URL with short key: %s", existingURL.ShortKey.Value())
		return existingURL
	}

	log.Printf("[Shorten] No existing URL found, proceeding with new short URL creation")

	return nil
}

// generateShortKey generates or validates a custom short key.
func (uc *ShortenURLUseCase) generateShortKey(ctx context.Context, customKey string) (*valueobject.ShortKey, int64, error) {
	if customKey != "" {
		return uc.processCustomKey(ctx, customKey)
	}

	return uc.generateNewKey()
}

// processCustomKey validates and processes a custom key.
func (uc *ShortenURLUseCase) processCustomKey(ctx context.Context, customKey string) (*valueobject.ShortKey, int64, error) {
	log.Printf("[Shorten] Using custom key: %s", customKey)

	shortKey, err := valueobject.NewShortKey(customKey)
	if err != nil {
		log.Printf("[Shorten] Error creating custom short key: %v", err)
		return nil, 0, err
	}

	log.Printf("[Shorten] Custom short key validation successful")

	exists, err := uc.urlRepo.ExistsByShortKey(ctx, shortKey)
	if err != nil {
		log.Printf("[Shorten] Error checking custom key existence: %v", err)
		return nil, 0, ErrInternalError
	}

	if exists {
		log.Printf("[Shorten] Custom key already exists: %s", customKey)
		return nil, 0, ErrCustomKeyExists
	}

	log.Printf("[Shorten] Custom key is available")

	id, err := uc.genService.GenerateID()
	if err != nil {
		log.Printf("[Shorten] Error generating ID: %v", err)
		return nil, 0, ErrInternalError
	}

	log.Printf("[Shorten] Generated ID: %d", id)

	return shortKey, id, nil
}

// generateNewKey generates a new short key and ID.
func (uc *ShortenURLUseCase) generateNewKey() (*valueobject.ShortKey, int64, error) {
	log.Printf("[Shorten] Generating short key using generator service")

	shortKey, id, err := uc.genService.GenerateShortKey()
	if err != nil {
		log.Printf("[Shorten] Error generating short key: %v", err)
		return nil, 0, ErrInternalError
	}

	log.Printf("[Shorten] Generated short key: %s, ID: %d", shortKey.Value(), id)

	return shortKey, id, nil
}

// createAndConfigureURL creates a URL entity and sets its expiration.
func (uc *ShortenURLUseCase) createAndConfigureURL(shortKey *valueobject.ShortKey, longURL *valueobject.LongURL, id int64, ttlSeconds int) *entity.URL {
	log.Printf("[Shorten] Creating URL entity")

	url := entity.NewURL(shortKey, longURL)
	url.ID = id
	log.Printf("[Shorten] URL entity created with ID: %d, ShortKey: %s", url.ID, url.ShortKey.Value())

	const defaultTTLSeconds = 24 * 60 * 60 // 24 hours

	ttl := ttlSeconds
	if ttl == 0 {
		ttl = defaultTTLSeconds
		log.Printf("[Shorten] No TTL specified, using default: %d seconds (24 hours)", defaultTTLSeconds)
	}

	log.Printf("[Shorten] Setting expiration: %d seconds", ttl)
	url.SetExpiration(time.Duration(ttl) * time.Second)
	log.Printf("[Shorten] Expiration set to: %v", url.ExpiresAt)

	return url
}

// cacheURL caches the URL mapping.
func (uc *ShortenURLUseCase) cacheURL(ctx context.Context, shortKey *valueobject.ShortKey, longURL *valueobject.LongURL, expiresAt *time.Time) {
	cacheTTL := uc.defaultTTL
	if expiresAt != nil {
		cacheTTL = time.Until(*expiresAt)
	}

	log.Printf("[Shorten] Caching URL mapping with TTL: %v", cacheTTL)

	if err := uc.cacheRepo.Set(ctx, shortKey.Value(), longURL.Value(), cacheTTL); err != nil {
		log.Printf("[Shorten] Warning: Failed to cache URL mapping: %v", err)
	} else {
		log.Printf("[Shorten] URL cached successfully")
	}
}

// GetLongURL retrieves the long URL from a short key.
func (uc *ShortenURLUseCase) GetLongURL(ctx context.Context, shortKeyStr string) (string, error) {
	shortKey, err := valueobject.NewShortKey(shortKeyStr)
	if err != nil {
		return "", err
	}

	var longURL string

	// Try cache first
	if cachedURL, err := uc.cacheRepo.Get(ctx, shortKey.Value()); err == nil && cachedURL != "" {
		longURL = cachedURL
	} else {
		// Fetch from database
		url, err := uc.urlRepo.FindByShortKey(ctx, shortKey)
		if err != nil {
			return "", ErrURLNotFound
		}

		// Check expiration
		if url.IsExpired() {
			_ = uc.urlRepo.Delete(ctx, shortKey)
			_ = uc.cacheRepo.Delete(ctx, shortKey.Value())

			return "", ErrURLExpired
		}

		longURL = url.LongURL.Value()

		// Update cache
		cacheTTL := uc.defaultTTL
		if url.ExpiresAt != nil {
			cacheTTL = time.Until(*url.ExpiresAt)
		}

		_ = uc.cacheRepo.Set(ctx, shortKey.Value(), longURL, cacheTTL)
	}

	// Always increment visit count regardless of cache hit/miss
	if err := uc.urlRepo.IncrementVisitCount(ctx, shortKey); err != nil {
		log.Printf("Warning: Failed to increment visit count for %s: %v", shortKey.Value(), err)
	}

	return longURL, nil
}

// GetStats retrieves statistics for a short URL.
func (uc *ShortenURLUseCase) GetStats(ctx context.Context, shortKeyStr string) (*dto.URLStatsResponse, error) {
	shortKey, err := valueobject.NewShortKey(shortKeyStr)
	if err != nil {
		return nil, err
	}

	url, err := uc.urlRepo.FindByShortKey(ctx, shortKey)
	if err != nil {
		return nil, ErrURLNotFound
	}

	if url.IsExpired() {
		return nil, ErrURLExpired
	}

	resp := &dto.URLStatsResponse{
		ShortKey:   url.ShortKey.Value(),
		LongURL:    url.LongURL.Value(),
		VisitCount: url.VisitCount,
		CreatedAt:  url.CreatedAt.Format(time.RFC3339),
	}

	if url.ExpiresAt != nil {
		resp.ExpiresAt = url.ExpiresAt.Format(time.RFC3339)
	}

	if url.LastAccessedAt != nil {
		resp.LastAccessedAt = url.LastAccessedAt.Format(time.RFC3339)
	}

	return resp, nil
}

// buildResponse builds a ShortenURLResponse from a URL entity.
func (uc *ShortenURLUseCase) buildResponse(url *entity.URL) *dto.ShortenURLResponse {
	resp := &dto.ShortenURLResponse{
		ShortURL:  fmt.Sprintf("%s/%s", uc.baseURL, url.ShortKey.Value()),
		ShortKey:  url.ShortKey.Value(),
		LongURL:   url.LongURL.Value(),
		CreatedAt: url.CreatedAt.Format(time.RFC3339),
	}

	if url.ExpiresAt != nil {
		resp.ExpiresAt = url.ExpiresAt.Format(time.RFC3339)
	}

	return resp
}
