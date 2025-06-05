package click

import (
	"context"
)

// Repository определяет интерфейс для работы с кликами
type Repository interface {
	// Create создает новый клик
	Create(ctx context.Context, click *Click) error

	// CreateBatch создает множество кликов за одну операцию (для оптимизации)
	CreateBatch(ctx context.Context, clicks []*Click) error
}
