package dto

import (
	"time"

	"github.com/clickcounter/app/internal/domain/stats"
)

// ClickResponse представляет ответ на регистрацию клика
type ClickResponse struct {
	Success   bool   `json:"success" example:"true"`
	BannerID  int64  `json:"banner_id" example:"1"`
	Message   string `json:"message" example:"Click registered successfully"`
	Timestamp string `json:"timestamp" example:"2024-12-12T10:00:00Z"`
}

// StatsResponse представляет ответ со статистикой
type StatsResponse struct {
	Stats []StatItem `json:"stats"`
}

// StatItem представляет элемент статистики
type StatItem struct {
	Timestamp string `json:"ts" example:"2024-12-12T10:00:00Z"`
	Value     int64  `json:"v" example:"4"`
}

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid banner ID"`
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"Banner with ID 999 not found"`
}

// HealthResponse представляет ответ health check
type HealthResponse struct {
	Status    string            `json:"status" example:"ok"`
	Timestamp string            `json:"timestamp" example:"2024-12-12T10:00:00Z"`
	Services  map[string]string `json:"services"`
}

// NewClickResponse создает новый ответ для клика
func NewClickResponse(bannerID int64) *ClickResponse {
	return &ClickResponse{
		Success:   true,
		BannerID:  bannerID,
		Message:   "Click registered successfully",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// NewStatsResponse создает новый ответ со статистикой
func NewStatsResponse(domainStats *stats.StatsResponse) *StatsResponse {
	items := make([]StatItem, len(domainStats.Stats))
	for i, stat := range domainStats.Stats {
		items[i] = StatItem{
			Timestamp: stat.Timestamp.Format(time.RFC3339),
			Value:     stat.Value,
		}
	}

	return &StatsResponse{
		Stats: items,
	}
}

// NewErrorResponse создает новый ответ с ошибкой
func NewErrorResponse(code int, err error, message string) *ErrorResponse {
	return &ErrorResponse{
		Error:   err.Error(),
		Code:    code,
		Message: message,
	}
}

// NewHealthResponse создает новый ответ health check
func NewHealthResponse(services map[string]string) *HealthResponse {
	return &HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services:  services,
	}
}
