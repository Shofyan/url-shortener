// Package router provides HTTP routing configuration for the URL shortener service.
//
// This package sets up the HTTP routing infrastructure using the Gin web
// framework, organizing endpoints into logical groups and applying appropriate
// middleware to each route. It serves as the central configuration point
// for all HTTP routes and their associated handlers.
//
// Route organization:
//   - API routes (/api/*) for programmatic access
//   - Web routes (/*) for browser-based interaction
//   - Health check and monitoring endpoints
//   - Static file serving for web assets
//
// Router features:
//   - RESTful route organization and versioning
//   - Middleware composition and route-specific configuration
//   - Content-type negotiation and response formatting
//   - Route parameter validation and binding
//   - Error handling and standardized responses
//   - OpenAPI/Swagger documentation generation
//
// The router maintains clean separation between different types of
// endpoints while providing consistent behavior across all routes.
package router
