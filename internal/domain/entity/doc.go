// Package entity defines the core domain entities for the URL shortener service.
//
// Entities represent the fundamental business objects that have identity
// and lifecycle within the domain. They encapsulate the most important
// business rules and maintain consistency of their internal state.
//
// This package contains:
//   - URL entity with business rules and behaviors
//   - Entity validation and state management
//   - Identity management for domain objects
//   - Core business logic that doesn't depend on external systems
//
// Entities are:
//   - Framework-independent and pure Go structs
//   - Self-validating with invariant protection
//   - Rich in behavior, not just data containers
//   - Persistent across different use cases
//
// The entities in this package form the heart of the domain model
// and should be protected from external concerns and dependencies.
package entity
