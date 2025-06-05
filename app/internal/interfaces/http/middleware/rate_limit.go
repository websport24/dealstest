package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimitMiddleware создает middleware для глобального ограничения частоты запросов
func RateLimitMiddleware() gin.HandlerFunc {
	// Создаем rate limiter: 5000 запросов в секунду с burst до 10000
	// Это покрывает требования 1000-5000 RPS с запасом
	limiter := rate.NewLimiter(rate.Limit(5000), 10000)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
