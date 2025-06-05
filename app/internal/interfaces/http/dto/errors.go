package dto

import "errors"

var (
	// ErrInvalidTimeRange возвращается когда время начала больше времени окончания
	ErrInvalidTimeRange = errors.New("invalid time range: from time must be before to time")

	// ErrInvalidBannerID возвращается при некорректном ID баннера
	ErrInvalidBannerID = errors.New("invalid banner ID")

	// ErrInvalidTimeFormat возвращается при некорректном формате времени
	ErrInvalidTimeFormat = errors.New("invalid time format, expected RFC3339")
)
