// Package dto provides Data Transfer Objects for the URL shortener application layer.
//
// DTOs serve as contracts between the application layer and external interfaces,
// ensuring clean separation between domain entities and API representations.
// They handle data validation, serialization, and transformation between
// different layers of the application.
//
// This package contains:
//   - Request DTOs for API endpoints
//   - Response DTOs with proper JSON tags
//   - Validation rules and constraints
//   - Data transformation utilities
//
// DTOs in this package should be lightweight, focused on data transfer,
// and free from business logic. They provide a stable interface for
// external consumers while allowing internal domain models to evolve
// independently.
package dto
