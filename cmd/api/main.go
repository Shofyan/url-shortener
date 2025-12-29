package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Shofyan/url-shortener/internal/application/usecase"
	"github.com/Shofyan/url-shortener/internal/domain/service"
	"github.com/Shofyan/url-shortener/internal/infrastructure/cache/redis"
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
	defer db.Close()
	log.Println("✓ Connected to PostgreSQL")

	// Initialize Redis
	redisClient, err := redis.NewRedisClient(
		cfg.Redis.GetRedisAddr(),
		cfg.Redis.Password,
		cfg.Redis.DB,
		cfg.Redis.PoolSize,
		cfg.Redis.MinIdleConns,
	)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("✓ Connected to Redis")

	// Initialize repositories
	urlRepo := postgres.NewURLRepository(db)
	cacheRepo := redis.NewCacheRepository(redisClient)

	// Initialize generators
	snowflakeGen, err := snowflake.NewGenerator(cfg.App.SnowflakeNodeID)
	if err != nil {
		log.Fatalf("Failed to create Snowflake generator: %v", err)
	}
	base62Gen := base62.NewGenerator()

	// Initialize domain services
	generatorService := service.NewGeneratorService(snowflakeGen, base62Gen)

	// Initialize use cases
	shortenUseCase := usecase.NewShortenURLUseCase(
		urlRepo,
		cacheRepo,
		generatorService,
		cfg.App.BaseURL,
		cfg.App.CacheTTL,
	)

	// Initialize handlers
	urlHandler := handler.NewURLHandler(shortenUseCase)
	webHandler := handler.NewWebHandler()

	// Initialize middleware
	rateLimiter := middleware.NewRateLimiter(cfg.App.RateLimitRequests, cfg.App.RateLimitRequests)

	// Setup router
	r := router.SetupRouter(urlHandler, webHandler, rateLimiter)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
