package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/clickcounter/app/internal/infrastructure/database/postgres"
	"github.com/clickcounter/app/internal/interfaces/http/dto"
)

// HealthHandler обрабатывает запросы проверки здоровья сервиса
type HealthHandler struct {
	dbConn *postgres.DB
	logger *logrus.Logger
}

// NewHealthHandler создает новый обработчик health check
func NewHealthHandler(dbConn *postgres.DB, logger *logrus.Logger) *HealthHandler {
	return &HealthHandler{
		dbConn: dbConn,
		logger: logger,
	}
}

// HealthCheck проверяет состояние сервиса и его зависимостей
// @Summary Проверка здоровья сервиса
// @Description Возвращает статус сервиса и его зависимостей (БД)
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} dto.HealthResponse
// @Failure 503 {object} dto.ErrorResponse
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	services := make(map[string]string)
	allHealthy := true

	// Проверяем PostgreSQL
	if err := h.dbConn.Ping(ctx); err != nil {
		h.logger.WithError(err).Error("PostgreSQL health check failed")
		services["postgresql"] = "unhealthy: " + err.Error()
		allHealthy = false
	} else {
		services["postgresql"] = "healthy"
	}

	// Добавляем информацию о приложении
	services["application"] = "healthy"

	if allHealthy {
		h.logger.Debug("Health check passed")
		c.JSON(http.StatusOK, dto.NewHealthResponse(services))
	} else {
		h.logger.Warn("Health check failed - some services are unhealthy")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "unhealthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"services":  services,
		})
	}
}
