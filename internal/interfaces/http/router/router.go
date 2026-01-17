package router

import (
	"os"

	"github.com/gin-gonic/gin"

	"github.com/Shofyan/url-shortener/internal/infrastructure/config"
	"github.com/Shofyan/url-shortener/internal/interfaces/http/handler"
	"github.com/Shofyan/url-shortener/internal/interfaces/http/middleware"
)

// SetupRouter configures all routes and middleware.
func SetupRouter(cfg *config.Config, urlHandler *handler.URLHandler, webHandler *handler.WebHandler, rateLimiter *middleware.RateLimiter) *gin.Engine {
	// Set Gin mode - prioritize environment variable, then config, then default to release
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		gin.SetMode(mode)
	} else if cfg.App.GinMode != "" {
		gin.SetMode(cfg.App.GinMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

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
