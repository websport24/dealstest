package postgres

import (
	"context"
	"fmt"

	"github.com/clickcounter/app/internal/domain/banner"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

// BannerRepository реализует интерфейс banner.Repository для PostgreSQL
type BannerRepository struct {
	db     *DB
	logger *logrus.Logger
}

// NewBannerRepository создает новый экземпляр репозитория баннеров
func NewBannerRepository(db *DB, logger *logrus.Logger) *BannerRepository {
	if logger == nil {
		logger = logrus.New()
	}

	return &BannerRepository{
		db:     db,
		logger: logger,
	}
}

// GetByID возвращает баннер по ID
func (r *BannerRepository) GetByID(ctx context.Context, id int64) (*banner.Banner, error) {
	query := `
		SELECT id, name, created_at, updated_at, is_active
		FROM banners
		WHERE id = $1 AND is_active = true
	`

	var b banner.Banner
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&b.ID,
		&b.Name,
		&b.CreatedAt,
		&b.UpdatedAt,
		&b.IsActive,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, banner.ErrBannerNotFound
		}
		r.logger.WithError(err).WithField("banner_id", id).Error("Failed to get banner by ID")
		return nil, fmt.Errorf("failed to get banner by ID: %w", err)
	}

	return &b, nil
}

// Exists проверяет существование баннера
func (r *BannerRepository) Exists(ctx context.Context, id int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM banners
			WHERE id = $1 AND is_active = true
		)
	`

	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		r.logger.WithError(err).WithField("banner_id", id).Error("Failed to check banner existence")
		return false, fmt.Errorf("failed to check banner existence: %w", err)
	}

	return exists, nil
}
