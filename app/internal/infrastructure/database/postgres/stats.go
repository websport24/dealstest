package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/clickcounter/app/internal/domain/stats"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

// StatsRepository реализует интерфейс stats.Repository для PostgreSQL
type StatsRepository struct {
	db     *DB
	logger *logrus.Logger
}

// StatsAggregationRepository реализует интерфейс stats.AggregationRepository для PostgreSQL
type StatsAggregationRepository struct {
	db     *DB
	logger *logrus.Logger
}

// NewStatsRepository создает новый экземпляр репозитория статистики
func NewStatsRepository(db *DB, logger *logrus.Logger) *StatsRepository {
	if logger == nil {
		logger = logrus.New()
	}

	return &StatsRepository{
		db:     db,
		logger: logger,
	}
}

// NewStatsAggregationRepository создает новый экземпляр репозитория агрегации
func NewStatsAggregationRepository(db *DB, logger *logrus.Logger) *StatsAggregationRepository {
	if logger == nil {
		logger = logrus.New()
	}

	return &StatsAggregationRepository{
		db:     db,
		logger: logger,
	}
}

// GetByBannerIDAndPeriod возвращает статистику по баннеру за период
func (r *StatsRepository) GetByBannerIDAndPeriod(ctx context.Context, bannerID int64, from, to time.Time) ([]*stats.Stat, error) {
	query := `
		SELECT id, banner_id, timestamp, count, created_at, updated_at
		FROM stats
		WHERE banner_id = $1 AND timestamp >= $2 AND timestamp <= $3
		ORDER BY timestamp ASC
	`

	rows, err := r.db.Pool.Query(ctx, query, bannerID, from, to)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"banner_id": bannerID,
			"from":      from,
			"to":        to,
		}).Error("Failed to get stats by banner ID and period")
		return nil, fmt.Errorf("failed to get stats by banner ID and period: %w", err)
	}
	defer rows.Close()

	var statsList []*stats.Stat
	for rows.Next() {
		var s stats.Stat
		err := rows.Scan(
			&s.ID,
			&s.BannerID,
			&s.Timestamp,
			&s.Count,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan stats row")
			return nil, fmt.Errorf("failed to scan stats row: %w", err)
		}
		statsList = append(statsList, &s)
	}

	if err := rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating stats rows")
		return nil, fmt.Errorf("error iterating stats rows: %w", err)
	}

	return statsList, nil
}

// GetAggregatedStats возвращает агрегированную статистику по минутам
func (r *StatsRepository) GetAggregatedStats(ctx context.Context, bannerID int64, from, to time.Time) ([]*stats.MinuteStat, error) {
	query := `
		SELECT timestamp, count
		FROM stats
		WHERE banner_id = $1 AND timestamp >= $2 AND timestamp <= $3
		ORDER BY timestamp ASC
	`

	rows, err := r.db.Pool.Query(ctx, query, bannerID, from, to)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"banner_id": bannerID,
			"from":      from,
			"to":        to,
		}).Error("Failed to get aggregated stats")
		return nil, fmt.Errorf("failed to get aggregated stats: %w", err)
	}
	defer rows.Close()

	var minuteStats []*stats.MinuteStat
	for rows.Next() {
		var timestamp time.Time
		var count int64
		err := rows.Scan(&timestamp, &count)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan aggregated stats row")
			return nil, fmt.Errorf("failed to scan aggregated stats row: %w", err)
		}
		minuteStats = append(minuteStats, stats.NewMinuteStat(timestamp, count))
	}

	if err := rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating aggregated stats rows")
		return nil, fmt.Errorf("error iterating aggregated stats rows: %w", err)
	}

	return minuteStats, nil
}

// CreateOrUpdate создает новую статистику или обновляет существующую
func (r *StatsRepository) CreateOrUpdate(ctx context.Context, s *stats.Stat) error {
	query := `
		INSERT INTO stats (banner_id, timestamp, count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (banner_id, timestamp)
		DO UPDATE SET
			count = stats.count + EXCLUDED.count,
			updated_at = EXCLUDED.updated_at
		RETURNING id
	`

	err := r.db.Pool.QueryRow(ctx, query,
		s.BannerID,
		s.Timestamp,
		s.Count,
		s.CreatedAt,
		s.UpdatedAt,
	).Scan(&s.ID)

	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"banner_id": s.BannerID,
			"timestamp": s.Timestamp,
			"count":     s.Count,
		}).Error("Failed to create or update stats")
		return fmt.Errorf("failed to create or update stats: %w", err)
	}

	return nil
}

