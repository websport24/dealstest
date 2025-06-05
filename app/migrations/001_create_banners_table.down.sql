-- Drop indexes
DROP INDEX IF EXISTS idx_banners_created_at;
DROP INDEX IF EXISTS idx_banners_name;
DROP INDEX IF EXISTS idx_banners_is_active;

-- Drop banners table
DROP TABLE IF EXISTS banners; 