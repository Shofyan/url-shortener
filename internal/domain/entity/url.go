package entity

import (
	"time"

	"github.com/Shofyan/url-shortener/internal/domain/valueobject"
)

// URL represents the URL entity in the domain
type URL struct {
	ID         int64
	ShortKey   *valueobject.ShortKey
	LongURL    *valueobject.LongURL
	CreatedAt  time.Time
	ExpiresAt  *time.Time
	VisitCount int64
}

// NewURL creates a new URL entity
func NewURL(shortKey *valueobject.ShortKey, longURL *valueobject.LongURL) *URL {
	return &URL{
		ShortKey:   shortKey,
		LongURL:    longURL,
		CreatedAt:  time.Now(),
		VisitCount: 0,
	}
}

// SetExpiration sets an expiration time for the URL
func (u *URL) SetExpiration(duration time.Duration) {
	expiresAt := time.Now().Add(duration)
	u.ExpiresAt = &expiresAt
}

// IsExpired checks if the URL has expired
func (u *URL) IsExpired() bool {
	if u.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*u.ExpiresAt)
}

// IncrementVisit increments the visit count
func (u *URL) IncrementVisit() {
	u.VisitCount++
}