// IncrementCount увеличивает счетчик для определенной минуты
func (r *StatsRepository) IncrementCount(ctx context.Context, bannerID int64, timestamp time.Time, delta int64) error {
	// Округляем до минуты
	minuteTimestamp := timestamp.Truncate(time.Minute)

	query := `
		INSERT INTO stats (banner_id, timestamp, count, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (banner_id, timestamp)
		DO UPDATE SET
			count = stats.count + $3,
			updated_at = NOW()
	`

	_, err := r.db.Pool.Exec(ctx, query, bannerID, minuteTimestamp, delta)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"banner_id": bannerID,
			"timestamp": minuteTimestamp,
			"delta":     delta,
		}).Error("Failed to increment stats count")
		return fmt.Errorf("failed to increment stats count: %w", err)
	}

	return nil
}

// GetTotalCount возвращает общее количество кликов за период
func (r *StatsRepository) GetTotalCount(ctx context.Context, bannerID int64, from, to time.Time) (int64, error) {
	query := `
		SELECT COALESCE(SUM(count), 0)
		FROM stats
		WHERE banner_id = $1 AND timestamp >= $2 AND timestamp <= $3
	`

	var total int64
	err := r.db.Pool.QueryRow(ctx, query, bannerID, from, to).Scan(&total)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"banner_id": bannerID,
			"from":      from,
			"to":        to,
		}).Error("Failed to get total count")
		return 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return total, nil
}

// GetStatByMinute возвращает статистику за конкретную минуту
func (r *StatsRepository) GetStatByMinute(ctx context.Context, bannerID int64, timestamp time.Time) (*stats.Stat, error) {
	minuteTimestamp := timestamp.Truncate(time.Minute)

	query := `
		SELECT id, banner_id, timestamp, count, created_at, updated_at
		FROM stats
		WHERE banner_id = $1 AND timestamp = $2
	`

	var s stats.Stat
	err := r.db.Pool.QueryRow(ctx, query, bannerID, minuteTimestamp).Scan(
		&s.ID,
		&s.BannerID,
		&s.Timestamp,
		&s.Count,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, stats.ErrStatNotFound
		}
		r.logger.WithError(err).WithFields(logrus.Fields{
			"banner_id": bannerID,
			"timestamp": minuteTimestamp,
		}).Error("Failed to get stat by minute")
		return nil, fmt.Errorf("failed to get stat by minute: %w", err)
	}

	return &s, nil
}

// BatchCreateOrUpdate создает или обновляет множество записей статистики
func (r *StatsRepository) BatchCreateOrUpdate(ctx context.Context, statsList []*stats.Stat) error {
	if len(statsList) == 0 {
		return nil
	}

	return r.db.WithTx(ctx, func(tx pgx.Tx) error {
		query := `
			INSERT INTO stats (banner_id, timestamp, count, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (banner_id, timestamp)
			DO UPDATE SET
				count = stats.count + EXCLUDED.count,
				updated_at = EXCLUDED.updated_at
		`

		batch := &pgx.Batch{}
		for _, s := range statsList {
			batch.Queue(query, s.BannerID, s.Timestamp, s.Count, s.CreatedAt, s.UpdatedAt)
		}

		results := tx.SendBatch(ctx, batch)
		defer results.Close()

		for i := 0; i < len(statsList); i++ {
			_, err := results.Exec()
			if err != nil {
				r.logger.WithError(err).WithField("batch_index", i).Error("Failed to execute batch stats insert")
				return fmt.Errorf("failed to execute batch stats insert at index %d: %w", i, err)
			}
		}

		r.logger.WithField("count", len(statsList)).Info("Batch stats created/updated successfully")
		return nil
	})
}

// DeleteOldStats удаляет старую статистику
func (r *StatsRepository) DeleteOldStats(ctx context.Context, before time.Time) (int64, error) {
	query := `
		DELETE FROM stats
		WHERE timestamp < $1
	`

	result, err := r.db.Pool.Exec(ctx, query, before)
	if err != nil {
		r.logger.WithError(err).WithField("before", before).Error("Failed to delete old stats")
		return 0, fmt.Errorf("failed to delete old stats: %w", err)
	}

	deleted := result.RowsAffected()
	r.logger.WithFields(logrus.Fields{
		"deleted": deleted,
		"before":  before,
	}).Info("Old stats deleted successfully")

	return deleted, nil
}

