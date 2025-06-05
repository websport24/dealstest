package stats

import (
	"context"
	"time"
)

// Repository определяет интерфейс для работы со статистикой
type Repository interface {
	// GetAggregatedStats возвращает агрегированную статистику по минутам
	GetAggregatedStats(ctx context.Context, bannerID int64, from, to time.Time) ([]*MinuteStat, error)

	// IncrementCount увеличивает счетчик для определенной минуты
	IncrementCount(ctx context.Context, bannerID int64, timestamp time.Time, delta int64) error
}

// AggregationRepository определяет интерфейс для агрегации данных из кликов
type AggregationRepository interface {
	// AggregateClicksForBanner агрегирует клики для конкретного баннера
	AggregateClicksForBanner(ctx context.Context, bannerID int64, from, to time.Time) error
}

// CacheRepository определяет интерфейс для кэширования статистики
type CacheRepository interface {
	// GetStats возвращает статистику из кэша
	GetStats(ctx context.Context, bannerID int64, from, to time.Time) (*StatsResponse, error)

	// SetStats сохраняет статистику в кэш
	SetStats(ctx context.Context, bannerID int64, from, to time.Time, stats *StatsResponse) error
}
