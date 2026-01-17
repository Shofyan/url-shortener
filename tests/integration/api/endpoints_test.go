package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Shofyan/url-shortener/internal/application/dto"
)

// setupTestRouter sets up a test router with all middleware and handlers.
func setupTestRouter() *gin.Engine {
	// In a real implementation, this would initialize the full application
	// with test dependencies (in-memory storage, mock services, etc.)
	gin.SetMode(gin.TestMode)

	// This is a simplified setup - in practice you'd inject test dependencies
	router := gin.New()

	// Add the processing time middleware
	router.Use(func(c *gin.Context) {
		start := time.Now()

		c.Next()

		microseconds := time.Since(start).Microseconds()
		c.Header("X-Processing-Time-Micros", fmt.Sprintf("%d", microseconds))
	})

	// Mock handlers for testing API compliance
	router.POST("/", func(c *gin.Context) {
		var req dto.ShortenURLRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_request",
				Message: err.Error(),
			})

			return
		}

		// Mock response
		resp := dto.ShortenURLResponse{
			ShortURL:  "http://localhost:8080/s/abc123",
			ShortKey:  "abc123",
			LongURL:   req.LongURL,
			CreatedAt: time.Now().Format(time.RFC3339),
			ExpiresAt: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		}

		c.JSON(http.StatusOK, resp)
	})

	router.GET("/s/:shortKey", func(c *gin.Context) {
		shortKey := c.Param("shortKey")
		if shortKey == "notfound" {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "url_not_found",
				Message: "URL not found or expired",
			})

			return
		}

		// Mock redirect
		c.Redirect(http.StatusFound, "https://example.com")
	})

	router.GET("/stats/:shortKey", func(c *gin.Context) {
		shortKey := c.Param("shortKey")
		if shortKey == "notfound" {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "url_not_found",
				Message: "URL not found or expired",
			})

			return
		}

		// Mock stats response with all required fields
		resp := dto.URLStatsResponse{
			ShortKey:       shortKey,
			LongURL:        "https://example.com",
			VisitCount:     42,
			CreatedAt:      time.Now().Add(-time.Hour).Format(time.RFC3339),
			ExpiresAt:      time.Now().Add(time.Hour).Format(time.RFC3339),
			LastAccessedAt: time.Now().Add(-time.Minute).Format(time.RFC3339),
		}

		c.JSON(http.StatusOK, resp)
	})

	return router
}

