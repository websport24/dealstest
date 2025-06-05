package postgres

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/clickcounter/app/internal/domain/click"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

// ClickRepository реализует интерфейс click.Repository для PostgreSQL
type ClickRepository struct {
	db     *DB
	logger *logrus.Logger
}

// ClickBatchRepository реализует интерфейс click.BatchRepository для PostgreSQL
type ClickBatchRepository struct {
	db        *DB
	logger    *logrus.Logger
	batch     []*click.Click
	batchSize int
	mutex     sync.RWMutex
}

// NewClickRepository создает новый экземпляр репозитория кликов
func NewClickRepository(db *DB, logger *logrus.Logger) *ClickRepository {
	if logger == nil {
		logger = logrus.New()
	}

	return &ClickRepository{
		db:     db,
		logger: logger,
	}
}

// NewClickBatchRepository создает новый экземпляр батчевого репозитория кликов
func NewClickBatchRepository(db *DB, logger *logrus.Logger, batchSize int) *ClickBatchRepository {
	if logger == nil {
		logger = logrus.New()
	}

	if batchSize <= 0 {
		batchSize = 1000
	}

	return &ClickBatchRepository{
		db:        db,
		logger:    logger,
		batch:     make([]*click.Click, 0, batchSize),
		batchSize: batchSize,
	}
}

// Create создает новый клик
func (r *ClickRepository) Create(ctx context.Context, c *click.Click) error {
	query := `
		INSERT INTO clicks (banner_id, timestamp, user_ip, user_agent)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := r.db.Pool.QueryRow(ctx, query,
		c.BannerID,
		c.Timestamp,
		c.UserIP,
		c.UserAgent,
	).Scan(&c.ID)

	if err != nil {
		r.logger.WithError(err).WithField("banner_id", c.BannerID).Error("Failed to create click")
		return fmt.Errorf("failed to create click: %w", err)
	}

	return nil
}

// CreateBatch создает множество кликов за одну операцию
func (r *ClickRepository) CreateBatch(ctx context.Context, clicks []*click.Click) error {
	if len(clicks) == 0 {
		return nil
	}

	return r.db.WithTx(ctx, func(tx pgx.Tx) error {
		query := `
			INSERT INTO clicks (banner_id, timestamp, user_ip, user_agent)
			VALUES ($1, $2, $3, $4)
		`

		batch := &pgx.Batch{}
		for _, c := range clicks {
			batch.Queue(query, c.BannerID, c.Timestamp, c.UserIP, c.UserAgent)
		}

		results := tx.SendBatch(ctx, batch)
		defer results.Close()

		for i := 0; i < len(clicks); i++ {
			_, err := results.Exec()
			if err != nil {
				r.logger.WithError(err).WithField("batch_index", i).Error("Failed to execute batch insert")
				return fmt.Errorf("failed to execute batch insert at index %d: %w", i, err)
			}
		}

		r.logger.WithField("count", len(clicks)).Info("Batch clicks created successfully")
		return nil
	})
}

// GetByID возвращает клик по ID
func (r *ClickRepository) GetByID(ctx context.Context, id int64) (*click.Click, error) {
	query := `
		SELECT id, banner_id, timestamp, user_ip, user_agent
		FROM clicks
		WHERE id = $1
	`

	var c click.Click
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&c.ID,
		&c.BannerID,
		&c.Timestamp,
		&c.UserIP,
		&c.UserAgent,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, click.ErrClickNotFound
		}
		r.logger.WithError(err).WithField("click_id", id).Error("Failed to get click by ID")
		return nil, fmt.Errorf("failed to get click by ID: %w", err)
	}

	return &c, nil
}

// GetByBannerID возвращает клики по ID баннера за период
func (r *ClickRepository) GetByBannerID(ctx context.Context, bannerID int64, from, to time.Time) ([]*click.Click, error) {
	query := `
		SELECT id, banner_id, timestamp, user_ip, user_agent
		FROM clicks
		WHERE banner_id = $1 AND timestamp >= $2 AND timestamp <= $3
		ORDER BY timestamp ASC
	`

	rows, err := r.db.Pool.Query(ctx, query, bannerID, from, to)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"banner_id": bannerID,
			"from":      from,
			"to":        to,
		}).Error("Failed to get clicks by banner ID")
		return nil, fmt.Errorf("failed to get clicks by banner ID: %w", err)
	}
	defer rows.Close()

	var clicks []*click.Click
	for rows.Next() {
		var c click.Click
		err := rows.Scan(
			&c.ID,
			&c.BannerID,
			&c.Timestamp,
			&c.UserIP,
			&c.UserAgent,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan click row")
			return nil, fmt.Errorf("failed to scan click row: %w", err)
		}
		clicks = append(clicks, &c)
	}

	if err := rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating click rows")
		return nil, fmt.Errorf("error iterating click rows: %w", err)
	}

	return clicks, nil
}

// GetByBannerIDWithPagination возвращает клики с пагинацией
func (r *ClickRepository) GetByBannerIDWithPagination(ctx context.Context, bannerID int64, from, to time.Time, limit, offset int) ([]*click.Click, error) {
	query := `
		SELECT id, banner_id, timestamp, user_ip, user_agent
		FROM clicks
		WHERE banner_id = $1 AND timestamp >= $2 AND timestamp <= $3
		ORDER BY timestamp ASC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.db.Pool.Query(ctx, query, bannerID, from, to, limit, offset)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"banner_id": bannerID,
			"from":      from,
			"to":        to,
			"limit":     limit,
			"offset":    offset,
		}).Error("Failed to get clicks with pagination")
		return nil, fmt.Errorf("failed to get clicks with pagination: %w", err)
	}
	defer rows.Close()

	var clicks []*click.Click
	for rows.Next() {
		var c click.Click
		err := rows.Scan(
			&c.ID,
			&c.BannerID,
			&c.Timestamp,
			&c.UserIP,
			&c.UserAgent,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan click row")
			return nil, fmt.Errorf("failed to scan click row: %w", err)
		}
		clicks = append(clicks, &c)
	}

	if err := rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating click rows")
		return nil, fmt.Errorf("error iterating click rows: %w", err)
	}

	return clicks, nil
}

