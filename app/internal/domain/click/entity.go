package click

import (
	"errors"
	"time"
)

// Click представляет сущность клика по баннеру
type Click struct {
	ID        int64     `json:"id" db:"id"`
	BannerID  int64     `json:"banner_id" db:"banner_id"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	UserIP    string    `json:"user_ip,omitempty" db:"user_ip"`
	UserAgent string    `json:"user_agent,omitempty" db:"user_agent"`
}

// Доменные ошибки
var (
	ErrInvalidBannerID  = errors.New("invalid banner ID")
	ErrInvalidTimestamp = errors.New("invalid timestamp")
	ErrClickNotFound    = errors.New("click not found")
)

// NewClick создает новый клик с валидацией
func NewClick(bannerID int64) (*Click, error) {
	if bannerID <= 0 {
		return nil, ErrInvalidBannerID
	}

	return &Click{
		BannerID:  bannerID,
		Timestamp: time.Now(),
	}, nil
}

// NewClickWithMetadata создает новый клик с дополнительными метаданными
func NewClickWithMetadata(bannerID int64, userIP, userAgent string) (*Click, error) {
	click, err := NewClick(bannerID)
	if err != nil {
		return nil, err
	}

	click.UserIP = userIP
	click.UserAgent = userAgent

	return click, nil
}

// IsValid проверяет валидность клика
func (c *Click) IsValid() error {
	if c.BannerID <= 0 {
		return ErrInvalidBannerID
	}

	if c.Timestamp.IsZero() {
		return ErrInvalidTimestamp
	}

	return nil
}

// GetMinuteTimestamp возвращает временную метку, округленную до минуты
// Используется для агрегации статистики по минутам
func (c *Click) GetMinuteTimestamp() time.Time {
	return c.Timestamp.Truncate(time.Minute)
}
