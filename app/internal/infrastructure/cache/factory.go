package cache

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/clickcounter/app/internal/domain/banner"
	"github.com/clickcounter/app/internal/domain/stats"
	"github.com/clickcounter/app/internal/infrastructure/cache/memory"
	"github.com/clickcounter/app/internal/infrastructure/config"
)

// CacheFactory фабрика для создания кэшей
type CacheFactory struct {
	config *config.Config
	logger *logrus.Logger
}

// NewCacheFactory создает новую фабрику кэшей
func NewCacheFactory(config *config.Config, logger *logrus.Logger) *CacheFactory {
	return &CacheFactory{
		config: config,
		logger: logger,
	}
}

// CreateBannerCache создает кэш баннеров в памяти
func (f *CacheFactory) CreateBannerCache() (banner.CacheRepository, error) {
	return f.createMemoryBannerCache()
}

// CreateStatsCache создает кэш статистики в памяти
func (f *CacheFactory) CreateStatsCache() (stats.CacheRepository, error) {
	return f.createMemoryStatsCache()
}

// createMemoryBannerCache создает кэш баннеров в памяти
func (f *CacheFactory) createMemoryBannerCache() (banner.CacheRepository, error) {
	cleanupInterval := time.Duration(f.config.Cache.Memory.CleanupInterval) * time.Second
	bannerTTL := time.Duration(f.config.Cache.Memory.BannerTTL) * time.Second

	memCache := memory.NewMemoryCache(cleanupInterval)
	bannerCache := memory.NewBannerCache(memCache, f.logger, bannerTTL)

	f.logger.WithFields(logrus.Fields{
		"cache_type":       "memory",
		"cleanup_interval": cleanupInterval,
		"banner_ttl":       bannerTTL,
	}).Info("Memory banner cache created")

	return bannerCache, nil
}

// createMemoryStatsCache создает кэш статистики в памяти
func (f *CacheFactory) createMemoryStatsCache() (stats.CacheRepository, error) {
	cleanupInterval := time.Duration(f.config.Cache.Memory.CleanupInterval) * time.Second
	statsTTL := time.Duration(f.config.Cache.Memory.StatsTTL) * time.Second

	memCache := memory.NewMemoryCache(cleanupInterval)
	statsCache := memory.NewStatsCache(memCache, f.logger, statsTTL)

	f.logger.WithFields(logrus.Fields{
		"cache_type":       "memory",
		"cleanup_interval": cleanupInterval,
		"stats_ttl":        statsTTL,
	}).Info("Memory stats cache created")

	return statsCache, nil
}

// GetCacheType возвращает тип используемого кэша
func (f *CacheFactory) GetCacheType() string {
	return "memory"
}

// IsMemoryCache проверяет, используется ли кэш в памяти
func (f *CacheFactory) IsMemoryCache() bool {
	return true
}
