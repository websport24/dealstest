-- Create banners table
CREATE TABLE IF NOT EXISTS banners (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    is_active BOOLEAN NOT NULL DEFAULT true
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_banners_is_active ON banners(is_active);
CREATE INDEX IF NOT EXISTS idx_banners_name ON banners(name);
CREATE INDEX IF NOT EXISTS idx_banners_created_at ON banners(created_at);

-- Add comments
COMMENT ON TABLE banners IS 'Таблица баннеров для системы счетчика кликов';
COMMENT ON COLUMN banners.id IS 'Уникальный идентификатор баннера';
COMMENT ON COLUMN banners.name IS 'Название баннера';
COMMENT ON COLUMN banners.created_at IS 'Время создания записи';
COMMENT ON COLUMN banners.updated_at IS 'Время последнего обновления записи';
COMMENT ON COLUMN banners.is_active IS 'Флаг активности баннера'; 