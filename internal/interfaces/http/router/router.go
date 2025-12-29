package router

import (
	"github.com/Shofyan/url-shortener/internal/interfaces/http/handler"
	"github.com/Shofyan/url-shortener/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures all routes and middleware
func SetupRouter(urlHandler *handler.URLHandler, rateLimiter *middleware.RateLimiter) *gin.Engine {
	// Set Gin to release mode in production
	// gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Global middleware
	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	// Health check endpoint (no rate limiting)
	router.GET("/health", urlHandler.HealthCheck)

	// API routes with rate limiting
	api := router.Group("/api")
	api.Use(rateLimiter.Limit())
	{
		// Create short URL
		api.POST("/shorten", urlHandler.ShortenURL)

		// Get URL statistics
		api.GET("/stats/:shortKey", urlHandler.GetStats)
	}

	// Short URL redirect (with rate limiting)
	router.GET("/:shortKey", rateLimiter.Limit(), urlHandler.RedirectURL)

	return router
}
