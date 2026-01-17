package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Shofyan/url-shortener/internal/application/dto"
	"github.com/Shofyan/url-shortener/internal/application/usecase"
	"github.com/Shofyan/url-shortener/internal/domain/service"
)

// URLHandler handles URL shortening HTTP requests.
type URLHandler struct {
	useCase        *usecase.ShortenURLUseCase
	cleanupService service.URLCleanupService
}

// NewURLHandler creates a new URLHandler.
func NewURLHandler(useCase *usecase.ShortenURLUseCase, cleanupService service.URLCleanupService) *URLHandler {
	return &URLHandler{
		useCase:        useCase,
		cleanupService: cleanupService,
	}
}

// ShortenURL handles POST /api/shorten requests.
func (h *URLHandler) ShortenURL(c *gin.Context) {
	var req dto.ShortenURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})

		return
	}

	resp, err := h.useCase.Shorten(c.Request.Context(), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorCode := "internal_error"
		errorMessage := err.Error()

		switch err {
		case usecase.ErrCustomKeyExists:
			statusCode = http.StatusConflict
			errorCode = "custom_key_exists"
		default:
			// Log the actual error for debugging
			_ = c.Error(err)
		}

		c.JSON(statusCode, dto.ErrorResponse{
			Error:   errorCode,
			Message: errorMessage,
		})

		return
	}

	c.JSON(http.StatusCreated, resp)
}

// RedirectURL handles GET /:shortKey requests.
func (h *URLHandler) RedirectURL(c *gin.Context) {
	shortKey := c.Param("shortKey")

	longURL, err := h.useCase.GetLongURL(c.Request.Context(), shortKey)
	if err != nil {
		statusCode := http.StatusNotFound
		errorCode := "not_found"

		if err == usecase.ErrURLExpired {
			statusCode = http.StatusGone
			errorCode = "url_expired"
		}

		c.JSON(statusCode, dto.ErrorResponse{
			Error:   errorCode,
			Message: err.Error(),
		})

		return
	}

	// 302 redirect for temporary redirect (allows tracking)
	// Use 301 for permanent redirect if tracking is not needed
	c.Redirect(http.StatusFound, longURL)
}

// GetStats handles GET /api/stats/:shortKey requests.
func (h *URLHandler) GetStats(c *gin.Context) {
	shortKey := c.Param("shortKey")

	stats, err := h.useCase.GetStats(c.Request.Context(), shortKey)
	if err != nil {
		statusCode := http.StatusNotFound
		errorCode := "not_found"

		if err == usecase.ErrURLExpired {
			statusCode = http.StatusGone
			errorCode = "url_expired"
		}

		c.JSON(statusCode, dto.ErrorResponse{
			Error:   errorCode,
			Message: err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, stats)
}

// HealthCheck handles GET /health requests.
func (h *URLHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "url-shortener",
	})
}

// GetCleanupStats handles GET /api/admin/cleanup/stats requests.
func (h *URLHandler) GetCleanupStats(c *gin.Context) {
	if h.cleanupService == nil {
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{
			Error:   "service_unavailable",
			Message: "Cleanup service is not available",
		})

		return
	}

	stats := h.cleanupService.GetCleanupStats()
	c.JSON(http.StatusOK, stats)
}

// TriggerManualCleanup handles POST /api/admin/cleanup/manual requests.
func (h *URLHandler) TriggerManualCleanup(c *gin.Context) {
	if h.cleanupService == nil {
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{
			Error:   "service_unavailable",
			Message: "Cleanup service is not available",
		})

		return
	}

	var req struct {
		BatchSize int `json:"batch_size"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})

		return
	}

	// Set default batch size if not provided or invalid
	if req.BatchSize <= 0 || req.BatchSize > 10000 {
		req.BatchSize = 1000
	}

	start := time.Now()
	cleanedCount, err := h.cleanupService.CleanupExpiredBatch(c.Request.Context(), req.BatchSize)
	duration := time.Since(start)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "cleanup_failed",
			Message: err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cleaned_count": cleanedCount,
		"batch_size":    req.BatchSize,
		"duration_ms":   float64(duration.Nanoseconds()) / 1000000,
		"timestamp":     time.Now().UTC(),
	})
}
