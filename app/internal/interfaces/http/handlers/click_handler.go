package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/clickcounter/app/internal/application/usecase"
	"github.com/clickcounter/app/internal/interfaces/http/dto"
)

// ClickHandler обрабатывает HTTP запросы для кликов
type ClickHandler struct {
	clickUseCase *usecase.ClickUseCase
	logger       *logrus.Logger
}

// NewClickHandler создает новый обработчик кликов
func NewClickHandler(clickUseCase *usecase.ClickUseCase, logger *logrus.Logger) *ClickHandler {
	return &ClickHandler{
		clickUseCase: clickUseCase,
		logger:       logger,
	}
}

// RegisterClick регистрирует клик по баннеру
// @Summary Регистрация клика
// @Description Регистрирует клик по баннеру с указанным ID
// @Tags clicks
// @Accept json
// @Produce json
// @Param bannerID path int true "ID баннера"
// @Success 200 {object} dto.ClickResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /counter/{bannerID} [get]
func (h *ClickHandler) RegisterClick(c *gin.Context) {
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

	// Регистрируем клик
	req := &usecase.RegisterClickRequest{
		BannerID:  bannerID,
		UserIP:    c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}

	response, err := h.clickUseCase.RegisterClick(c.Request.Context(), req)
	if err != nil {
		h.logger.WithError(err).WithField("bannerID", bannerID).Error("Failed to register click")

		// Определяем тип ошибки для правильного HTTP статуса
		if response != nil && !response.Success {
			switch response.Message {
			case "Banner not found":
				c.JSON(http.StatusNotFound, dto.NewErrorResponse(
					http.StatusNotFound,
					err,
					response.Message,
				))
			default:
				c.JSON(http.StatusBadRequest, dto.NewErrorResponse(
					http.StatusBadRequest,
					err,
					response.Message,
				))
			}
		} else {
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
				http.StatusInternalServerError,
				err,
				"Internal server error while registering click",
			))
		}
		return
	}

	// Логируем успешную регистрацию
	h.logger.WithField("bannerID", bannerID).Info("Click registered successfully")

	// Возвращаем успешный ответ
	c.JSON(http.StatusOK, dto.NewClickResponse(bannerID))
}
