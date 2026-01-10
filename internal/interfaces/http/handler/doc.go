// Package handler provides HTTP request handlers for the URL shortener API.
//
// This package contains the HTTP handlers that process incoming requests,
// validate input data, coordinate with use cases, and format responses.
// Handlers serve as the entry point for external interactions and implement
// the REST API endpoints for the URL shortener service.
//
// API endpoints provided:
//   - POST /api/shorten - Create shortened URLs
//   - GET /api/urls/{id} - Retrieve URL information
//   - GET /{shortKey} - Redirect to original URL
//   - GET /api/analytics/{shortKey} - Get click analytics
//
// Handler responsibilities:
//   - HTTP request parsing and validation
//   - Authentication and authorization
//   - Input sanitization and security checks
//   - Use case orchestration and error handling
//   - Response formatting and status codes
//   - Content negotiation and CORS handling
//
// The handlers maintain clean separation between HTTP concerns
// and business logic, delegating domain operations to use cases.
package handler
