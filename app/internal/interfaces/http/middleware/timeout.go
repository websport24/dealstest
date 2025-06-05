package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// TimeoutMiddleware создает middleware для ограничения времени выполнения запроса
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Заменяем контекст в запросе
		c.Request = c.Request.WithContext(ctx)

		// Канал для отслеживания завершения обработки
		finished := make(chan struct{})

		go func() {
			defer close(finished)
			c.Next()
		}()

		select {
		case <-finished:
			// Запрос завершился вовремя
			return
		case <-ctx.Done():
			// Превышен таймаут
			c.JSON(http.StatusRequestTimeout, gin.H{
				"error":   "Request timeout",
				"message": "Request took too long to process",
			})
			c.Abort()
			return
		}
	}
}