// Aggregation Repository Methods

// AggregateClicksToStats агрегирует клики в статистику по минутам
func (r *StatsAggregationRepository) AggregateClicksToStats(ctx context.Context, from, to time.Time) error {
	query := `
		INSERT INTO stats (banner_id, timestamp, count, created_at, updated_at)
		SELECT 
			banner_id,
			date_trunc('minute', timestamp) as minute_timestamp,
			COUNT(*) as click_count,
			NOW() as created_at,
			NOW() as updated_at
		FROM clicks
		WHERE timestamp >= $1 AND timestamp <= $2
		GROUP BY banner_id, date_trunc('minute', timestamp)
		ON CONFLICT (banner_id, timestamp)
		DO UPDATE SET
			count = stats.count + EXCLUDED.count,
			updated_at = EXCLUDED.updated_at
	`

	result, err := r.db.Pool.Exec(ctx, query, from, to)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"from": from,
			"to":   to,
		}).Error("Failed to aggregate clicks to stats")
		return fmt.Errorf("failed to aggregate clicks to stats: %w", err)
	}

	rowsAffected := result.RowsAffected()
	r.logger.WithFields(logrus.Fields{
		"rows_affected": rowsAffected,
		"from":          from,
		"to":            to,
	}).Info("Clicks aggregated to stats successfully")

	return nil
}

// AggregateClicksForBanner агрегирует клики для конкретного баннера
func (r *StatsAggregationRepository) AggregateClicksForBanner(ctx context.Context, bannerID int64, from, to time.Time) error {
	query := `
		INSERT INTO stats (banner_id, timestamp, count, created_at, updated_at)
		SELECT 
			banner_id,
			date_trunc('minute', timestamp) as minute_timestamp,
			COUNT(*) as click_count,
			NOW() as created_at,
			NOW() as updated_at
		FROM clicks
		WHERE banner_id = $1 AND timestamp >= $2 AND timestamp <= $3
		GROUP BY banner_id, date_trunc('minute', timestamp)
		ON CONFLICT (banner_id, timestamp)
		DO UPDATE SET
			count = stats.count + EXCLUDED.count,
			updated_at = EXCLUDED.updated_at
	`

	result, err := r.db.Pool.Exec(ctx, query, bannerID, from, to)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"banner_id": bannerID,
			"from":      from,
			"to":        to,
		}).Error("Failed to aggregate clicks for banner")
		return fmt.Errorf("failed to aggregate clicks for banner: %w", err)
	}

	rowsAffected := result.RowsAffected()
	r.logger.WithFields(logrus.Fields{
		"banner_id":     bannerID,
		"rows_affected": rowsAffected,
		"from":          from,
		"to":            to,
	}).Info("Clicks aggregated for banner successfully")

	return nil
}

// GetClickCountsByMinute возвращает количество кликов, сгруппированных по минутам
func (r *StatsAggregationRepository) GetClickCountsByMinute(ctx context.Context, bannerID int64, from, to time.Time) (map[time.Time]int64, error) {
	query := `
		SELECT 
			date_trunc('minute', timestamp) as minute_timestamp,
			COUNT(*) as click_count
		FROM clicks
		WHERE banner_id = $1 AND timestamp >= $2 AND timestamp <= $3
		GROUP BY date_trunc('minute', timestamp)
		ORDER BY minute_timestamp ASC
	`

	rows, err := r.db.Pool.Query(ctx, query, bannerID, from, to)
	if err != nil {
		r.logger.WithError(err).WithFields(logrus.Fields{
			"banner_id": bannerID,
			"from":      from,
			"to":        to,
		}).Error("Failed to get click counts by minute")
		return nil, fmt.Errorf("failed to get click counts by minute: %w", err)
	}
	defer rows.Close()

	clickCounts := make(map[time.Time]int64)
	for rows.Next() {
		var timestamp time.Time
		var count int64
		err := rows.Scan(&timestamp, &count)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan click counts row")
			return nil, fmt.Errorf("failed to scan click counts row: %w", err)
		}
		clickCounts[timestamp] = count
	}

	if err := rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating click counts rows")
		return nil, fmt.Errorf("error iterating click counts rows: %w", err)
	}

	return clickCounts, nil
}
