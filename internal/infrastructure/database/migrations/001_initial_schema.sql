-- Initial Database Schema for URL Shortener
-- This migration creates the complete database structure with all necessary tables,
-- indexes, and constraints in a single consolidated migration file.

-- Create urls table
CREATE TABLE IF NOT EXISTS urls (
    id BIGINT PRIMARY KEY,
    short_key VARCHAR(12) UNIQUE NOT NULL,
    long_url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    visit_count BIGINT NOT NULL DEFAULT 0,
    last_accessed_at TIMESTAMP
);

-- Create indexes on urls table
CREATE INDEX IF NOT EXISTS idx_urls_short_key ON urls(short_key);
CREATE INDEX IF NOT EXISTS idx_urls_long_url ON urls(long_url);
CREATE INDEX IF NOT EXISTS idx_urls_expires_at ON urls(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_urls_created_at ON urls(created_at);
CREATE INDEX IF NOT EXISTS idx_urls_last_accessed_at ON urls(last_accessed_at);

-- Add comments for documentation
COMMENT ON TABLE urls IS 'Main table storing URL mappings and metadata';
COMMENT ON COLUMN urls.id IS 'Unique identifier for URL record';
COMMENT ON COLUMN urls.short_key IS 'Short key used in URLs (up to 12 characters)';
COMMENT ON COLUMN urls.long_url IS 'Original long URL to redirect to';
COMMENT ON COLUMN urls.visit_count IS 'Number of times this URL has been accessed';
COMMENT ON COLUMN urls.last_accessed_at IS 'Timestamp of last URL access for statistics';
