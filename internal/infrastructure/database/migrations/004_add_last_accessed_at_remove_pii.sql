-- Migration: Add last_accessed_at field and remove PII violations
-- This migration adds the missing last_accessed_at field for stats tracking
-- and removes PII fields that violate privacy requirements

-- Add last_accessed_at field to urls table
ALTER TABLE urls ADD COLUMN last_accessed_at TIMESTAMP;

-- Create index for performance on last_accessed_at queries
CREATE INDEX IF NOT EXISTS idx_urls_last_accessed_at ON urls(last_accessed_at);

-- Remove PII violations from analytics table
ALTER TABLE url_analytics DROP COLUMN IF EXISTS ip_address;
ALTER TABLE url_analytics DROP COLUMN IF EXISTS city;

-- Fix data type inconsistency between urls and analytics tables
ALTER TABLE url_analytics ALTER COLUMN short_key TYPE VARCHAR(12);

-- Add comment for tracking
COMMENT ON COLUMN urls.last_accessed_at IS 'Timestamp of last URL access for statistics';
