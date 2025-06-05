package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/clickcounter/app/internal/domain/banner"
	"github.com/clickcounter/app/internal/domain/click"
	"github.com/sirupsen/logrus"
)

// ClickUseCase представляет use case для работы с кликами
type ClickUseCase struct {
	clickService  *click.Service
	bannerService *banner.Service
	logger        *logrus.Logger
}

// NewClickUseCase создает новый экземпляр ClickUseCase
func NewClickUseCase(
	clickService *click.Service,
	bannerService *banner.Service,
	logger *logrus.Logger,
) *ClickUseCase {
	if logger == nil {
		logger = logrus.New()
	}

	return &ClickUseCase{
		clickService:  clickService,
		bannerService: bannerService,
		logger:        logger,
	}
}

// RegisterClickRequest представляет запрос на регистрацию клика
type RegisterClickRequest struct {
	BannerID  int64  `json:"banner_id" validate:"required,min=1"`
	UserIP    string `json:"user_ip,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// RegisterClickResponse представляет ответ на регистрацию клика
type RegisterClickResponse struct {
	Success   bool      `json:"success"`
	ClickID   int64     `json:"click_id,omitempty"`
	BannerID  int64     `json:"banner_id"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message,omitempty"`
}

// RegisterClick регистрирует новый клик по баннеру
func (uc *ClickUseCase) RegisterClick(ctx context.Context, req *RegisterClickRequest) (*RegisterClickResponse, error) {
	// Валидация запроса
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	if req.BannerID <= 0 {
		return &RegisterClickResponse{
			Success:  false,
			BannerID: req.BannerID,
			Message:  "Invalid banner ID",
		}, fmt.Errorf("invalid banner ID: %d", req.BannerID)
	}

	// Проверяем существование баннера
	exists, err := uc.bannerService.Exists(ctx, req.BannerID)
	if err != nil {
		uc.logger.WithError(err).WithField("banner_id", req.BannerID).Error("Failed to check banner existence")
		return &RegisterClickResponse{
			Success:  false,
			BannerID: req.BannerID,
			Message:  "Failed to verify banner",
		}, fmt.Errorf("failed to check banner existence: %w", err)
	}

	if !exists {
		uc.logger.WithField("banner_id", req.BannerID).Warn("Attempt to click non-existent banner")
		return &RegisterClickResponse{
			Success:  false,
			BannerID: req.BannerID,
			Message:  "Banner not found",
		}, fmt.Errorf("banner not found: %d", req.BannerID)
	}

	// Регистрируем клик через сервис
	clickEntity, err := uc.clickService.RegisterClick(ctx, req.BannerID, req.UserIP, req.UserAgent)
	if err != nil {
		uc.logger.WithError(err).WithFields(logrus.Fields{
			"banner_id": req.BannerID,
			"user_ip":   req.UserIP,
		}).Error("Failed to register click")
		return &RegisterClickResponse{
			Success:  false,
			BannerID: req.BannerID,
			Message:  "Failed to register click",
		}, fmt.Errorf("failed to register click: %w", err)
	}

	uc.logger.WithFields(logrus.Fields{
		"banner_id": req.BannerID,
		"click_id":  clickEntity.ID,
		"user_ip":   req.UserIP,
	}).Info("Click registered successfully")

	return &RegisterClickResponse{
		Success:   true,
		ClickID:   clickEntity.ID,
		BannerID:  req.BannerID,
		Timestamp: clickEntity.Timestamp,
		Message:   "Click registered successfully",
	}, nil
}

// FlushPendingClicks принудительно сбрасывает все накопленные клики в БД
func (uc *ClickUseCase) FlushPendingClicks(ctx context.Context) error {
	if err := uc.clickService.FlushPendingClicks(ctx); err != nil {
		uc.logger.WithError(err).Error("Failed to flush pending clicks")
		return fmt.Errorf("failed to flush pending clicks: %w", err)
	}

	uc.logger.Info("Pending clicks flushed successfully")
	return nil
}
