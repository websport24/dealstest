-- Drop indexes
DROP INDEX IF EXISTS idx_clicks_timestamp_cleanup;
DROP INDEX IF EXISTS idx_clicks_recent;
DROP INDEX IF EXISTS idx_clicks_banner_id;
DROP INDEX IF EXISTS idx_clicks_timestamp;
DROP INDEX IF EXISTS idx_clicks_banner_timestamp;

-- Drop clicks table
DROP TABLE IF EXISTS clicks; 