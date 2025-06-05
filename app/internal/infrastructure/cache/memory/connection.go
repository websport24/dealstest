package memory

import (
	"context"
	"sync"
	"time"
)

// CacheItem представляет элемент кэша с TTL
type CacheItem struct {
	Value     interface{}
	ExpiresAt time.Time
}

// IsExpired проверяет, истек ли срок действия элемента
func (item *CacheItem) IsExpired() bool {
	return time.Now().After(item.ExpiresAt)
}

// MemoryCache представляет кэш в памяти
type MemoryCache struct {
	data    map[string]*CacheItem
	mutex   sync.RWMutex
	cleaner *time.Ticker
	done    chan bool
}

// NewMemoryCache создает новый экземпляр кэша в памяти
func NewMemoryCache(cleanupInterval time.Duration) *MemoryCache {
	cache := &MemoryCache{
		data:    make(map[string]*CacheItem),
		cleaner: time.NewTicker(cleanupInterval),
		done:    make(chan bool),
	}

	// Запускаем горутину для очистки истекших элементов
	go cache.startCleanup()

	return cache
}

// startCleanup запускает периодическую очистку истекших элементов
func (c *MemoryCache) startCleanup() {
	for {
		select {
		case <-c.cleaner.C:
			c.cleanupExpired()
		case <-c.done:
			c.cleaner.Stop()
			return
		}
	}
}

// cleanupExpired удаляет истекшие элементы из кэша
func (c *MemoryCache) cleanupExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for key, item := range c.data {
		if item.IsExpired() {
			delete(c.data, key)
		}
	}
}

// Set устанавливает значение в кэш с TTL
func (c *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = &CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}

	return nil
}

// Get получает значение из кэша
func (c *MemoryCache) Get(ctx context.Context, key string) (interface{}, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return nil, ErrKeyNotFound
	}

	if item.IsExpired() {
		// Удаляем истекший элемент
		c.mutex.RUnlock()
		c.mutex.Lock()
		delete(c.data, key)
		c.mutex.Unlock()
		c.mutex.RLock()
		return nil, ErrKeyNotFound
	}

	return item.Value, nil
}

// Delete удаляет значение из кэша
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.data, key)
	return nil
}

// Exists проверяет существование ключа в кэше
func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return false, nil
	}

	if item.IsExpired() {
		// Удаляем истекший элемент
		c.mutex.RUnlock()
		c.mutex.Lock()
		delete(c.data, key)
		c.mutex.Unlock()
		c.mutex.RLock()
		return false, nil
	}

	return true, nil
}

// IncrBy увеличивает числовое значение на указанную величину
func (c *MemoryCache) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	item, exists := c.data[key]
	if !exists || item.IsExpired() {
		// Создаем новый элемент
		c.data[key] = &CacheItem{
			Value:     value,
			ExpiresAt: time.Now().Add(time.Hour), // Дефолтный TTL для счетчиков
		}
		return value, nil
	}

	// Пытаемся привести к int64
	currentValue, ok := item.Value.(int64)
	if !ok {
		return 0, ErrInvalidType
	}

	newValue := currentValue + value
	item.Value = newValue

	return newValue, nil
}

// MGet получает множественные значения
func (c *MemoryCache) MGet(ctx context.Context, keys []string) (map[string]interface{}, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	result := make(map[string]interface{})

	for _, key := range keys {
		if item, exists := c.data[key]; exists && !item.IsExpired() {
			result[key] = item.Value
		}
	}

	return result, nil
}

// MSet устанавливает множественные значения
func (c *MemoryCache) MSet(ctx context.Context, values map[string]interface{}, ttl time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	expiresAt := time.Now().Add(ttl)

	for key, value := range values {
		c.data[key] = &CacheItem{
			Value:     value,
			ExpiresAt: expiresAt,
		}
	}

	return nil
}

// Clear очищает весь кэш
func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data = make(map[string]*CacheItem)
	return nil
}

// Size возвращает количество элементов в кэше
func (c *MemoryCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.data)
}

// Keys возвращает все ключи в кэше
func (c *MemoryCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var keys []string
	for key := range c.data {
		keys = append(keys, key)
	}

	return keys, nil
}

// Close закрывает кэш и останавливает фоновые процессы
func (c *MemoryCache) Close() error {
	close(c.done)
	return nil
}

// Ping проверяет доступность кэша
func (c *MemoryCache) Ping(ctx context.Context) error {
	return nil // Кэш в памяти всегда доступен
}

// FlushAll очищает весь кэш (алиас для Clear)
func (c *MemoryCache) FlushAll(ctx context.Context) error {
	return c.Clear(ctx)
}
