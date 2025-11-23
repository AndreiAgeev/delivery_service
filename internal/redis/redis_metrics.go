package redis

import "sync/atomic"

// RedisMetrics - структура метрик кеша. Для счётчиков используются атомики
type RedisMetrics struct {
	CacheHit  atomic.Uint64
	CacheMiss atomic.Uint64
}

// Hit увеличивает счётчик CacheHit на 1
func (m *RedisMetrics) Hit() {
	m.CacheHit.Add(1)
}

// Miss увеличивает счётчик CacheMiss на 1
func (m *RedisMetrics) Miss() {
	m.CacheMiss.Add(1)
}
