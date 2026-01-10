// Package main provides database migration utilities for the URL shortener service.
//
// This package contains the migration tool that manages database schema changes
// by applying SQL migration files in the correct order. It tracks applied
// migrations in a schema_migrations table to ensure consistency across
// different environments.
//
// Features:
//   - Automatic migration discovery from the migrations directory
//   - Sequential migration application based on filename ordering
//   - Migration status tracking to prevent duplicate applications
//   - Support for both up and down migrations
//   - Rollback capabilities for schema changes
//   - PostgreSQL database support
//
// Usage:
//
//	go run cmd/migrate/main.go [up|down] [steps]
//
// The migration files should be placed in internal/infrastructure/database/migrations/
// and follow the naming convention: XXX_description.sql where XXX is a sequential number.
package main
