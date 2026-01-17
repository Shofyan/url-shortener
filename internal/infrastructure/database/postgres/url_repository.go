package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/Shofyan/url-shortener/internal/domain/entity"
	"github.com/Shofyan/url-shortener/internal/domain/valueobject"

	// Import postgres driver for database/sql.
	_ "github.com/lib/pq"
)

var (
	// ErrNotFound is returned when a URL is not found in the database.
	ErrNotFound = errors.New("URL not found")
)

// URLRepository implements the URLRepository interface for PostgreSQL.
type URLRepository struct {
	db *sql.DB
}

// NewURLRepository creates a new PostgreSQL URL repository.
func NewURLRepository(db *sql.DB) *URLRepository {
	return &URLRepository{db: db}
}

// Save saves a new URL mapping.
func (r *URLRepository) Save(ctx context.Context, url *entity.URL) error {
	query := `
		INSERT INTO urls (id, short_key, long_url, created_at, expires_at, visit_count, last_accessed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		url.ID,
		url.ShortKey.Value(),
		url.LongURL.Value(),
		url.CreatedAt,
		url.ExpiresAt,
		url.VisitCount,
		url.LastAccessedAt,
	)

	return err
}

// FindByShortKey retrieves a URL by its short key.
func (r *URLRepository) FindByShortKey(ctx context.Context, shortKey *valueobject.ShortKey) (*entity.URL, error) {
	query := `
		SELECT id, short_key, long_url, created_at, expires_at, visit_count, last_accessed_at
		FROM urls
		WHERE short_key = $1
	`

	row := r.db.QueryRowContext(ctx, query, shortKey.Value())

	var (
		id             int64
		shortKeyStr    string
		longURLStr     string
		createdAt      time.Time
		expiresAt      sql.NullTime
		visitCount     int64
		lastAccessedAt sql.NullTime
	)

	err := row.Scan(&id, &shortKeyStr, &longURLStr, &createdAt, &expiresAt, &visitCount, &lastAccessedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, err
	}

	sk, _ := valueobject.NewShortKey(shortKeyStr)
	lu, _ := valueobject.NewLongURL(longURLStr)

	url := &entity.URL{
		ID:         id,
		ShortKey:   sk,
		LongURL:    lu,
		CreatedAt:  createdAt,
		VisitCount: visitCount,
	}

	if expiresAt.Valid {
		url.ExpiresAt = &expiresAt.Time
	}

	if lastAccessedAt.Valid {
		url.LastAccessedAt = &lastAccessedAt.Time
	}

	return url, nil
}

// FindByLongURL retrieves a URL by its long URL.
func (r *URLRepository) FindByLongURL(ctx context.Context, longURL *valueobject.LongURL) (*entity.URL, error) {
	query := `
		SELECT id, short_key, long_url, created_at, expires_at, visit_count, last_accessed_at
		FROM urls
		WHERE long_url = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, longURL.Value())

	var (
		id             int64
		shortKeyStr    string
		longURLStr     string
		createdAt      time.Time
		expiresAt      sql.NullTime
		visitCount     int64
		lastAccessedAt sql.NullTime
	)

	err := row.Scan(&id, &shortKeyStr, &longURLStr, &createdAt, &expiresAt, &visitCount, &lastAccessedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, err
	}

	sk, _ := valueobject.NewShortKey(shortKeyStr)
	lu, _ := valueobject.NewLongURL(longURLStr)

	url := &entity.URL{
		ID:         id,
		ShortKey:   sk,
		LongURL:    lu,
		CreatedAt:  createdAt,
		VisitCount: visitCount,
	}

	if expiresAt.Valid {
		url.ExpiresAt = &expiresAt.Time
	}

	if lastAccessedAt.Valid {
		url.LastAccessedAt = &lastAccessedAt.Time
	}

	return url, nil
}

