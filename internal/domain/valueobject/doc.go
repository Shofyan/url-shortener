// Package valueobject defines value objects for the URL shortener domain.
//
// Value objects are immutable objects that represent descriptive aspects
// of the domain with no conceptual identity. They are defined by their
// attributes rather than their identity and provide type safety and
// encapsulation for domain concepts.
//
// This package contains:
//   - ShortKey value object for URL short codes
//   - LongURL value object for original URLs
//   - Validation logic for URL formats and constraints
//   - Immutable data structures with equality semantics
//   - Domain-specific formatting and parsing
//
// Value objects provide:
//   - Type safety to prevent primitive obsession
//   - Encapsulation of validation rules
//   - Immutability for thread safety
//   - Rich behavior for domain operations
//   - Clear expression of domain concepts
//
// These objects ensure data integrity and provide meaningful
// abstractions for URL-related operations throughout the domain.
package valueobject
