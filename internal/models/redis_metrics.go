package models

// RedisMetricsResponse - структура ответа статистики кеша
type RedisMetricsResponse struct {
	HitRate   float64 `json:"hit_rate"`
	MissRate  float64 `json:"miss_rate"`
	CacheSize int64   `json:"cache_size"`
}
