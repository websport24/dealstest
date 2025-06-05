package banner

import (
	"context"
	"fmt"
)

// Service представляет доменный сервис для работы с баннерами
type Service struct {
	repo      Repository
	cacheRepo CacheRepository
}

// NewService создает новый экземпляр сервиса баннеров
func NewService(repo Repository, cacheRepo CacheRepository) *Service {
	return &Service{
		repo:      repo,
		cacheRepo: cacheRepo,
	}
}

// Exists проверяет существование активного баннера
func (s *Service) Exists(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, ErrInvalidBannerID
	}

	// Пытаемся получить из кэша
	if s.cacheRepo != nil {
		if banner, err := s.cacheRepo.Get(ctx, id); err == nil && banner != nil {
			return banner.IsActive, nil
		}
	}

	// Получаем из основного репозитория
	banner, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == ErrBannerNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to check banner existence: %w", err)
	}

	// Сохраняем в кэш
	if s.cacheRepo != nil {
		_ = s.cacheRepo.Set(ctx, banner)
	}

	return banner.IsActive, nil
}
