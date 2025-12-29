-- Create urls table
CREATE TABLE IF NOT EXISTS urls (
    id BIGINT PRIMARY KEY,
    short_key VARCHAR(12) UNIQUE NOT NULL,
    long_url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    visit_count BIGINT NOT NULL DEFAULT 0
);

-- Create index on short_key for fast lookups
CREATE INDEX IF NOT EXISTS idx_urls_short_key ON urls(short_key);

-- Create index on long_url for duplicate detection
CREATE INDEX IF NOT EXISTS idx_urls_long_url ON urls(long_url);

-- Create index on expires_at for cleanup queries
CREATE INDEX IF NOT EXISTS idx_urls_expires_at ON urls(expires_at) WHERE expires_at IS NOT NULL;

-- Create index on created_at for analytics
CREATE INDEX IF NOT EXISTS idx_urls_created_at ON urls(created_at);