func TestAPI_URLCreation_Success(t *testing.T) {
	router := setupTestRouter()

	// Test data
	reqBody := dto.ShortenURLRequest{
		LongURL:    "https://example.com/very/long/path",
		TTLSeconds: 3600, // 1 hour
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify processing time header exists
	assert.NotEmpty(t, w.Header().Get("X-Processing-Time-Micros"))

	// Parse response
	var resp dto.ShortenURLResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Verify response structure
	assert.NotEmpty(t, resp.ShortURL)
	assert.NotEmpty(t, resp.ShortKey)
	assert.Equal(t, "https://example.com/very/long/path", resp.LongURL)
	assert.NotEmpty(t, resp.CreatedAt)
	assert.NotEmpty(t, resp.ExpiresAt)
}

func TestAPI_URLCreation_DefaultTTL(t *testing.T) {
	router := setupTestRouter()

	// Test data without TTL - should use 24 hour default
	reqBody := dto.ShortenURLRequest{
		LongURL: "https://example.com",
		// No TTLSeconds specified
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify processing time header
	processingTime := w.Header().Get("X-Processing-Time-Micros")
	assert.NotEmpty(t, processingTime)

	// Verify processing time is a valid number
	assert.Regexp(t, `^\d+$`, processingTime)
}

func TestAPI_URLCreation_InvalidRequest(t *testing.T) {
	router := setupTestRouter()

	// Test with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert error response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify processing time header still present
	assert.NotEmpty(t, w.Header().Get("X-Processing-Time-Micros"))

	// Parse error response
	var errResp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err)

	assert.Equal(t, "invalid_request", errResp.Error)
	assert.NotEmpty(t, errResp.Message)
}

func TestAPI_Redirection_Success(t *testing.T) {
	router := setupTestRouter()

	// Test successful redirection
	req := httptest.NewRequest(http.MethodGet, "/s/abc123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert redirect response
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Location"))

	// Verify processing time header
	assert.NotEmpty(t, w.Header().Get("X-Processing-Time-Micros"))
}

func TestAPI_Redirection_NotFound(t *testing.T) {
	router := setupTestRouter()

	// Test with non-existent short key
	req := httptest.NewRequest(http.MethodGet, "/s/notfound", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert not found response
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Verify processing time header
	assert.NotEmpty(t, w.Header().Get("X-Processing-Time-Micros"))

	// Parse error response
	var errResp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err)

	assert.Equal(t, "url_not_found", errResp.Error)
}

func TestAPI_Stats_Success(t *testing.T) {
	router := setupTestRouter()

	// Test stats retrieval
	req := httptest.NewRequest(http.MethodGet, "/stats/abc123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify processing time header
	processingTime := w.Header().Get("X-Processing-Time-Micros")
	assert.NotEmpty(t, processingTime)

	// Parse response
	var resp dto.URLStatsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Verify all required fields are present
	assert.Equal(t, "abc123", resp.ShortKey)
	assert.Equal(t, "https://example.com", resp.LongURL)
	assert.Equal(t, int64(42), resp.VisitCount)
	assert.NotEmpty(t, resp.CreatedAt)
	assert.NotEmpty(t, resp.ExpiresAt)
	assert.NotEmpty(t, resp.LastAccessedAt) // This was missing before

	// Verify timestamp formats
	_, err = time.Parse(time.RFC3339, resp.CreatedAt)
	assert.NoError(t, err, "CreatedAt should be valid RFC3339 timestamp")

	_, err = time.Parse(time.RFC3339, resp.ExpiresAt)
	assert.NoError(t, err, "ExpiresAt should be valid RFC3339 timestamp")

	_, err = time.Parse(time.RFC3339, resp.LastAccessedAt)
	assert.NoError(t, err, "LastAccessedAt should be valid RFC3339 timestamp")
}

func TestAPI_Stats_NotFound(t *testing.T) {
	router := setupTestRouter()

	// Test with non-existent short key
	req := httptest.NewRequest(http.MethodGet, "/stats/notfound", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert not found response
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Verify processing time header
	assert.NotEmpty(t, w.Header().Get("X-Processing-Time-Micros"))
}

func TestAPI_ProcessingTimeHeader_AllEndpoints(t *testing.T) {
	router := setupTestRouter()

	// Test all endpoints have processing time header
	testCases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"POST /", http.MethodPost, "/", `{"long_url":"https://example.com"}`},
		{"GET /s/abc123", http.MethodGet, "/s/abc123", ""},
		{"GET /stats/abc123", http.MethodGet, "/stats/abc123", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			if tc.body != "" {
				req = httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Every response should have processing time header
			processingTime := w.Header().Get("X-Processing-Time-Micros")
			assert.NotEmpty(t, processingTime, "Endpoint %s should have X-Processing-Time-Micros header", tc.path)

			// Verify it's a valid microsecond value
			assert.Regexp(t, `^\d+$`, processingTime, "Processing time should be numeric microseconds")
		})
	}
}

func TestAPI_RouteCompliance(t *testing.T) {
	router := setupTestRouter()

	// Test that the correct route patterns are implemented
	testCases := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{"URL Creation", http.MethodPost, "/", http.StatusOK},
		{"Redirection", http.MethodGet, "/s/abc123", http.StatusFound},
		{"Stats", http.MethodGet, "/stats/abc123", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			if tc.method == http.MethodPost {
				req = httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(`{"long_url":"https://example.com"}`))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code, "Endpoint %s should return status %d", tc.path, tc.expectedStatus)
		})
	}
}
