package banner

import (
	"context"
)

// Repository определяет интерфейс для работы с баннерами
type Repository interface {
	// GetByID возвращает баннер по ID
	GetByID(ctx context.Context, id int64) (*Banner, error)

	// Exists проверяет существование баннера
	Exists(ctx context.Context, id int64) (bool, error)
}

// CacheRepository определяет интерфейс для кэширования баннеров
type CacheRepository interface {
	// Get возвращает баннер из кэша
	Get(ctx context.Context, id int64) (*Banner, error)

	// Set сохраняет баннер в кэш
	Set(ctx context.Context, banner *Banner) error

	// Delete удаляет баннер из кэша
	Delete(ctx context.Context, id int64) error

	// Clear очищает весь кэш баннеров
	Clear(ctx context.Context) error
}
