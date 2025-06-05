package middleware

import (
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

// Простые атомарные счетчики для базовых метрик
var (
	activeConnections int64
	totalRequests     int64
)

// MetricsMiddleware создает middleware для сбора базовых метрик
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Увеличиваем счетчик активных соединений
		atomic.AddInt64(&activeConnections, 1)
		atomic.AddInt64(&totalRequests, 1)

		// Обрабатываем запрос
		c.Next()

		// Уменьшаем счетчик активных соединений
		atomic.AddInt64(&activeConnections, -1)
	}
}
