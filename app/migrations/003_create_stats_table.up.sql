-- Create stats table
CREATE TABLE IF NOT EXISTS stats (
    id BIGSERIAL PRIMARY KEY,
    banner_id BIGINT NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_stats_banner_id FOREIGN KEY (banner_id) REFERENCES banners(id) ON DELETE CASCADE,
    CONSTRAINT uq_stats_banner_timestamp UNIQUE (banner_id, timestamp)
);

-- Create indexes for performance
-- Primary composite index for queries by banner and time range
CREATE INDEX IF NOT EXISTS idx_stats_banner_timestamp ON stats(banner_id, timestamp);

-- Index for timestamp range queries across all banners
CREATE INDEX IF NOT EXISTS idx_stats_timestamp ON stats(timestamp);

-- Index for banner_id queries
CREATE INDEX IF NOT EXISTS idx_stats_banner_id ON stats(banner_id);

-- Index for aggregation queries (sum, count)
CREATE INDEX IF NOT EXISTS idx_stats_count ON stats(count) WHERE count > 0;

-- Note: Partial indexes with NOW() function removed due to PostgreSQL IMMUTABLE requirement
-- These can be created manually if needed with specific timestamps

-- Add comments
COMMENT ON TABLE stats IS 'Таблица агрегированной статистики кликов по минутам';
COMMENT ON COLUMN stats.id IS 'Уникальный идентификатор записи статистики';
COMMENT ON COLUMN stats.banner_id IS 'Идентификатор баннера';
COMMENT ON COLUMN stats.timestamp IS 'Временная метка (округленная до минуты)';
COMMENT ON COLUMN stats.count IS 'Количество кликов за минуту';
COMMENT ON COLUMN stats.created_at IS 'Время создания записи';
COMMENT ON COLUMN stats.updated_at IS 'Время последнего обновления записи';
COMMENT ON CONSTRAINT uq_stats_banner_timestamp ON stats IS 'Уникальность по баннеру и временной метке'; 