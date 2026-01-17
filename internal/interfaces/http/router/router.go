package router

import (
	"github.com/Shofyan/url-shortener/internal/interfaces/http/handler"
	"github.com/Shofyan/url-shortener/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures all routes and middleware
func SetupRouter(urlHandler *handler.URLHandler, webHandler *handler.WebHandler, rateLimiter *middleware.RateLimiter) *gin.Engine {
	// Set Gin to release mode in production
	// gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Global middleware
	router.Use(middleware.Recovery())
	router.Use(middleware.ProcessingTime())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	// Load HTML templates
	router.LoadHTMLGlob("web/templates/*")

	// Serve static files
	router.Static("/static", "web/static")

	// Health check endpoint (no rate limiting)
	router.GET("/health", urlHandler.HealthCheck)

	// URL Creation endpoint (POST /)
	router.POST("/", rateLimiter.Limit(), urlHandler.ShortenURL)

	// Short URL redirect (GET /s/{short_code})
	router.GET("/s/:shortKey", rateLimiter.Limit(), urlHandler.RedirectURL)

	// Stats endpoint (GET /stats/{short_code})
	router.GET("/stats/:shortKey", urlHandler.GetStats)

	// Web UI routes (serve after API routes to avoid conflicts)
	router.GET("/web", webHandler.Index)
	router.GET("/web/*filepath", webHandler.Index)

	return router
}