// CountByBannerID возвращает количество кликов по баннеру за период
func (r *ClickRepository) CountByBannerID(ctx context.Context, bannerID int64, from, to time.Time) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM clicks
		WHERE banner_id = $1 AND timestamp >= $2 AND timestamp <= $3
	`

	var count int64
	err := r.db.Pool.QueryRow(ctx, query, bannerID, from, to).Scan(&count)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"banner_id": bannerID,
			"from":      from,
			"to":        to,
		}).Error("Failed to count clicks by banner ID")
		return 0, fmt.Errorf("failed to count clicks by banner ID: %w", err)
	}

	return count, nil
}

// GetClicksForPeriod возвращает все клики за период
func (r *ClickRepository) GetClicksForPeriod(ctx context.Context, from, to time.Time) ([]*click.Click, error) {
	query := `
		SELECT id, banner_id, timestamp, user_ip, user_agent
		FROM clicks
		WHERE timestamp >= $1 AND timestamp <= $2
		ORDER BY timestamp ASC
	`

	rows, err := r.db.Pool.Query(ctx, query, from, to)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"from": from,
			"to":   to,
		}).Error("Failed to get clicks for period")
		return nil, fmt.Errorf("failed to get clicks for period: %w", err)
	}
	defer rows.Close()

	var clicks []*click.Click
	for rows.Next() {
		var c click.Click
		err := rows.Scan(
			&c.ID,
			&c.BannerID,
			&c.Timestamp,
			&c.UserIP,
			&c.UserAgent,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan click row")
			return nil, fmt.Errorf("failed to scan click row: %w", err)
		}
		clicks = append(clicks, &c)
	}

	if err := rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating click rows")
		return nil, fmt.Errorf("error iterating click rows: %w", err)
	}

	return clicks, nil
}

// DeleteOldClicks удаляет старые клики
func (r *ClickRepository) DeleteOldClicks(ctx context.Context, before time.Time) (int64, error) {
	query := `
		DELETE FROM clicks
		WHERE timestamp < $1
	`

	result, err := r.db.Pool.Exec(ctx, query, before)
	if err != nil {
		r.logger.WithError(err).WithField("before", before).Error("Failed to delete old clicks")
		return 0, fmt.Errorf("failed to delete old clicks: %w", err)
	}

	deleted := result.RowsAffected()
	r.logger.WithFields(logrus.Fields{
		"deleted": deleted,
		"before":  before,
	}).Info("Old clicks deleted successfully")

	return deleted, nil
}

// Batch Repository Methods

// AddToBatch добавляет клик в батч
func (r *ClickBatchRepository) AddToBatch(ctx context.Context, c *click.Click) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.batch = append(r.batch, c)
	return nil
}

// FlushBatch записывает все накопленные клики в БД
func (r *ClickBatchRepository) FlushBatch(ctx context.Context) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if len(r.batch) == 0 {
		return nil
	}

	// Создаем копию батча для обработки
	batchToProcess := make([]*click.Click, len(r.batch))
	copy(batchToProcess, r.batch)

	// Очищаем батч
	r.batch = r.batch[:0]

	// Обрабатываем батч
	return r.db.WithTx(ctx, func(tx pgx.Tx) error {
		query := `
			INSERT INTO clicks (banner_id, timestamp, user_ip, user_agent)
			VALUES ($1, $2, $3, $4)
		`

		batch := &pgx.Batch{}
		for _, c := range batchToProcess {
			batch.Queue(query, c.BannerID, c.Timestamp, c.UserIP, c.UserAgent)
		}

		results := tx.SendBatch(ctx, batch)
		defer results.Close()

		for i := 0; i < len(batchToProcess); i++ {
			_, err := results.Exec()
			if err != nil {
				r.logger.WithError(err).WithField("batch_index", i).Error("Failed to execute batch insert")
				return fmt.Errorf("failed to execute batch insert at index %d: %w", i, err)
			}
		}

		r.logger.WithField("count", len(batchToProcess)).Info("Batch clicks flushed successfully")
		return nil
	})
}

// GetBatchSize возвращает текущий размер батча
func (r *ClickBatchRepository) GetBatchSize() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.batch)
}

// SetBatchSize устанавливает размер батча
func (r *ClickBatchRepository) SetBatchSize(size int) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.batchSize = size
}
