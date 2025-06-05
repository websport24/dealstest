package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/clickcounter/app/internal/application/usecase"
	"github.com/clickcounter/app/internal/domain/stats"
	"github.com/clickcounter/app/internal/interfaces/http/dto"
)

// StatsHandler обрабатывает HTTP запросы для статистики
type StatsHandler struct {
	statsUseCase *usecase.StatsUseCase
	logger       *logrus.Logger
}

// NewStatsHandler создает новый обработчик статистики
func NewStatsHandler(statsUseCase *usecase.StatsUseCase, logger *logrus.Logger) *StatsHandler {
	return &StatsHandler{
		statsUseCase: statsUseCase,
		logger:       logger,
	}
}

// GetStats возвращает статистику кликов по баннеру за период
// @Summary Получение статистики
// @Description Возвращает поминутную статистику кликов по баннеру за указанный период
// @Tags stats
// @Accept json
// @Produce json
// @Param bannerID path int true "ID баннера"
// @Param request body dto.StatsRequest true "Параметры запроса статистики"
// @Success 200 {object} dto.StatsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /stats/{bannerID} [post]
func (h *StatsHandler) GetStats(c *gin.Context) {
	// Извлекаем ID баннера из URL
	bannerIDStr := c.Param("bannerID")
	bannerID, err := strconv.ParseInt(bannerIDStr, 10, 64)
	if err != nil {
		h.logger.WithError(err).WithField("bannerID", bannerIDStr).Error("Invalid banner ID format")
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			http.StatusBadRequest,
			dto.ErrInvalidBannerID,
			"Banner ID must be a valid integer",
		))
		return
	}

	// Валидируем ID баннера
	if bannerID <= 0 {
		h.logger.WithField("bannerID", bannerID).Error("Invalid banner ID value")
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			http.StatusBadRequest,
			dto.ErrInvalidBannerID,
			"Banner ID must be positive",
		))
		return
	}

	// Парсим тело запроса
	var req dto.StatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to parse request body")
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			http.StatusBadRequest,
			err,
			"Invalid request body format",
		))
		return
	}

	// Валидируем запрос
	if err := req.Validate(); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"from": req.From,
			"to":   req.To,
		}).Error("Invalid time range in request")
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			http.StatusBadRequest,
			err,
			"Invalid time range parameters",
		))
		return
	}

	// Получаем временные метки
	fromTime, err := req.GetFromTime()
	if err != nil {
		h.logger.WithError(err).WithField("from", req.From).Error("Invalid from time format")
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			http.StatusBadRequest,
			dto.ErrInvalidTimeFormat,
			"Invalid 'from' time format, expected RFC3339",
		))
		return
	}

	toTime, err := req.GetToTime()
	if err != nil {
		h.logger.WithError(err).WithField("to", req.To).Error("Invalid to time format")
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
			http.StatusBadRequest,
			dto.ErrInvalidTimeFormat,
			"Invalid 'to' time format, expected RFC3339",
		))
		return
	}

	// Получаем статистику
	statsReq := &usecase.GetStatsRequest{
		BannerID: bannerID,
		From:     fromTime,
		To:       toTime,
	}

	statsResponse, err := h.statsUseCase.GetStats(c.Request.Context(), statsReq)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"bannerID": bannerID,
			"from":     fromTime,
			"to":       toTime,
		}).Error("Failed to get stats")

		// Определяем тип ошибки для правильного HTTP статуса
		switch err.Error() {
		case "banner not found":
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(
				http.StatusNotFound,
				err,
				"Banner with specified ID not found",
			))
		case "time period is too large":
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
				http.StatusBadRequest,
				err,
				"Requested time period is too large",
			))
		default:
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
				http.StatusInternalServerError,
				err,
				"Internal server error while retrieving statistics",
			))
		}
		return
	}

	// Логируем успешный запрос
	h.logger.WithFields(logrus.Fields{
		"bannerID":   bannerID,
		"from":       fromTime,
		"to":         toTime,
		"statsCount": len(statsResponse.Stats),
	}).Info("Stats retrieved successfully")

	// Конвертируем в доменную модель для DTO
	domainStats := &stats.StatsResponse{
		Stats: make([]*stats.MinuteStat, len(statsResponse.Stats)),
	}

	for i, stat := range statsResponse.Stats {
		domainStats.Stats[i] = &stats.MinuteStat{
			Timestamp: stat.Timestamp,
			Value:     stat.Count,
		}
	}

	// Возвращаем статистику
	c.JSON(http.StatusOK, dto.NewStatsResponse(domainStats))
}
