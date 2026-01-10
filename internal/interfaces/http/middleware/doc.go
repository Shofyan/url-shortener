// Package middleware provides HTTP middleware components for the URL shortener service.
//
// Middleware components intercept HTTP requests and responses to provide
// cross-cutting concerns such as logging, rate limiting, security, and
// error handling. They form a processing pipeline that enhances the
// core functionality with operational and security features.
//
// Available middleware:
//   - CORS: Cross-Origin Resource Sharing configuration
//   - Rate Limiting: Request throttling and abuse prevention
//   - Logger: Request/response logging and monitoring
//   - Recovery: Panic recovery and graceful error handling
//   - Authentication: User authentication and session management
//   - Compression: Response compression for bandwidth optimization
//
// Middleware features:
//   - Configurable rate limits per IP/user
//   - Structured logging with request tracing
//   - Comprehensive CORS policy management
//   - Graceful error recovery and reporting
//   - Performance monitoring and metrics collection
//
// All middleware components are designed to be composable and
// can be easily enabled or disabled through configuration.
package middleware
