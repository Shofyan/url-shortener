-- Alter short_key column to allow longer keys (up to 12 characters)
ALTER TABLE urls ALTER COLUMN short_key TYPE VARCHAR(12);
