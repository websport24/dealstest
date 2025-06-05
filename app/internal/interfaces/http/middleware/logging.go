package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// LoggingMiddleware создает middleware для логирования HTTP запросов
func LoggingMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Создаем структурированный лог
		fields := logrus.Fields{
			"method":     param.Method,
			"path":       param.Path,
			"status":     param.StatusCode,
			"latency":    param.Latency,
			"client_ip":  param.ClientIP,
			"user_agent": param.Request.UserAgent(),
			"error":      param.ErrorMessage,
		}

		// Определяем уровень логирования на основе статуса
		switch {
		case param.StatusCode >= 500:
			logger.WithFields(fields).Error("HTTP request completed with server error")
		case param.StatusCode >= 400:
			logger.WithFields(fields).Warn("HTTP request completed with client error")
		case param.StatusCode >= 300:
			logger.WithFields(fields).Info("HTTP request completed with redirect")
		default:
			// Для успешных запросов используем Debug для /health endpoint
			if param.Path == "/health" {
				logger.WithFields(fields).Debug("HTTP request completed successfully")
			} else {
				logger.WithFields(fields).Info("HTTP request completed successfully")
			}
		}

		// Возвращаем пустую строку, так как мы уже залогировали
		return ""
	})
}
