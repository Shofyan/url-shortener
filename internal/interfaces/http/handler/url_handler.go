package handler

import (
	"net/http"

	"github.com/Shofyan/url-shortener/internal/application/dto"
	"github.com/Shofyan/url-shortener/internal/application/usecase"
	"github.com/gin-gonic/gin"
)

// URLHandler handles URL shortening HTTP requests
type URLHandler struct {
	useCase *usecase.ShortenURLUseCase
}

// NewURLHandler creates a new URLHandler
func NewURLHandler(useCase *usecase.ShortenURLUseCase) *URLHandler {
	return &URLHandler{
		useCase: useCase,
	}
}

// ShortenURL handles POST /api/shorten requests
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

// RedirectURL handles GET /:shortKey requests
func (h *URLHandler) RedirectURL(c *gin.Context) {
	shortKey := c.Param("shortKey")

	longURL, err := h.useCase.GetLongURL(c.Request.Context(), shortKey)
	if err != nil {
		statusCode := http.StatusNotFound
		errorCode := "not_found"

		switch err {
		case usecase.ErrURLExpired:
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

// GetStats handles GET /api/stats/:shortKey requests
func (h *URLHandler) GetStats(c *gin.Context) {
	shortKey := c.Param("shortKey")

	stats, err := h.useCase.GetStats(c.Request.Context(), shortKey)
	if err != nil {
		statusCode := http.StatusNotFound
		errorCode := "not_found"

		switch err {
		case usecase.ErrURLExpired:
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

// HealthCheck handles GET /health requests
func (h *URLHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "url-shortener",
	})
}
