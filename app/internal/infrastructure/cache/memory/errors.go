package memory

import "errors"

var (
	// ErrKeyNotFound возвращается когда ключ не найден в кэше
	ErrKeyNotFound = errors.New("key not found in cache")

	// ErrInvalidType возвращается при попытке операции с неподходящим типом
	ErrInvalidType = errors.New("invalid type for operation")

	// ErrCacheClosed возвращается при попытке операции с закрытым кэшем
	ErrCacheClosed = errors.New("cache is closed")
)
