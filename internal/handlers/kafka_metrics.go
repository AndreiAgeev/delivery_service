package handlers

import (
	"delivery-system/internal/logger"
	"delivery-system/internal/services"
	"net/http"
)

type KafkaMetricsHandler struct {
	metricsService services.KafkaMetricsServiceInterface
	log            *logger.Logger
}

func NewKafkaMetricsHandler(metricsService services.KafkaMetricsServiceInterface, log *logger.Logger) *KafkaMetricsHandler {
	return &KafkaMetricsHandler{
		metricsService: metricsService,
		log:            log,
	}
}

func (h *KafkaMetricsHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем данные метрик
	data := h.metricsService.GetStatistics()
	h.log.Info("Kafka metrics obtained successfully")
	WriteJSONResponse(w, http.StatusOK, data)
}
