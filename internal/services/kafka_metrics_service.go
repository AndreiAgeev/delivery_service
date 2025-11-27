package services

import (
	"delivery-system/internal/kafka"
	"delivery-system/internal/models"
)

type KafkaMetricsService struct {
	metrics *kafka.KafkaMetrics
}

func NewKafkaMetricsService(metrics *kafka.KafkaMetrics) *KafkaMetricsService {
	return &KafkaMetricsService{metrics: metrics}
}

func (s *KafkaMetricsService) GetStatistics() *models.KafkaMetricsResponse {
	return s.metrics.GetStatistics()
}
