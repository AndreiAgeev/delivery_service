package kafka

import (
	"delivery-system/internal/config"
	"delivery-system/internal/logger"
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

// DLQProducer представляет собой Kafka Producer, работающий с Dead Letter Queue
type DLQProducer struct {
	producer sarama.SyncProducer
	log      *logger.Logger
	topic    string
}

// NewDLQProducer возвращает экземпляр объекта DLQProducer
func NewDLQProducer(cfg *config.KafkaConfig, log *logger.Logger) (*DLQProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Producer.Return.Successes = true
	config.Producer.Compression = sarama.CompressionSnappy

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		log.WithError(err).Error("failed to create DLQ producer")
		return nil, fmt.Errorf("failed to create DLQ producer: %w", err)
	}

	log.Info("DLQ producer created successfully")

	return &DLQProducer{
		producer: producer,
		log:      log,
		topic:    "dead_letter_queue",
	}, nil
}

// Close закрывает DLQProducer
func (p *DLQProducer) Close() error { return p.producer.Close() }

// PublishFailedEvent публикует неуспешно обработанное сообщение в DLQ
func (p *DLQProducer) PublishFailedEvent(msg *sarama.ConsumerMessage, originalErr, correlationID string) error {
	message := &sarama.ProducerMessage{
		Topic:     msg.Topic,
		Key:       sarama.StringEncoder(msg.Key),
		Value:     sarama.ByteEncoder(msg.Value),
		Partition: msg.Partition,
		Offset:    msg.Offset,
		Timestamp: time.Now(),
		Headers: []sarama.RecordHeader{
			{Key: []byte("original_error"), Value: []byte(originalErr)},
			{Key: []byte("correlation_id"), Value: []byte(correlationID)},
		},
	}
	_, _, err := p.producer.SendMessage(message)
	return err
}
