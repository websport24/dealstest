package dto

import (
	"time"
)

// StatsRequest представляет запрос для получения статистики
type StatsRequest struct {
	From string `json:"from" binding:"required" example:"2024-12-12T10:00:00Z"`
	To   string `json:"to" binding:"required" example:"2024-12-12T10:05:00Z"`
}

// Validate проверяет корректность временных меток
func (r *StatsRequest) Validate() error {
	fromTime, err := time.Parse(time.RFC3339, r.From)
	if err != nil {
		return err
	}

	toTime, err := time.Parse(time.RFC3339, r.To)
	if err != nil {
		return err
	}

	if fromTime.After(toTime) {
		return ErrInvalidTimeRange
	}

	return nil
}

// GetFromTime возвращает время начала периода
func (r *StatsRequest) GetFromTime() (time.Time, error) {
	return time.Parse(time.RFC3339, r.From)
}

// GetToTime возвращает время окончания периода
func (r *StatsRequest) GetToTime() (time.Time, error) {
	return time.Parse(time.RFC3339, r.To)
}
