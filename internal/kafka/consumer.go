package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"delivery-system/internal/config"
	"delivery-system/internal/logger"
	"delivery-system/internal/models"

	"github.com/IBM/sarama"
)

// EventHandler представляет обработчик событий
type EventHandler func(ctx context.Context, event *models.Event) error

// Consumer представляет Kafka consumer
type Consumer struct {
	consumer    sarama.ConsumerGroup
	log         *logger.Logger
	handlers    map[models.EventType]EventHandler
	topics      []string
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	maxRetries  int
	dlqProducer *DLQProducer
	metrics     *KafkaMetrics
}

// NewConsumer создает новый Kafka consumer
func NewConsumer(cfg *config.KafkaConfig, log *logger.Logger, dlqProducer *DLQProducer, metrics *KafkaMetrics) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Session.Timeout = 10000000000   // 10 секунд
	config.Consumer.Group.Heartbeat.Interval = 3000000000 // 3 секунды

	consumer, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	topics := []string{cfg.Topics.Orders, cfg.Topics.Couriers, cfg.Topics.Locations}

	log.Info("Kafka consumer created successfully")

	return &Consumer{
		consumer:    consumer,
		log:         log,
		handlers:    make(map[models.EventType]EventHandler),
		topics:      topics,
		ctx:         ctx,
		cancel:      cancel,
		maxRetries:  3,
		dlqProducer: dlqProducer,
		metrics:     metrics,
	}, nil
}

// RegisterHandler регистрирует обработчик для определенного типа события
func (c *Consumer) RegisterHandler(eventType models.EventType, handler EventHandler) {
	c.handlers[eventType] = handler
	c.log.WithField("event_type", eventType).Info("Event handler registered")
}

// Start запускает consumer
func (c *Consumer) Start() error {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-c.ctx.Done():
				return
			default:
				if err := c.consumer.Consume(c.ctx, c.topics, c); err != nil {
					c.log.WithError(err).Error("Error consuming messages")
				}
			}
		}
	}()

	c.log.Info("Kafka consumer started")
	return nil
}

// Stop останавливает consumer
func (c *Consumer) Stop() error {
	c.cancel()
	c.wg.Wait()
	return c.consumer.Close()
}

// Setup реализует интерфейс sarama.ConsumerGroupHandler
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup реализует интерфейс sarama.ConsumerGroupHandler
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim реализует интерфейс sarama.ConsumerGroupHandler
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			// Получаем correlationID
			correlationID := getCorrelationID(message)

			// Обрабатываем сообщение, определяем время обработки duration
			start := time.Now() // отслеживаем время обработки события в секундах
			err := c.processMessageWithRetries(message, correlationID)
			duration := time.Since(start).Milliseconds()

			// Обновляем метрики
			if err != nil {
				go func() { c.metrics.RecordEvent(message.Topic, duration, false) }()
			} else {
				go func() { c.metrics.RecordEvent(message.Topic, duration, true) }()
			}

			// Проверяем успешность обработки и, при необходимости, отправляем сообщение в DLQ
			if err != nil {
				c.log.WithFields(map[string]interface{}{
					"correlation_id": correlationID,
					"error":          err,
					"topic":          message.Topic,
					"partition":      message.Partition,
					"offset":         message.Offset,
				}).Error("Failed to process message")
				if dlqErr := c.dlqProducer.PublishFailedEvent(message, err.Error(), correlationID); dlqErr != nil {
					return dlqErr
				}
			}
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// processMessage обрабатывает полученное сообщение
func (c *Consumer) processMessageWithRetries(message *sarama.ConsumerMessage, correlationID string) error {
	var event models.Event
	if err := json.Unmarshal(message.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	c.log.WithFields(map[string]interface{}{
		"correlation_id": correlationID,
		"event_type":     event.Type,
		"event_id":       event.ID,
		"topic":          message.Topic,
	}).Debug("Processing event...")

	// Находим обработчик для данного типа события
	handler, exists := c.handlers[event.Type]
	if !exists {
		c.log.WithFields(map[string]interface{}{
			"correlation_id": correlationID,
			"event_type":     event.Type,
		}).Warn("No handler registered for event type")
		return fmt.Errorf("no handler registered for event type %s", event.Type)
	}

	// Обрабатываем сообщение c.maxRetries раз
	for i := 1; i <= c.maxRetries; i++ {
		if err := handler(c.ctx, &event); err == nil {
			c.log.WithField("event_id", event.ID.String()).Info("Message was successfully processed")
			return nil
		}
		c.log.WithFields(map[string]interface{}{
			"correlation_id": correlationID,
			"event_id":       event.ID.String(),
			"attempt":        i,
		}).Warn("Failed to process message")
	}

	return nil
}
