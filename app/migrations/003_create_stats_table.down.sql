-- Drop indexes
DROP INDEX IF EXISTS idx_stats_timestamp_cleanup;
DROP INDEX IF EXISTS idx_stats_recent;
DROP INDEX IF EXISTS idx_stats_count;
DROP INDEX IF EXISTS idx_stats_banner_id;
DROP INDEX IF EXISTS idx_stats_timestamp;
DROP INDEX IF EXISTS idx_stats_banner_timestamp;

-- Drop stats table
DROP TABLE IF EXISTS stats; 