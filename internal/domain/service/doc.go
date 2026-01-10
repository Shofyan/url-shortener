// Package service provides domain services for the URL shortener application.
//
// Domain services contain business logic that doesn't naturally fit within
// a single entity or value object. They operate on domain objects and
// coordinate complex business operations that span multiple entities.
//
// This package includes:
//   - ID generation services for creating unique identifiers
//   - URL validation and processing services
//   - Business rule enforcement across entities
//   - Complex domain calculations and algorithms
//
// Domain services are:
//   - Stateless and focused on specific domain operations
//   - Independent of external infrastructure concerns
//   - Testable through dependency injection
//   - Reusable across different use cases
//
// Services in this layer collaborate with entities and value objects
// to implement core business functionality while maintaining clean
// separation from application and infrastructure concerns.
package service
