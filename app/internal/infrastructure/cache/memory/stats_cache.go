package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/clickcounter/app/internal/domain/stats"
)

// StatsCache реализует кэширование статистики в памяти
type StatsCache struct {
	cache  *MemoryCache
	logger *logrus.Logger
	ttl    time.Duration
}

// NewStatsCache создает новый экземпляр кэша статистики
func NewStatsCache(cache *MemoryCache, logger *logrus.Logger, ttl time.Duration) *StatsCache {
	return &StatsCache{
		cache:  cache,
		logger: logger,
		ttl:    ttl,
	}
}

// GetStats получает статистику из кэша
func (c *StatsCache) GetStats(ctx context.Context, bannerID int64, from, to time.Time) (*stats.StatsResponse, error) {
	key := c.statsKey(bannerID, from, to)

	data, err := c.cache.Get(ctx, key)
	if err != nil {
		if err == ErrKeyNotFound {
			return nil, stats.ErrStatNotFound
		}
		c.logger.WithError(err).WithFields(logrus.Fields{
			"banner_id": bannerID,
			"from":      from,
			"to":        to,
		}).Error("Failed to get stats from cache")
		return nil, err
	}

	// Десериализуем JSON
	jsonData, ok := data.(string)
	if !ok {
		c.logger.WithFields(logrus.Fields{
			"banner_id": bannerID,
			"from":      from,
			"to":        to,
		}).Error("Invalid data type in cache")
		return nil, ErrInvalidType
	}

	var statsResp stats.StatsResponse
	if err := json.Unmarshal([]byte(jsonData), &statsResp); err != nil {
		c.logger.WithError(err).WithFields(logrus.Fields{
			"banner_id": bannerID,
			"from":      from,
			"to":        to,
		}).Error("Failed to unmarshal stats from cache")
		return nil, err
	}

	return &statsResp, nil
}

// SetStats сохраняет статистику в кэш
func (c *StatsCache) SetStats(ctx context.Context, bannerID int64, from, to time.Time, statsResp *stats.StatsResponse) error {
	key := c.statsKey(bannerID, from, to)

	// Сериализуем в JSON
	data, err := json.Marshal(statsResp)
	if err != nil {
		c.logger.WithError(err).WithFields(logrus.Fields{
			"banner_id": bannerID,
			"from":      from,
			"to":        to,
		}).Error("Failed to marshal stats for cache")
		return err
	}

	if err := c.cache.Set(ctx, key, string(data), c.ttl); err != nil {
		c.logger.WithError(err).WithFields(logrus.Fields{
			"banner_id": bannerID,
			"from":      from,
			"to":        to,
		}).Error("Failed to set stats in cache")
		return err
	}

	c.logger.WithFields(logrus.Fields{
		"banner_id": bannerID,
		"from":      from,
		"to":        to,
	}).Debug("Stats cached successfully")
	return nil
}

// DeleteStats удаляет статистику из кэша
func (c *StatsCache) DeleteStats(ctx context.Context, bannerID int64, from, to time.Time) error {
	key := c.statsKey(bannerID, from, to)

	if err := c.cache.Delete(ctx, key); err != nil {
		c.logger.WithError(err).WithFields(logrus.Fields{
			"banner_id": bannerID,
			"from":      from,
			"to":        to,
		}).Error("Failed to delete stats from cache")
		return err
	}

	c.logger.WithFields(logrus.Fields{
		"banner_id": bannerID,
		"from":      from,
		"to":        to,
	}).Debug("Stats deleted from cache")
	return nil
}

// statsKey генерирует ключ для статистики
func (c *StatsCache) statsKey(bannerID int64, from, to time.Time) string {
	return fmt.Sprintf("stats:%d:%d:%d", bannerID, from.Unix(), to.Unix())
}
