package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/Shofyan/url-shortener/internal/application/usecase"
	"github.com/Shofyan/url-shortener/internal/domain/service"
	redisCache "github.com/Shofyan/url-shortener/internal/infrastructure/cache/redis"
	"github.com/Shofyan/url-shortener/internal/infrastructure/config"
	"github.com/Shofyan/url-shortener/internal/infrastructure/database/postgres"
	"github.com/Shofyan/url-shortener/internal/infrastructure/generator/base62"
	"github.com/Shofyan/url-shortener/internal/infrastructure/generator/snowflake"
	"github.com/Shofyan/url-shortener/internal/interfaces/http/handler"
	"github.com/Shofyan/url-shortener/internal/interfaces/http/middleware"
	"github.com/Shofyan/url-shortener/internal/interfaces/http/router"
)

func main() {
	// Load configuration
	cfg, err := config.Load("./config")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize dependencies
	db, redisClient := initializeDependencies(cfg)
	defer closeDependencies(db, redisClient)

	// Initialize and start services
	srv, cleanupService := initializeServices(cfg, db, redisClient)
	startServer(srv, cleanupService)
}

// initializeDependencies sets up database and Redis connections.
func initializeDependencies(cfg *config.Config) (*sql.DB, *redis.Client) {
	// Initialize database
	db, err := postgres.NewDB(
		cfg.Database.GetDSN(),
		cfg.Database.MaxOpenConns,
		cfg.Database.MaxIdleConns,
		cfg.Database.ConnMaxLifetime,
	)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("✓ Connected to PostgreSQL")

	// Initialize Redis
	redisClient, err := redisCache.NewRedisClient(
		cfg.Redis.GetRedisAddr(),
		cfg.Redis.Password,
		cfg.Redis.DB,
		cfg.Redis.PoolSize,
		cfg.Redis.MinIdleConns,
	)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("✓ Connected to Redis")

	return db, redisClient
}

// closeDependencies closes database and Redis connections.
func closeDependencies(db *sql.DB, redisClient *redis.Client) {
	if err := db.Close(); err != nil {
		log.Printf("Failed to close database connection: %v", err)
	}

	if err := redisClient.Close(); err != nil {
		log.Printf("Failed to close Redis connection: %v", err)
	}
}

// initializeServices sets up all services and HTTP server.
func initializeServices(cfg *config.Config, db *sql.DB, redisClient *redis.Client) (*http.Server, *service.BackgroundURLCleanupService) {
	// Initialize repositories
	urlRepo := postgres.NewURLRepository(db)
	cacheRepo := redisCache.NewCacheRepository(redisClient)

	// Initialize generators
	snowflakeGen, err := snowflake.NewGenerator(cfg.App.SnowflakeNodeID)
	if err != nil {
		log.Fatalf("Failed to create Snowflake generator: %v", err)
	}

	base62Gen := base62.NewGenerator()
	generatorService := service.NewGeneratorService(snowflakeGen, base62Gen)

	// Initialize cleanup service
	cleanupService := service.NewBackgroundURLCleanupService(
		urlRepo,
		cacheRepo,
		cfg.App.GetCleanupConfig(),
	)

	// Initialize use cases
	shortenUseCase := usecase.NewShortenURLUseCase(
		urlRepo,
		cacheRepo,
		generatorService,
		cfg.App.BaseURL,
		cfg.App.CacheTTL,
	)

	// Initialize handlers and middleware
	urlHandler := handler.NewURLHandler(shortenUseCase, cleanupService)
	webHandler := handler.NewWebHandler()
	rateLimiter := middleware.NewRateLimiter(cfg.App.RateLimitRequests, cfg.App.RateLimitRequests)

	// Setup router and server
	r := router.SetupRouter(cfg, urlHandler, webHandler, rateLimiter)
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return srv, cleanupService
}

// startServer starts the HTTP server and cleanup service with graceful shutdown.
func startServer(srv *http.Server, cleanupService *service.BackgroundURLCleanupService) {
	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Server starting on port %s", strings.TrimPrefix(srv.Addr, ":"))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Start cleanup service
	ctx := context.Background()
	if err := cleanupService.StartCleanup(ctx); err != nil {
		log.Printf("Warning: Failed to start cleanup service: %v", err)
	} else {
		log.Println("✓ URL cleanup service started")
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Stop cleanup service first
	if err := cleanupService.StopCleanup(); err != nil {
		log.Printf("Warning: Failed to stop cleanup service: %v", err)
	} else {
		log.Println("✓ URL cleanup service stopped")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	if err := srv.Shutdown(ctx); err != nil {
		cancel()
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	cancel()

	log.Println("Server exited")
}
