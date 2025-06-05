package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/clickcounter/app/internal/domain/banner"
	"github.com/clickcounter/app/internal/domain/stats"
	"github.com/sirupsen/logrus"
)

// StatsUseCase представляет use case для работы со статистикой
type StatsUseCase struct {
	statsService  *stats.Service
	bannerService *banner.Service
	logger        *logrus.Logger
}

// NewStatsUseCase создает новый экземпляр StatsUseCase
func NewStatsUseCase(
	statsService *stats.Service,
	bannerService *banner.Service,
	logger *logrus.Logger,
) *StatsUseCase {
	if logger == nil {
		logger = logrus.New()
	}

	return &StatsUseCase{
		statsService:  statsService,
		bannerService: bannerService,
		logger:        logger,
	}
}

// GetStatsRequest представляет запрос на получение статистики
type GetStatsRequest struct {
	BannerID int64     `json:"banner_id" validate:"required,min=1"`
	From     time.Time `json:"from" validate:"required"`
	To       time.Time `json:"to" validate:"required"`
}

// GetStatsResponse представляет ответ со статистикой (согласно ТЗ)
type GetStatsResponse struct {
	Stats []*MinuteStatDTO `json:"stats"`
}

// MinuteStatDTO представляет статистику за минуту (формат ТЗ)
type MinuteStatDTO struct {
	Timestamp time.Time `json:"ts"`
	Count     int64     `json:"v"`
}

// GetStats возвращает статистику кликов по баннеру за период
func (uc *StatsUseCase) GetStats(ctx context.Context, req *GetStatsRequest) (*GetStatsResponse, error) {
	// Валидация запроса
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	if req.BannerID <= 0 {
		return nil, fmt.Errorf("invalid banner ID: %d", req.BannerID)
	}

	if req.From.After(req.To) {
		return nil, fmt.Errorf("from time cannot be after to time")
	}

	// Ограничиваем максимальный период запроса (30 дней)
	maxPeriod := 30 * 24 * time.Hour
	if req.To.Sub(req.From) > maxPeriod {
		return nil, fmt.Errorf("period too large, maximum allowed: %v", maxPeriod)
	}

	// Проверяем существование баннера
	exists, err := uc.bannerService.Exists(ctx, req.BannerID)
	if err != nil {
		uc.logger.WithError(err).WithField("banner_id", req.BannerID).Error("Failed to check banner existence")
		return nil, fmt.Errorf("failed to check banner existence: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("banner not found: %d", req.BannerID)
	}

	// Получаем статистику через сервис
	statsResponse, err := uc.statsService.GetStatsWithFallback(ctx, req.BannerID, req.From, req.To)
	if err != nil {
		uc.logger.WithError(err).WithFields(logrus.Fields{
			"banner_id": req.BannerID,
			"from":      req.From,
			"to":        req.To,
		}).Error("Failed to get aggregated stats")
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	// Конвертируем в простой формат ТЗ
	statsDTO := make([]*MinuteStatDTO, len(statsResponse.Stats))
	for i, stat := range statsResponse.Stats {
		statsDTO[i] = &MinuteStatDTO{
			Timestamp: stat.Timestamp,
			Count:     stat.Value,
		}
	}

	uc.logger.WithFields(logrus.Fields{
		"banner_id":   req.BannerID,
		"stats_count": len(statsDTO),
	}).Info("Stats retrieved successfully")

	return &GetStatsResponse{
		Stats: statsDTO,
	}, nil
}
