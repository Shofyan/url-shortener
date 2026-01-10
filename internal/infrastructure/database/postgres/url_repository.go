package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Shofyan/url-shortener/internal/domain/entity"
	"github.com/Shofyan/url-shortener/internal/domain/valueobject"

	// Import postgres driver for database/sql
	_ "github.com/lib/pq"
)

var (
	// ErrNotFound is returned when a URL is not found in the database
	ErrNotFound = errors.New("URL not found")
)

// URLRepository implements the URLRepository interface for PostgreSQL
type URLRepository struct {
	db *sql.DB
}

// NewURLRepository creates a new PostgreSQL URL repository
func NewURLRepository(db *sql.DB) *URLRepository {
	return &URLRepository{db: db}
}

// Save saves a new URL mapping
func (r *URLRepository) Save(ctx context.Context, url *entity.URL) error {
	query := `
		INSERT INTO urls (id, short_key, long_url, created_at, expires_at, visit_count)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		url.ID,
		url.ShortKey.Value(),
		url.LongURL.Value(),
		url.CreatedAt,
		url.ExpiresAt,
		url.VisitCount,
	)

	return err
}

// FindByShortKey retrieves a URL by its short key
func (r *URLRepository) FindByShortKey(ctx context.Context, shortKey *valueobject.ShortKey) (*entity.URL, error) {
	query := `
		SELECT id, short_key, long_url, created_at, expires_at, visit_count
		FROM urls
		WHERE short_key = $1
	`

	row := r.db.QueryRowContext(ctx, query, shortKey.Value())

	var (
		id          int64
		shortKeyStr string
		longURLStr  string
		createdAt   time.Time
		expiresAt   sql.NullTime
		visitCount  int64
	)

	err := row.Scan(&id, &shortKeyStr, &longURLStr, &createdAt, &expiresAt, &visitCount)
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

	return url, nil
}

// FindByLongURL retrieves a URL by its long URL
func (r *URLRepository) FindByLongURL(ctx context.Context, longURL *valueobject.LongURL) (*entity.URL, error) {
	query := `
		SELECT id, short_key, long_url, created_at, expires_at, visit_count
		FROM urls
		WHERE long_url = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, longURL.Value())

	var (
		id          int64
		shortKeyStr string
		longURLStr  string
		createdAt   time.Time
		expiresAt   sql.NullTime
		visitCount  int64
	)

	err := row.Scan(&id, &shortKeyStr, &longURLStr, &createdAt, &expiresAt, &visitCount)
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

	return url, nil
}

// Update updates an existing URL
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

// Delete deletes a URL by its short key
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

// ExistsByShortKey checks if a short key already exists
func (r *URLRepository) ExistsByShortKey(ctx context.Context, shortKey *valueobject.ShortKey) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE short_key = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, shortKey.Value()).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// NewDB creates a new database connection
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
