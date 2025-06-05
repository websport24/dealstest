package banner

import (
	"errors"
	"time"
)

// Banner представляет сущность баннера в системе
type Banner struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	IsActive  bool      `json:"is_active" db:"is_active"`
}

// Доменные ошибки
var (
	ErrBannerNotFound  = errors.New("banner not found")
	ErrInvalidBannerID = errors.New("invalid banner ID")
)

// IsValid проверяет валидность баннера
func (b *Banner) IsValid() error {
	if b.ID < 0 {
		return ErrInvalidBannerID
	}

	return nil
}
