package kafka

import (
	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

// getCorrelationID либо выделяет из заголовков сообшения correlation_id, либо создаёт его
func getCorrelationID(msg *sarama.ConsumerMessage) string {
	for _, header := range msg.Headers {
		if string(header.Key) == "correlation_id" {
			return string(header.Value)
		}
	}
	return uuid.New().String()
}
