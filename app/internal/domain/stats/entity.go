package stats

import (
	"errors"
	"time"
)

// Stat представляет статистику кликов за определенный период
type Stat struct {
	ID        int64     `json:"id" db:"id"`
	BannerID  int64     `json:"banner_id" db:"banner_id"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Count     int64     `json:"count" db:"count"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// StatPeriod представляет период для статистики
type StatPeriod struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// StatsResponse представляет ответ с агрегированной статистикой
type StatsResponse struct {
	BannerID int64         `json:"banner_id"`
	Period   StatPeriod    `json:"period"`
	Stats    []*MinuteStat `json:"stats"`
	Total    int64         `json:"total"`
}

// MinuteStat представляет статистику за одну минуту
type MinuteStat struct {
	Timestamp time.Time `json:"ts"`
	Value     int64     `json:"v"`
}

// Доменные ошибки
var (
	ErrInvalidBannerID  = errors.New("invalid banner ID")
	ErrInvalidTimestamp = errors.New("invalid timestamp")
	ErrInvalidPeriod    = errors.New("invalid time period")
	ErrStatNotFound     = errors.New("stat not found")
	ErrPeriodTooLarge   = errors.New("time period is too large")
	ErrNegativeCount    = errors.New("count cannot be negative")
)

// Константы для валидации
const (
	MaxPeriodDays = 30 // Максимальный период запроса в днях
)

// NewStat создает новую статистику с валидацией
func NewStat(bannerID int64, timestamp time.Time, count int64) (*Stat, error) {
	if bannerID <= 0 {
		return nil, ErrInvalidBannerID
	}

	if timestamp.IsZero() {
		return nil, ErrInvalidTimestamp
	}

	if count < 0 {
		return nil, ErrNegativeCount
	}

	now := time.Now()
	return &Stat{
		BannerID:  bannerID,
		Timestamp: timestamp.Truncate(time.Minute), // Округляем до минуты
		Count:     count,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// NewStatPeriod создает новый период с валидацией
func NewStatPeriod(from, to time.Time) (*StatPeriod, error) {
	if from.IsZero() || to.IsZero() {
		return nil, ErrInvalidTimestamp
	}

	if from.After(to) {
		return nil, ErrInvalidPeriod
	}

	// Проверяем, что период не слишком большой
	if to.Sub(from) > time.Duration(MaxPeriodDays)*24*time.Hour {
		return nil, ErrPeriodTooLarge
	}

	return &StatPeriod{
		From: from,
		To:   to,
	}, nil
}

// NewMinuteStat создает статистику за минуту
func NewMinuteStat(timestamp time.Time, value int64) *MinuteStat {
	return &MinuteStat{
		Timestamp: timestamp.Truncate(time.Minute),
		Value:     value,
	}
}

// IsValid проверяет валидность статистики
func (s *Stat) IsValid() error {
	if s.BannerID <= 0 {
		return ErrInvalidBannerID
	}

	if s.Timestamp.IsZero() {
		return ErrInvalidTimestamp
	}

	if s.Count < 0 {
		return ErrNegativeCount
	}

	return nil
}

// CalculateTotal вычисляет общую сумму кликов
func (sr *StatsResponse) CalculateTotal() {
	total := int64(0)
	for _, stat := range sr.Stats {
		total += stat.Value
	}
	sr.Total = total
}

// SortByTimestamp сортирует статистику по времени
func (sr *StatsResponse) SortByTimestamp() {
	// Простая сортировка пузырьком (для небольших массивов)
	n := len(sr.Stats)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if sr.Stats[j].Timestamp.After(sr.Stats[j+1].Timestamp) {
				sr.Stats[j], sr.Stats[j+1] = sr.Stats[j+1], sr.Stats[j]
			}
		}
	}
}
