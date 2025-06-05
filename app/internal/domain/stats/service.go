package stats

import (
	"context"
	"fmt"
	"time"
)

// Service представляет доменный сервис для работы со статистикой
type Service struct {
	repo      Repository
	aggRepo   AggregationRepository
	cacheRepo CacheRepository
	cacheTTL  time.Duration
}

// NewService создает новый экземпляр сервиса статистики
func NewService(repo Repository, aggRepo AggregationRepository, cacheRepo CacheRepository) *Service {
	return &Service{
		repo:      repo,
		aggRepo:   aggRepo,
		cacheRepo: cacheRepo,
		cacheTTL:  5 * time.Minute, // TTL кэша по умолчанию
	}
}

// GetStats возвращает статистику кликов по баннеру за период
func (s *Service) GetStats(ctx context.Context, bannerID int64, from, to time.Time) (*StatsResponse, error) {
	if bannerID <= 0 {
		return nil, ErrInvalidBannerID
	}

	period, err := NewStatPeriod(from, to)
	if err != nil {
		return nil, fmt.Errorf("invalid period: %w", err)
	}

	// Пытаемся получить из кэша
	if s.cacheRepo != nil {
		if cached, err := s.cacheRepo.GetStats(ctx, bannerID, from, to); err == nil && cached != nil {
			return cached, nil
		}
	}

	// Получаем агрегированную статистику
	minuteStats, err := s.repo.GetAggregatedStats(ctx, bannerID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get aggregated stats: %w", err)
	}

	// Формируем ответ
	response := &StatsResponse{
		BannerID: bannerID,
		Period:   *period,
		Stats:    minuteStats,
	}

	// Вычисляем общую сумму
	response.CalculateTotal()

	// Сортируем по времени
	response.SortByTimestamp()

	// Сохраняем в кэш
	if s.cacheRepo != nil {
		_ = s.cacheRepo.SetStats(ctx, bannerID, from, to, response)
	}

	return response, nil
}

// GetStatsWithFallback возвращает статистику с fallback на агрегацию кликов
func (s *Service) GetStatsWithFallback(ctx context.Context, bannerID int64, from, to time.Time) (*StatsResponse, error) {
	// Сначала пытаемся получить готовую статистику
	response, err := s.GetStats(ctx, bannerID, from, to)
	if err == nil && len(response.Stats) > 0 {
		return response, nil
	}

	// Если статистики нет, агрегируем из кликов
	if s.aggRepo != nil {
		if err := s.aggRepo.AggregateClicksForBanner(ctx, bannerID, from, to); err != nil {
			return nil, fmt.Errorf("failed to aggregate clicks: %w", err)
		}

		// Повторно получаем статистику
		return s.GetStats(ctx, bannerID, from, to)
	}

	return response, err
}
