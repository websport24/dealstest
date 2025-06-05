package click

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Service представляет доменный сервис для работы с кликами с batch processing
type Service struct {
	repo Repository

	// Batch processing для highload
	batchMutex    sync.Mutex
	batchQueue    []*Click
	batchSize     int
	flushInterval time.Duration
	flushTimer    *time.Timer
}

// NewService создает новый экземпляр сервиса кликов
func NewService(repo Repository, batchSize int, flushInterval time.Duration) *Service {
	// Устанавливаем разумные значения по умолчанию если переданы некорректные
	if batchSize <= 0 {
		batchSize = 100
	}
	if flushInterval <= 0 {
		flushInterval = 2 * time.Second
	}

	return &Service{
		repo:          repo,
		batchQueue:    make([]*Click, 0),
		batchSize:     batchSize,
		flushInterval: flushInterval,
	}
}

// RegisterClick регистрирует новый клик с batch processing
func (s *Service) RegisterClick(ctx context.Context, bannerID int64, userIP, userAgent string) (*Click, error) {
	click, err := NewClickWithMetadata(bannerID, userIP, userAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to create click: %w", err)
	}

	// Добавляем в batch queue для highload оптимизации
	s.batchMutex.Lock()
	defer s.batchMutex.Unlock()

	s.batchQueue = append(s.batchQueue, click)

	// Если батч заполнен - сбрасываем немедленно
	if len(s.batchQueue) >= s.batchSize {
		if err := s.flushBatchUnsafe(ctx); err != nil {
			return nil, fmt.Errorf("failed to flush batch: %w", err)
		}
	} else {
		// Устанавливаем таймер для принудительного сброса
		s.resetFlushTimer(ctx)
	}

	return click, nil
}

// flushBatchUnsafe сбрасывает накопленные клики в БД (без мьютекса)
func (s *Service) flushBatchUnsafe(ctx context.Context) error {
	if len(s.batchQueue) == 0 {
		return nil
	}

	// Копируем батч для сброса
	batchToFlush := make([]*Click, len(s.batchQueue))
	copy(batchToFlush, s.batchQueue)

	// Очищаем очередь
	s.batchQueue = s.batchQueue[:0]

	// Останавливаем таймер
	if s.flushTimer != nil {
		s.flushTimer.Stop()
		s.flushTimer = nil
	}

	// Сбрасываем в БД
	if err := s.repo.CreateBatch(ctx, batchToFlush); err != nil {
		return fmt.Errorf("failed to save batch clicks: %w", err)
	}

	return nil
}

// resetFlushTimer устанавливает/сбрасывает таймер для принудительного сброса
func (s *Service) resetFlushTimer(ctx context.Context) {
	if s.flushTimer != nil {
		s.flushTimer.Stop()
	}

	s.flushTimer = time.AfterFunc(s.flushInterval, func() {
		s.batchMutex.Lock()
		defer s.batchMutex.Unlock()
		_ = s.flushBatchUnsafe(ctx)
	})
}

// FlushPendingClicks принудительно сбрасывает все накопленные клики
func (s *Service) FlushPendingClicks(ctx context.Context) error {
	s.batchMutex.Lock()
	defer s.batchMutex.Unlock()
	return s.flushBatchUnsafe(ctx)
}
