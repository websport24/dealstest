package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// Config содержит настройки подключения к PostgreSQL
type Config struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// DB представляет подключение к базе данных
type DB struct {
	Pool   *pgxpool.Pool
	config *Config
	logger *logrus.Logger
}

// NewDB создает новое подключение к PostgreSQL
func NewDB(config *Config, logger *logrus.Logger) (*DB, error) {
	if config == nil {
		return nil, fmt.Errorf("database config is required")
	}

	if logger == nil {
		logger = logrus.New()
	}

	// Формируем DSN
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
		config.SSLMode,
	)

	// Настройки пула соединений
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Оптимизация для высокой нагрузки
	poolConfig.MaxConns = int32(config.MaxOpenConns)
	poolConfig.MinConns = int32(config.MaxIdleConns)
	poolConfig.MaxConnLifetime = config.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	// Настройки соединения для производительности
	poolConfig.ConnConfig.RuntimeParams["application_name"] = "clickcounter"
	poolConfig.ConnConfig.RuntimeParams["timezone"] = "UTC"

	// Создаем пул соединений
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	db := &DB{
		Pool:   pool,
		config: config,
		logger: logger,
	}

	// Проверяем подключение
	if err := db.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"host":      config.Host,
		"port":      config.Port,
		"database":  config.DBName,
		"max_conns": config.MaxOpenConns,
	}).Info("Connected to PostgreSQL")

	return db, nil
}

// Ping проверяет подключение к базе данных
func (db *DB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// Close закрывает подключение к базе данных
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		db.logger.Info("Database connection closed")
	}
}

// GetStats возвращает статистику пула соединений
func (db *DB) GetStats() *pgxpool.Stat {
	return db.Pool.Stat()
}

// WithTx выполняет функцию в транзакции
func (db *DB) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			db.logger.WithError(rbErr).Error("Failed to rollback transaction")
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// LogStats логирует статистику пула соединений
func (db *DB) LogStats() {
	stats := db.GetStats()
	db.logger.WithFields(logrus.Fields{
		"acquired_conns":     stats.AcquiredConns(),
		"constructing_conns": stats.ConstructingConns(),
		"idle_conns":         stats.IdleConns(),
		"max_conns":          stats.MaxConns(),
		"total_conns":        stats.TotalConns(),
		"acquire_count":      stats.AcquireCount(),
		"acquire_duration":   stats.AcquireDuration(),
	}).Info("Database pool statistics")
}

// HealthCheck проверяет здоровье базы данных
func (db *DB) HealthCheck(ctx context.Context) error {
	// Проверяем подключение
	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Проверяем, что можем выполнить простой запрос
	var result int
	err := db.Pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected query result: %d", result)
	}

	// Проверяем статистику пула
	stats := db.GetStats()
	if stats.TotalConns() == 0 {
		return fmt.Errorf("no database connections available")
	}

	return nil
}
