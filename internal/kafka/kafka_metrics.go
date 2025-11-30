package kafka

import (
	"delivery-system/internal/config"
	"delivery-system/internal/logger"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"delivery-system/internal/models"

	"github.com/IBM/sarama"
)

// TopicMetrics - структура для ведения метрик конкретного топика Kafka Consumer
type TopicMetrics struct {
	TotalProcessedEvents    atomic.Uint64
	Errors                  atomic.Uint64
	TotalProcessingDuration atomic.Uint64
}

// KafkaMetrics - структура ведения метрик Kafka Consumer
type KafkaMetrics struct {
	mux        sync.RWMutex
	Statistics map[string]*TopicMetrics
	TotalLag   int64
}

func NewKafkaMetrics() *KafkaMetrics {
	return &KafkaMetrics{
		mux:        sync.RWMutex{},
		Statistics: make(map[string]*TopicMetrics),
		TotalLag:   0,
	}
}

func (m *KafkaMetrics) RecordEvent(topic string, duration int64, hasError bool) {
	m.mux.Lock()
	defer m.mux.Unlock()

	// Проверяем наличие TopicMetrics для конкретного топика в мапе
	if _, exists := m.Statistics[topic]; !exists {
		m.Statistics[topic] = &TopicMetrics{}
	}

	// Записываем результаты события
	topicMetrics := m.Statistics[topic]
	topicMetrics.TotalProcessedEvents.Add(1)
	topicMetrics.TotalProcessingDuration.Add(uint64(duration))

	if hasError {
		topicMetrics.Errors.Add(1)
	}
}

func (m *KafkaMetrics) GetStatistics() *models.KafkaMetricsResponse {
	m.mux.RLock()
	defer m.mux.RUnlock()

	// Создаём объект KafkaMetricsResponse для JSON-ответа
	stats := &models.KafkaMetricsResponse{
		TotalLag:   m.TotalLag,
		Statistics: make([]models.KafkaTopicMetricsResponse, 0, len(m.Statistics)),
	}

	// Проходим по каждому метрикам каждого топика
	for topic, metrics := range m.Statistics {
		totalEvents := metrics.TotalProcessedEvents.Load()
		totalDuration := metrics.TotalProcessingDuration.Load()

		// Высчитываем среднее время обработки сообщения
		avg := "0 ms"
		if totalEvents > 0 {
			avgDur := totalDuration / totalEvents
			avg = fmt.Sprintf("%d ms", avgDur)
		}

		// Создаём объект KafkaTopicMetricsResponse для метрик топика
		topicStats := models.KafkaTopicMetricsResponse{
			Topic:                 topic,
			TotalProcessedEvents:  totalEvents,
			Errors:                metrics.Errors.Load(),
			AvgProcessingDuration: avg,
		}

		// Добавляем объект в KafkaMetricsResponse
		stats.Statistics = append(stats.Statistics, topicStats)
	}

	return stats
}

type LagMonitor struct {
	client      sarama.Client
	admin       sarama.ClusterAdmin
	groupID     string
	log         *logger.Logger
	consumerLag int64
	ticker      *time.Ticker
}

func NewLagMonitor(cfg *config.KafkaConfig, log *logger.Logger) (*LagMonitor, error) {
	clientConfig := sarama.NewConfig()
	client, err := sarama.NewClient(cfg.Brokers, clientConfig)
	if err != nil {
		log.WithError(err).Error("Error occurred during creation of Kafka client for Lag Monitor")
		return nil, err
	}
	admin, err := sarama.NewClusterAdmin(cfg.Brokers, clientConfig)
	if err != nil {
		log.WithError(err).Error("Error occurred during creation of Kafka admin client for Lag Monitor")
		return nil, err
	}

	interval := time.Duration(cfg.MonitorInterval) * time.Minute

	return &LagMonitor{
		client:      client,
		admin:       admin,
		groupID:     cfg.GroupID,
		log:         log,
		consumerLag: cfg.ConsumerLag,
		ticker:      time.NewTicker(interval),
	}, nil
}

// Start запускает монитора лагов
func (m *LagMonitor) Start(metrics *KafkaMetrics) {
	go func() {
		for range m.ticker.C {
			m.RecordLag(metrics)
		}
	}()
}

// Stop останавливает монитора лагов
func (m *LagMonitor) Stop() {
	m.ticker.Stop()
}

// RecordLag определяет значение лага в системе и записывает в метрики
func (m *LagMonitor) RecordLag(metrics *KafkaMetrics) {
	// Получаем мапу, содержащую данные о каждом оффсете всех партиций всех топиков
	groupOffsets, err := m.admin.ListConsumerGroupOffsets("abc", nil)
	if err != nil {
		m.log.WithError(err).Error("Failed to get consumer group offsets")
	}

	totalLag := int64(0)
	for topic, partitions := range groupOffsets.Blocks {
		for partitionID, partitionBlock := range partitions {
			if partitionBlock.Offset == -1 {
				continue
			}

			// Получаем оффсет, который будет получен при следующей записис в партицию
			newestOffset, err := m.client.GetOffset(topic, partitionID, sarama.OffsetNewest)
			if err != nil {
				m.log.WithError(err).Error("Failed to get partition offset")
				continue
			}

			// Высчитываем лаг для партиции
			lag := newestOffset - partitionBlock.Offset
			totalLag += lag
		}
		// Делаем алерт в логах, если общий лаг приложения превышает заданный порог
		if totalLag > m.consumerLag {
			m.log.WithFields(map[string]interface{}{
				"lag": totalLag,
			}).Warn("ALERT: High Consumer Lag detected in whole system")
		}
		metrics.TotalLag = totalLag
	}
}
