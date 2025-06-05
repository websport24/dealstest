package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/clickcounter/app/internal/domain/banner"
)

// BannerCache реализует кэширование баннеров в памяти
type BannerCache struct {
	cache  *MemoryCache
	logger *logrus.Logger
	ttl    time.Duration
}

// NewBannerCache создает новый экземпляр кэша баннеров
func NewBannerCache(cache *MemoryCache, logger *logrus.Logger, ttl time.Duration) *BannerCache {
	return &BannerCache{
		cache:  cache,
		logger: logger,
		ttl:    ttl,
	}
}

// Get получает баннер из кэша
func (c *BannerCache) Get(ctx context.Context, id int64) (*banner.Banner, error) {
	key := c.bannerKey(id)

	data, err := c.cache.Get(ctx, key)
	if err != nil {
		if err == ErrKeyNotFound {
			return nil, banner.ErrBannerNotFound
		}
		c.logger.WithError(err).WithField("banner_id", id).Error("Failed to get banner from cache")
		return nil, err
	}

	// Десериализуем JSON
	jsonData, ok := data.(string)
	if !ok {
		c.logger.WithField("banner_id", id).Error("Invalid data type in cache")
		return nil, ErrInvalidType
	}

	var b banner.Banner
	if err := json.Unmarshal([]byte(jsonData), &b); err != nil {
		c.logger.WithError(err).WithField("banner_id", id).Error("Failed to unmarshal banner from cache")
		return nil, err
	}

	return &b, nil
}

// Set сохраняет баннер в кэш
func (c *BannerCache) Set(ctx context.Context, b *banner.Banner) error {
	key := c.bannerKey(b.ID)

	// Сериализуем в JSON
	data, err := json.Marshal(b)
	if err != nil {
		c.logger.WithError(err).WithField("banner_id", b.ID).Error("Failed to marshal banner for cache")
		return err
	}

	if err := c.cache.Set(ctx, key, string(data), c.ttl); err != nil {
		c.logger.WithError(err).WithField("banner_id", b.ID).Error("Failed to set banner in cache")
		return err
	}

	c.logger.WithField("banner_id", b.ID).Debug("Banner cached successfully")
	return nil
}

// Delete удаляет баннер из кэша
func (c *BannerCache) Delete(ctx context.Context, id int64) error {
	key := c.bannerKey(id)

	if err := c.cache.Delete(ctx, key); err != nil {
		c.logger.WithError(err).WithField("banner_id", id).Error("Failed to delete banner from cache")
		return err
	}

	c.logger.WithField("banner_id", id).Debug("Banner deleted from cache")
	return nil
}

// Clear очищает весь кэш баннеров
func (c *BannerCache) Clear(ctx context.Context) error {
	// Получаем все ключи и удаляем только ключи баннеров
	keys, err := c.cache.Keys(ctx, "")
	if err != nil {
		c.logger.WithError(err).Error("Failed to get cache keys")
		return err
	}

	bannerKeyPrefix := "banner:"
	for _, key := range keys {
		if len(key) > len(bannerKeyPrefix) && key[:len(bannerKeyPrefix)] == bannerKeyPrefix {
			if err := c.cache.Delete(ctx, key); err != nil {
				c.logger.WithError(err).WithField("key", key).Error("Failed to delete banner key from cache")
			}
		}
	}

	c.logger.Debug("All banner cache cleared")
	return nil
}

// bannerKey генерирует ключ для баннера
func (c *BannerCache) bannerKey(id int64) string {
	return fmt.Sprintf("banner:%d", id)
}
