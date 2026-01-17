// Package config provides configuration management for the URL shortener service.
//
// This package handles loading, parsing, and validation of application
// configuration from multiple sources including YAML files, environment
// variables, and command-line flags. It provides a centralized configuration
// structure that can be easily accessed throughout the application.
//
// Configuration sources (in order of precedence):
//   - Command-line flags
//   - Environment variables
//   - YAML configuration files
//   - Default values
//
// Supported configuration:
//   - Server settings (host, port, timeouts)
//   - Database connection parameters
//   - Redis cache configuration
//   - Rate limiting settings
//   - Logging configuration
//   - Security and CORS settings
//   - ID generation parameters
//
// The configuration is loaded once at application startup and
// provides immutable access to settings throughout the application lifecycle.
package config
