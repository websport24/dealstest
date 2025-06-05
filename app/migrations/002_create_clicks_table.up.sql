-- Create clicks table
CREATE TABLE IF NOT EXISTS clicks (
    id BIGSERIAL PRIMARY KEY,
    banner_id BIGINT NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    user_ip INET,
    user_agent TEXT,
    CONSTRAINT fk_clicks_banner_id FOREIGN KEY (banner_id) REFERENCES banners(id) ON DELETE CASCADE
);

-- Create indexes for high-performance queries
-- Composite index for banner_id + timestamp (most common query pattern)
CREATE INDEX IF NOT EXISTS idx_clicks_banner_timestamp ON clicks(banner_id, timestamp);

-- Index for timestamp range queries
CREATE INDEX IF NOT EXISTS idx_clicks_timestamp ON clicks(timestamp);

-- Index for banner_id queries
CREATE INDEX IF NOT EXISTS idx_clicks_banner_id ON clicks(banner_id);

-- Note: Partial indexes with NOW() function removed due to PostgreSQL IMMUTABLE requirement
-- These can be created manually if needed with specific timestamps

-- Add comments
COMMENT ON TABLE clicks IS 'Таблица кликов по баннерам';
COMMENT ON COLUMN clicks.id IS 'Уникальный идентификатор клика';
COMMENT ON COLUMN clicks.banner_id IS 'Идентификатор баннера';
COMMENT ON COLUMN clicks.timestamp IS 'Время клика';
COMMENT ON COLUMN clicks.user_ip IS 'IP адрес пользователя';
COMMENT ON COLUMN clicks.user_agent IS 'User-Agent браузера пользователя'; 