package valueobject

import (
	"errors"
	"net/url"
	"strings"
)

var (
	// ErrInvalidURL is returned when the provided URL format is invalid.
	ErrInvalidURL = errors.New("invalid URL format")
	// ErrEmptyURL is returned when the provided URL is empty.
	ErrEmptyURL = errors.New("URL cannot be empty")
	// ErrURLTooLong is returned when the URL exceeds the maximum allowed length.
	ErrURLTooLong = errors.New("URL exceeds maximum length")
	// ErrInvalidShortKey is returned when the provided short key format is invalid.
	ErrInvalidShortKey = errors.New("invalid short key format")
	// ErrEmptyShortKey is returned when the provided short key is empty.
	ErrEmptyShortKey = errors.New("short key cannot be empty")
)

const (
	// MaxURLLength defines the maximum allowed length for a URL.
	MaxURLLength = 2048
	// MaxShortKeyLength defines the maximum allowed length for a short key.
	MaxShortKeyLength = 12
)

// LongURL represents the original URL value object.
type LongURL struct {
	value string
}

// NewLongURL creates a new LongURL value object with validation.
func NewLongURL(rawURL string) (*LongURL, error) {
	if rawURL == "" {
		return nil, ErrEmptyURL
	}

	if len(rawURL) > MaxURLLength {
		return nil, ErrURLTooLong
	}

	// Validate URL format
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return nil, ErrInvalidURL
	}

	// Ensure it has a scheme
	if parsedURL.Scheme == "" {
		return nil, ErrInvalidURL
	}

	return &LongURL{value: rawURL}, nil
}

// Value returns the string value of the LongURL.
func (l *LongURL) Value() string {
	return l.value
}

// ShortKey represents the shortened URL key value object.
type ShortKey struct {
	value string
}

// NewShortKey creates a new ShortKey value object with validation.
func NewShortKey(key string) (*ShortKey, error) {
	if key == "" {
		return nil, ErrEmptyShortKey
	}

	if len(key) > MaxShortKeyLength {
		return nil, ErrInvalidShortKey
	}

	// Validate that it only contains alphanumeric characters
	if !isAlphanumeric(key) {
		return nil, ErrInvalidShortKey
	}

	return &ShortKey{value: key}, nil
}

// Value returns the string value of the ShortKey.
func (s *ShortKey) Value() string {
	return s.value
}

// isAlphanumeric checks if a string contains only alphanumeric characters, hyphens, and underscores.
func isAlphanumeric(s string) bool {
	for _, char := range s {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '-' && char != '_' {
			return false
		}
	}

	return true
}

// NormalizeURL normalizes a URL by ensuring it has a scheme.
func NormalizeURL(rawURL string) string {
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		return "https://" + rawURL
	}

	return rawURL
}
