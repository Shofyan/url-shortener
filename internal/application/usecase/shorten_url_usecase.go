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
	// ErrURLNotFound is returned when the requested URL is not found in the system
	ErrURLNotFound     = errors.New("URL not found")
	ErrURLExpired      = errors.New("URL has expired")
	ErrCustomKeyExists = errors.New("custom short key already exists")
	ErrInternalError   = errors.New("internal server error")
)

// ShortenURLUseCase handles URL shortening business logic
type ShortenURLUseCase struct {
	urlRepo    repository.URLRepository
	cacheRepo  repository.CacheRepository
	genService *service.GeneratorService
	baseURL    string
	defaultTTL time.Duration
}

// NewShortenURLUseCase creates a new ShortenURLUseCase
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

// Shorten creates a short URL from a long URL
func (uc *ShortenURLUseCase) Shorten(ctx context.Context, req *dto.ShortenURLRequest) (*dto.ShortenURLResponse, error) {
	log.Printf("[Shorten] Starting URL shortening process for: %s", req.LongURL)

	// Normalize and validate long URL
	normalizedURL := valueobject.NormalizeURL(req.LongURL)
	log.Printf("[Shorten] Normalized URL: %s", normalizedURL)

	longURL, err := valueobject.NewLongURL(normalizedURL)
	if err != nil {
		log.Printf("[Shorten] Error creating long URL value object: %v", err)
		return nil, err
	}
	log.Printf("[Shorten] Long URL validation successful")

	// Check if URL already exists (only if no custom key is provided)
	if req.CustomKey == "" {
		log.Printf("[Shorten] Checking if URL already exists in database")
		existingURL, err := uc.urlRepo.FindByLongURL(ctx, longURL)
		if err == nil && existingURL != nil && !existingURL.IsExpired() {
			log.Printf("[Shorten] Found existing URL with short key: %s", existingURL.ShortKey.Value())
			return uc.buildResponse(existingURL), nil
		}
		log.Printf("[Shorten] No existing URL found, proceeding with new short URL creation")
	} else {
		log.Printf("[Shorten] Custom key provided, skipping duplicate check for long URL")
	}

	// Generate or use custom short key
	var shortKey *valueobject.ShortKey
	var id int64

	if req.CustomKey != "" {
		log.Printf("[Shorten] Using custom key: %s", req.CustomKey)
		// Use custom key
		shortKey, err = valueobject.NewShortKey(req.CustomKey)
		if err != nil {
			log.Printf("[Shorten] Error creating custom short key: %v", err)
			return nil, err
		}
		log.Printf("[Shorten] Custom short key validation successful")

		// Check if custom key already exists
		log.Printf("[Shorten] Checking if custom key already exists")
		exists, err := uc.urlRepo.ExistsByShortKey(ctx, shortKey)
		if err != nil {
			log.Printf("[Shorten] Error checking custom key existence: %v", err)
			return nil, ErrInternalError
		}
		if exists {
			log.Printf("[Shorten] Custom key already exists: %s", req.CustomKey)
			return nil, ErrCustomKeyExists
		}
		log.Printf("[Shorten] Custom key is available")

		// Generate ID using Snowflake even for custom keys (for database constraints)
		log.Printf("[Shorten] Generating ID for custom key")
		id, err = uc.genService.GenerateID()
		if err != nil {
			log.Printf("[Shorten] Error generating ID: %v", err)
			return nil, ErrInternalError
		}
		log.Printf("[Shorten] Generated ID: %d", id)
	} else {
		log.Printf("[Shorten] Generating short key using generator service")
		// Generate short key
		shortKey, id, err = uc.genService.GenerateShortKey()
		if err != nil {
			log.Printf("[Shorten] Error generating short key: %v", err)
			return nil, ErrInternalError
		}
		log.Printf("[Shorten] Generated short key: %s, ID: %d", shortKey.Value(), id)
	}

	// Create URL entity
	log.Printf("[Shorten] Creating URL entity")
	url := entity.NewURL(shortKey, longURL)
	url.ID = id
	log.Printf("[Shorten] URL entity created with ID: %d, ShortKey: %s", url.ID, url.ShortKey.Value())

	// Set expiration if provided
	if req.ExpiresIn > 0 {
		log.Printf("[Shorten] Setting expiration: %d seconds", req.ExpiresIn)
		url.SetExpiration(time.Duration(req.ExpiresIn) * time.Second)
		log.Printf("[Shorten] Expiration set to: %v", url.ExpiresAt)
	} else {
		log.Printf("[Shorten] No expiration set")
	}

	// Save to database
	log.Printf("[Shorten] Saving URL to database")
	if err := uc.urlRepo.Save(ctx, url); err != nil {
		log.Printf("[Shorten] Error saving URL to database: %v", err)
		return nil, fmt.Errorf("failed to save URL: %w", err)
	}
	log.Printf("[Shorten] URL saved successfully to database")

	// Cache the mapping
	cacheTTL := uc.defaultTTL
	if url.ExpiresAt != nil {
		cacheTTL = time.Until(*url.ExpiresAt)
	}
	log.Printf("[Shorten] Caching URL mapping with TTL: %v", cacheTTL)
	if err := uc.cacheRepo.Set(ctx, shortKey.Value(), longURL.Value(), cacheTTL); err != nil {
		log.Printf("[Shorten] Warning: Failed to cache URL mapping: %v", err)
	} else {
		log.Printf("[Shorten] URL cached successfully")
	}

	log.Printf("[Shorten] URL shortening completed successfully. Short URL: %s/%s", uc.baseURL, shortKey.Value())
	return uc.buildResponse(url), nil
}

// GetLongURL retrieves the long URL from a short key
func (uc *ShortenURLUseCase) GetLongURL(ctx context.Context, shortKeyStr string) (string, error) {
	shortKey, err := valueobject.NewShortKey(shortKeyStr)
	if err != nil {
		return "", err
	}

	// Try cache first
	if cachedURL, err := uc.cacheRepo.Get(ctx, shortKey.Value()); err == nil && cachedURL != "" {
		return cachedURL, nil
	}

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

	// Increment visit count asynchronously
	go func() {
		url.IncrementVisit()
		_ = uc.urlRepo.Update(context.Background(), url)
	}()

	// Update cache
	cacheTTL := uc.defaultTTL
	if url.ExpiresAt != nil {
		cacheTTL = time.Until(*url.ExpiresAt)
	}
	_ = uc.cacheRepo.Set(ctx, shortKey.Value(), url.LongURL.Value(), cacheTTL)

	return url.LongURL.Value(), nil
}

// GetStats retrieves statistics for a short URL
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

	return resp, nil
}

// buildResponse builds a ShortenURLResponse from a URL entity
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
