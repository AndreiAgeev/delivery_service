package handler_tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"delivery-system/internal/handlers"
	"delivery-system/internal/logger"
	"delivery-system/internal/services/services_mocks"

	"github.com/gavv/httpexpect/v2"
)

// TestKafkaGetStatistics выполняет тестирование на получение метрик Kafka
func TestKafkaGetStatistics(t *testing.T) {
	mockKafkaMetrics := services_mocks.NewMockKafkaMetricsServiceInterface(t)
	discardLogger := logger.NewTest()

	mockKafkaMetrics.On("GetStatistics").Return(kafkaMetrics)

	h := handlers.NewKafkaMetricsHandler(mockKafkaMetrics, discardLogger)
	mux := setupTestKafkaMetricsRoute(h)

	server := httptest.NewServer(mux)
	defer server.Close()

	e := httpexpect.Default(t, server.URL)
	obj := e.GET("/api/kafka/stats").Expect().Status(http.StatusOK).JSON().Object()
	obj.Value("total_lag").IsEqual(kafkaMetrics.TotalLag)
	for i, array := range obj.Value("statistics").Array().Iter() {
		topicStats := kafkaMetrics.Statistics[i]
		array.Object().Value("topic").IsEqual(topicStats.Topic)
		array.Object().Value("total_processed_events").IsEqual(topicStats.TotalProcessedEvents)
		array.Object().Value("errors").IsEqual(topicStats.Errors)
		array.Object().Value("avg_processing_duration").IsEqual(topicStats.AvgProcessingDuration)
	}
	mockKafkaMetrics.AssertExpectations(t)
}
