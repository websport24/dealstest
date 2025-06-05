package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof" // Профилирование
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/clickcounter/app/internal/application/usecase"
	"github.com/clickcounter/app/internal/domain/banner"
	"github.com/clickcounter/app/internal/domain/click"
	"github.com/clickcounter/app/internal/domain/stats"
	"github.com/clickcounter/app/internal/infrastructure/cache"
	"github.com/clickcounter/app/internal/infrastructure/config"
	"github.com/clickcounter/app/internal/infrastructure/database/postgres"
	"github.com/clickcounter/app/internal/interfaces/http/handlers"
	"github.com/clickcounter/app/internal/interfaces/http/router"
	"github.com/clickcounter/app/pkg/logger"
)

// @title Click Counter API
// @version 1.0
// @description Высоконагруженный сервис для подсчета кликов по баннерам
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http https

func main() {
	// Инициализация логгера
	appLogger := logger.NewLogger()
	appLogger.Info("Starting Click Counter service...")

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to load configuration")
	}

	// Настройка уровня логирования
	if level, err := logrus.ParseLevel(cfg.Logger.Level); err == nil {
		appLogger.SetLevel(level)
	}

	appLogger.WithFields(logrus.Fields{
		"environment": cfg.Environment,
		"log_level":   cfg.Logger.Level,
		"port":        cfg.Server.Port,
	}).Info("Configuration loaded successfully")

	// Инициализация базы данных
	dbConfig := &postgres.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.Database,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: time.Duration(cfg.Database.ConnMaxLifetime) * time.Second,
	}

	dbConn, err := postgres.NewDB(dbConfig, appLogger)
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to connect to database")
	}
	defer dbConn.Close()

	appLogger.Info("Database connection established")

	// Инициализация кэшей через фабрику
	cacheFactory := cache.NewCacheFactory(cfg, appLogger)

	bannerCache, err := cacheFactory.CreateBannerCache()
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to create banner cache")
	}

	statsCache, err := cacheFactory.CreateStatsCache()
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to create stats cache")
	}

	appLogger.WithField("cache_type", cacheFactory.GetCacheType()).Info("Cache initialized successfully")

	// Инициализация репозиториев
	bannerRepo := postgres.NewBannerRepository(dbConn, appLogger)
	clickRepo := postgres.NewClickRepository(dbConn, appLogger)
	statsRepo := postgres.NewStatsRepository(dbConn, appLogger)
	statsAggRepo := postgres.NewStatsAggregationRepository(dbConn, appLogger)

	// Инициализация доменных сервисов с кэшами
	bannerService := banner.NewService(bannerRepo, bannerCache)

	// Создаем сервис кликов с параметрами из конфигурации
	clickService := click.NewService(
		clickRepo,
		cfg.ClickFlusher.BatchSize,
		time.Duration(cfg.ClickFlusher.Interval)*time.Second,
	)

	statsService := stats.NewService(statsRepo, statsAggRepo, statsCache)

	// Инициализация use cases
	clickUseCase := usecase.NewClickUseCase(
		clickService,
		bannerService,
		appLogger,
	)

	statsUseCase := usecase.NewStatsUseCase(
		statsService,
		bannerService,
		appLogger,
	)

	// Инициализация handlers
	clickHandler := handlers.NewClickHandler(clickUseCase, appLogger)
	statsHandler := handlers.NewStatsHandler(statsUseCase, appLogger)
	healthHandler := handlers.NewHealthHandler(dbConn, appLogger)

	// Инициализация роутера
	appRouter := router.NewRouter(clickHandler, statsHandler, healthHandler, appLogger)
	appRouter.Setup()

	// Оптимизация Gin для продакшена
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else if cfg.Environment == "development" {
		// Запускаем pprof сервер для профилирования в development режиме
		go func() {
			appLogger.Info("Starting pprof server on :6060")
			if err := http.ListenAndServe("localhost:6060", nil); err != nil {
				appLogger.WithError(err).Error("Failed to start pprof server")
			}
		}()
	}

	// Настройка HTTP сервера с оптимизациями для высокой нагрузки
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      appRouter.GetEngine(),
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,

		// Оптимизации для высокой нагрузки
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Включаем keep-alive соединения
	server.SetKeepAlivesEnabled(true)

	// Канал для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Запуск сервера в горутине
	go func() {
		appLogger.WithField("address", server.Addr).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Ожидание сигнала завершения
	<-quit
	appLogger.Info("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Принудительно сбрасываем накопленные клики перед остановкой
	appLogger.Info("Flushing pending clicks...")
	if err := clickUseCase.FlushPendingClicks(shutdownCtx); err != nil {
		appLogger.WithError(err).Error("Failed to flush pending clicks during shutdown")
	}

	// Останавливаем HTTP сервер
	if err := server.Shutdown(shutdownCtx); err != nil {
		appLogger.WithError(err).Error("Server forced to shutdown")
	}

	appLogger.Info("Server exited")
}
