// Package main provides the main entry point for the URL shortener API server.
//
// This package initializes and starts the HTTP server with all necessary dependencies
// including database connections, cache, generators, and middleware. It handles
// graceful shutdown and dependency injection for the application.
//
// The server supports:
//   - RESTful API endpoints for URL shortening
//   - Web interface for user-friendly access
//   - Configurable rate limiting and CORS
//   - Database migration support
//   - Redis caching for performance
//   - Multiple ID generation strategies
//
// Configuration is loaded from YAML files and environment variables.
// The server gracefully handles shutdown signals for clean termination.
package main
