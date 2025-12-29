-- Create analytics table for detailed URL tracking
CREATE TABLE IF NOT EXISTS url_analytics (
    id BIGSERIAL PRIMARY KEY,
    short_key VARCHAR(10) NOT NULL,
    visited_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ip_address INET,
    user_agent TEXT,
    referer TEXT,
    country VARCHAR(2),
    city VARCHAR(100),
    FOREIGN KEY (short_key) REFERENCES urls(short_key) ON DELETE CASCADE
);

-- Create index on short_key for analytics queries
CREATE INDEX IF NOT EXISTS idx_analytics_short_key ON url_analytics(short_key);

-- Create index on visited_at for time-based queries
CREATE INDEX IF NOT EXISTS idx_analytics_visited_at ON url_analytics(visited_at);

-- Create index on country for geographical analytics
CREATE INDEX IF NOT EXISTS idx_analytics_country ON url_analytics(country) WHERE country IS NOT NULL;
