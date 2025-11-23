package services

import (
	"context"
	"delivery-system/internal/logger"
	"delivery-system/internal/models"
	"delivery-system/internal/redis"
)

// RedisService - сервис для работы с метриками
type RedisService struct {
	redisClient *redis.Client
	log         *logger.Logger
}

// NewRedisService ссылку на эзкемпляр RedisService
func NewRedisService(redisClient *redis.Client, log *logger.Logger) *RedisService {
	return &RedisService{redisClient: redisClient, log: log}
}

// GetStatistics возвращает статистику по кешу (hit_rate, miss_rate, cache_size)
func (s *RedisService) GetStatistics(ctx context.Context) (*models.RedisMetricsResponse, error) {
	hits, misses, cacheSize, err := s.redisClient.GetMetrics(ctx)
	if err != nil {
		s.log.WithError(err).Error("error getting metrics")
	}
	total := hits + misses

	hitRate := 0.0
	missRate := 0.0
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
		missRate = float64(misses) / float64(total) * 100
	}
	return &models.RedisMetricsResponse{
		HitRate:   hitRate,
		MissRate:  missRate,
		CacheSize: cacheSize,
	}, nil
}
