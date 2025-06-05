package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config представляет конфигурацию приложения
type Config struct {
	Environment     string                `mapstructure:"environment"`
	Server          ServerConfig          `mapstructure:"server"`
	Database        DatabaseConfig        `mapstructure:"database"`
	Cache           CacheConfig           `mapstructure:"cache"`
	Logger          LoggerConfig          `mapstructure:"logger"`
	ClickFlusher    ClickFlusherConfig    `mapstructure:"click_flusher"`
	StatsAggregator StatsAggregatorConfig `mapstructure:"stats_aggregator"`
}

// ServerConfig конфигурация HTTP сервера
type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
}

// DatabaseConfig конфигурация базы данных
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	SSLMode         string `mapstructure:"ssl_mode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// CacheConfig конфигурация кэша
type CacheConfig struct {
	Type   string       `mapstructure:"type" yaml:"type"` // только "memory"
	Memory MemoryConfig `mapstructure:"memory" yaml:"memory"`
}

// MemoryConfig конфигурация кэша в памяти
type MemoryConfig struct {
	CleanupInterval int `mapstructure:"cleanup_interval" yaml:"cleanup_interval"` // в секундах
	BannerTTL       int `mapstructure:"banner_ttl" yaml:"banner_ttl"`             // в секундах
	StatsTTL        int `mapstructure:"stats_ttl" yaml:"stats_ttl"`               // в секундах
}

// LoggerConfig конфигурация логгера
type LoggerConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// ClickFlusherConfig конфигурация сброса кликов
type ClickFlusherConfig struct {
	Interval  int `mapstructure:"interval"`
	BatchSize int `mapstructure:"batch_size"`
}

// StatsAggregatorConfig конфигурация агрегации статистики
type StatsAggregatorConfig struct {
	Interval int `mapstructure:"interval"`
}

// Load загружает конфигурацию из файла и переменных окружения
func Load() (*Config, error) {
	// Настройка переменных окружения
	viper.SetEnvPrefix("CLICKCOUNTER")
	viper.AutomaticEnv()

	// Значения по умолчанию
	setDefaults()

	// Проверяем переменную окружения для пути к конфигурации
	configPath := os.Getenv("CLICKCOUNTER_CONFIG_PATH")
	if configPath != "" {
		// Используем указанный путь к конфигурации
		viper.SetConfigFile(configPath)
	} else {
		// Используем стандартный поиск конфигурации
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		// Добавляем пути поиска конфигурации
		viper.AddConfigPath("./configs")
		viper.AddConfigPath("../configs")
		viper.AddConfigPath("../../configs")

		// Получаем рабочую директорию
		if wd, err := os.Getwd(); err == nil {
			viper.AddConfigPath(filepath.Join(wd, "configs"))
			viper.AddConfigPath(filepath.Join(filepath.Dir(wd), "configs"))
		}
	}

	// Читаем конфигурационный файл
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Конфигурационный файл не найден, используем значения по умолчанию
			fmt.Println("Config file not found, using defaults and environment variables")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Парсим конфигурацию
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Валидируем конфигурацию
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults устанавливает значения по умолчанию
func setDefaults() {
	// Общие настройки
	viper.SetDefault("environment", "development")

	// Сервер
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)
	viper.SetDefault("server.idle_timeout", 120)

	// База данных
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "clickcounter")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.database", "clickcounter")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.conn_max_lifetime", 3600)

	// Кэш
	viper.SetDefault("cache.type", "memory")

	// Memory кэш
	viper.SetDefault("cache.memory.cleanup_interval", 300) // 5 минут
	viper.SetDefault("cache.memory.banner_ttl", 3600)      // 1 час
	viper.SetDefault("cache.memory.stats_ttl", 900)        // 15 минут

	// Логгер
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "json")

	// Сброс кликов
	viper.SetDefault("click_flusher.interval", 5)
	viper.SetDefault("click_flusher.batch_size", 1000)

	// Агрегация статистики
	viper.SetDefault("stats_aggregator.interval", 60)
}

// validateConfig валидирует конфигурацию
func validateConfig(config *Config) error {
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if config.Database.Port <= 0 || config.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", config.Database.Port)
	}

	// Валидация кэша
	if config.Cache.Type != "memory" {
		return fmt.Errorf("invalid cache type: %s (must be 'memory')", config.Cache.Type)
	}

	// Валидация memory кэша
	if config.Cache.Type == "memory" {
		if config.Cache.Memory.CleanupInterval <= 0 {
			return fmt.Errorf("memory cache cleanup interval must be positive")
		}

		if config.Cache.Memory.BannerTTL <= 0 {
			return fmt.Errorf("memory cache banner TTL must be positive")
		}

		if config.Cache.Memory.StatsTTL <= 0 {
			return fmt.Errorf("memory cache stats TTL must be positive")
		}
	}

	if config.ClickFlusher.Interval <= 0 {
		return fmt.Errorf("click flusher interval must be positive")
	}

	if config.StatsAggregator.Interval <= 0 {
		return fmt.Errorf("stats aggregator interval must be positive")
	}

	return nil
}

// GetDSN возвращает строку подключения к базе данных
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}
