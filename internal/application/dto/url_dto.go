package dto

// ShortenURLRequest represents the request to shorten a URL
type ShortenURLRequest struct {
	LongURL    string `json:"long_url" binding:"required"`
	CustomKey  string `json:"custom_key,omitempty"`
	TTLSeconds int64  `json:"ttl_seconds,omitempty"` // Time-to-live in seconds
}

// ShortenURLResponse represents the response after shortening a URL
type ShortenURLResponse struct {
	ShortURL  string `json:"short_url"`
	ShortKey  string `json:"short_key"`
	LongURL   string `json:"long_url"`
	CreatedAt string `json:"created_at"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

// URLStatsResponse represents URL statistics
type URLStatsResponse struct {
	ShortKey       string `json:"short_key"`
	LongURL        string `json:"long_url"`
	VisitCount     int64  `json:"visit_count"`
	CreatedAt      string `json:"created_at"`
	ExpiresAt      string `json:"expires_at,omitempty"`
	LastAccessedAt string `json:"last_accessed_at,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
