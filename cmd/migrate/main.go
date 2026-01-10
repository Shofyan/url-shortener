package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

const (
	migrationsDir   = "./internal/infrastructure/database/migrations"
	migrationsTable = "schema_migrations"
)

type migration struct {
	version  string
	filename string
	content  string
}

func main() {
	// Get database connection string from environment or use default
	dsn := getEnvOrDefault("DATABASE_DSN",
		"host=localhost port=5432 user=postgres password=postgres dbname=urlshortener sslmode=disable")

	log.Println("Connecting to database...")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close database connection: %v", err)
		}
	}()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✓ Database connection successful")

	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		log.Fatalf("Failed to get applied migrations: %v", err)
	}

	// Get pending migrations
	migrations, err := getPendingMigrations(appliedMigrations)
	if err != nil {
		log.Fatalf("Failed to get pending migrations: %v", err)
	}

	if len(migrations) == 0 {
		log.Println("✓ No pending migrations")
		return
	}

	// Apply migrations
	log.Printf("Found %d pending migration(s)\n", len(migrations))
	for _, m := range migrations {
		if err := applyMigration(db, m); err != nil {
			log.Fatalf("Failed to apply migration %s: %v", m.filename, err)
		}
		log.Printf("✓ Applied migration: %s", m.filename)
	}

	log.Println("✓ All migrations completed successfully")
}

func createMigrationsTable(db *sql.DB) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`, migrationsTable)

	_, err := db.Exec(query)
	return err
}

func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	query := fmt.Sprintf("SELECT version FROM %s", migrationsTable)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			log.Printf("Failed to close rows: %v", err)
		}
	}()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

func getPendingMigrations(appliedMigrations map[string]bool) ([]migration, error) {
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %v", err)
	}

	var migrations []migration
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		version := strings.TrimSuffix(file.Name(), ".sql")
		if appliedMigrations[version] {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %v", file.Name(), err)
		}

		migrations = append(migrations, migration{
			version:  version,
			filename: file.Name(),
			content:  string(content),
		})
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	return migrations, nil
}

func applyMigration(db *sql.DB, m migration) error {
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Rollback()
		if err != nil && err != sql.ErrTxDone {
			log.Printf("Failed to rollback transaction: %v", err)
		}
	}()

	// Execute migration
	if _, err := tx.Exec(m.content); err != nil {
		return fmt.Errorf("failed to execute migration: %v", err)
	}

	// Record migration
	query := fmt.Sprintf("INSERT INTO %s (version) VALUES ($1)", migrationsTable)
	if _, err := tx.Exec(query, m.version); err != nil {
		return fmt.Errorf("failed to record migration: %v", err)
	}

	// Commit transaction
	return tx.Commit()
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
