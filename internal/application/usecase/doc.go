// Package usecase implements application use cases for the URL shortener service.
//
// This package contains the business logic orchestration layer, coordinating
// between domain services, repositories, and external dependencies to fulfill
// specific application requirements. Use cases are the entry points for all
// business operations and define the application's behavior.
//
// Key responsibilities:
//   - Orchestrate complex business workflows
//   - Coordinate multiple domain services and repositories
//   - Handle transaction boundaries and error management
//   - Implement application-specific business rules
//   - Provide clean interfaces for external layers
//
// Use cases in this package:
//   - URL shortening with validation and persistence
//   - URL retrieval with caching strategies
//   - Analytics and click tracking
//   - URL expiration management
//
// Each use case is designed to be testable, reusable, and independent
// of external frameworks or delivery mechanisms.
package usecase
