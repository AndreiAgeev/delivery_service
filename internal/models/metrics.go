package models

// RedisMetricsResponse - структура JSON-ответа статистики кеша
type RedisMetricsResponse struct {
	HitRate   float64 `json:"hit_rate"`
	MissRate  float64 `json:"miss_rate"`
	CacheSize int64   `json:"cache_size"`
}

// KafkaTopicMetricsResponse - структура JSON-ответа статистики топика Kafka
type KafkaTopicMetricsResponse struct {
	Topic                 string `json:"topic"`
	TotalProcessedEvents  uint64 `json:"total_processed_events"`
	Errors                uint64 `json:"errors"`
	AvgProcessingDuration string `json:"avg_processing_duration"`
}

// KafkaMetricsResponse - структура JSON-ответа общей статистики Kafka
type KafkaMetricsResponse struct {
	TotalLag   int64                       `json:"total_lag"`
	Statistics []KafkaTopicMetricsResponse `json:"statistics"`
}
