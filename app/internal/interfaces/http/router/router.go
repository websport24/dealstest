package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/clickcounter/app/internal/interfaces/http/handlers"
	"github.com/clickcounter/app/internal/interfaces/http/middleware"
)

// Router представляет HTTP роутер
type Router struct {
	engine        *gin.Engine
	clickHandler  *handlers.ClickHandler
	statsHandler  *handlers.StatsHandler
	healthHandler *handlers.HealthHandler
	logger        *logrus.Logger
}

// NewRouter создает новый HTTP роутер
func NewRouter(
	clickHandler *handlers.ClickHandler,
	statsHandler *handlers.StatsHandler,
	healthHandler *handlers.HealthHandler,
	logger *logrus.Logger,
) *Router {
	// Настраиваем Gin в зависимости от режима
	if logger.Level == logrus.DebugLevel {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	return &Router{
		engine:        engine,
		clickHandler:  clickHandler,
		statsHandler:  statsHandler,
		healthHandler: healthHandler,
		logger:        logger,
	}
}

// Setup настраивает все маршруты и middleware
func (r *Router) Setup() {
	// Middleware
	r.setupMiddleware()

	// API маршруты
	r.setupAPIRoutes()

	// Health check маршруты
	r.setupHealthRoutes()

	// Swagger документация
	r.setupSwagger()
}

// setupMiddleware настраивает middleware
func (r *Router) setupMiddleware() {
	// Recovery middleware
	r.engine.Use(gin.Recovery())

	// Логирование запросов
	r.engine.Use(middleware.LoggingMiddleware(r.logger))

	// CORS
	r.engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Rate limiting middleware
	r.engine.Use(middleware.RateLimitMiddleware())

	// Request timeout middleware
	r.engine.Use(middleware.TimeoutMiddleware(30 * time.Second))

	// Metrics middleware
	r.engine.Use(middleware.MetricsMiddleware())
}

// setupAPIRoutes настраивает API маршруты
func (r *Router) setupAPIRoutes() {
	// API v1 группа
	v1 := r.engine.Group("/api/v1")
	{
		// Основные endpoints согласно ТЗ
		v1.GET("/counter/:bannerID", r.clickHandler.RegisterClick)
		v1.POST("/stats/:bannerID", r.statsHandler.GetStats)
	}

	// Корневые маршруты (для совместимости с примером из ТЗ)
	r.engine.GET("/counter/:bannerID", r.clickHandler.RegisterClick)
	r.engine.POST("/stats/:bannerID", r.statsHandler.GetStats)
}

// setupHealthRoutes настраивает маршруты для проверки здоровья
func (r *Router) setupHealthRoutes() {
	// Основной health check маршрут
	r.engine.GET("/health", r.healthHandler.HealthCheck)
}

// setupSwagger настраивает Swagger документацию
func (r *Router) setupSwagger() {
	// Swagger UI
	r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Редирект с корня на Swagger (для удобства разработки)
	r.engine.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/swagger/index.html")
	})
}

// GetEngine возвращает Gin engine
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// Start запускает HTTP сервер
func (r *Router) Start(addr string) error {
	r.logger.WithField("address", addr).Info("Starting HTTP server")
	return r.engine.Run(addr)
}
