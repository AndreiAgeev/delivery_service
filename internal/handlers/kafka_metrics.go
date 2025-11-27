package handlers

import (
	"delivery-system/internal/logger"
	"delivery-system/internal/services"
	"net/http"
)

type KafkaMetricsHandler struct {
	metricsService *services.KafkaMetricsService
	log            *logger.Logger
}

func NewKafkaMetricsHandler(metricsService *services.KafkaMetricsService, log *logger.Logger) *KafkaMetricsHandler {
	return &KafkaMetricsHandler{
		metricsService: metricsService,
		log:            log,
	}
}

func (h *KafkaMetricsHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем данные метрик
	data := h.metricsService.GetStatistics()
	h.log.Info("Kafka metrics obtained successfully")
	writeJSONResponse(w, http.StatusOK, data)
}