// Update updates an existing URL.
func (r *URLRepository) Update(ctx context.Context, url *entity.URL) error {
	query := `
		UPDATE urls
		SET long_url = $1, expires_at = $2, visit_count = $3
		WHERE short_key = $4
	`

	result, err := r.db.ExecContext(ctx, query,
		url.LongURL.Value(),
		url.ExpiresAt,
		url.VisitCount,
		url.ShortKey.Value(),
	)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// Delete deletes a URL by its short key.
func (r *URLRepository) Delete(ctx context.Context, shortKey *valueobject.ShortKey) error {
	query := `DELETE FROM urls WHERE short_key = $1`

	result, err := r.db.ExecContext(ctx, query, shortKey.Value())
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// ExistsByShortKey checks if a short key already exists.
func (r *URLRepository) ExistsByShortKey(ctx context.Context, shortKey *valueobject.ShortKey) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE short_key = $1)`

	var exists bool

	err := r.db.QueryRowContext(ctx, query, shortKey.Value()).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// IncrementVisitCount atomically increments visit count and updates last_accessed_at.
func (r *URLRepository) IncrementVisitCount(ctx context.Context, shortKey *valueobject.ShortKey) error {
	query := `
		UPDATE urls
		SET visit_count = visit_count + 1,
			last_accessed_at = CURRENT_TIMESTAMP
		WHERE short_key = $1
	`

	result, err := r.db.ExecContext(ctx, query, shortKey.Value())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// FindExpiredURLs returns URLs that expired before the given timestamp.
func (r *URLRepository) FindExpiredURLs(ctx context.Context, before time.Time, maxResults int) ([]*entity.URL, error) {
	query := `
		SELECT id, short_key, long_url, created_at, expires_at, visit_count, last_accessed_at
		FROM urls
		WHERE expires_at IS NOT NULL AND expires_at < $1
		ORDER BY expires_at ASC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, before, maxResults)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []*entity.URL

	for rows.Next() {
		var url entity.URL

		var shortKeyValue, longURLValue string

		err := rows.Scan(
			&url.ID,
			&shortKeyValue,
			&longURLValue,
			&url.CreatedAt,
			&url.ExpiresAt,
			&url.VisitCount,
			&url.LastAccessedAt,
		)
		if err != nil {
			return nil, err
		}

		// Create value objects
		shortKey, err := valueobject.NewShortKey(shortKeyValue)
		if err != nil {
			return nil, err
		}

		longURL, err := valueobject.NewLongURL(longURLValue)
		if err != nil {
			return nil, err
		}

		url.ShortKey = shortKey
		url.LongURL = longURL

		urls = append(urls, &url)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

// DeleteExpiredBatch deletes multiple URLs by their short keys in a single transaction.
func (r *URLRepository) DeleteExpiredBatch(ctx context.Context, shortKeys []*valueobject.ShortKey) error {
	if len(shortKeys) == 0 {
		return nil
	}

	// Begin transaction for batch delete
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Printf("Warning: Failed to rollback transaction: %v", rollbackErr)
		}
	}()

	// Use batch delete with IN clause for better performance
	query := `DELETE FROM urls WHERE short_key = ANY($1)`

	// Convert short keys to string array
	keyValues := make([]string, len(shortKeys))
	for i, key := range shortKeys {
		keyValues[i] = key.Value()
	}

	result, err := tx.ExecContext(ctx, query, keyValues)
	if err != nil {
		return err
	}

	// Log the number of deleted records for monitoring
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected != int64(len(shortKeys)) {
		// Some records might have already been deleted - this is acceptable
		// in a concurrent environment where multiple cleanup processes might run
		log.Printf("Deleted %d out of %d requested URLs (some may have been already deleted)", rowsAffected, len(shortKeys))
	}

	return tx.Commit()
}

// GetExpiredCount returns the total count of expired URLs for monitoring.
func (r *URLRepository) GetExpiredCount(ctx context.Context, before time.Time) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM urls
		WHERE expires_at IS NOT NULL AND expires_at < $1
	`

	var count int64

	err := r.db.QueryRowContext(ctx, query, before).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// NewDB creates a new database connection.
func NewDB(dsn string, maxOpenConns, maxIdleConns int, connMaxLifetime time.Duration) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
