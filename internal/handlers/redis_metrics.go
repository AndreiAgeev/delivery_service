package handlers

import (
	"delivery-system/internal/logger"
	"delivery-system/internal/services"
	"net/http"
)

// RedisMetricsHandler - хендлер для статистике по кешу
type RedisMetricsHandler struct {
	redisService services.RedisServiceInterface
	log          *logger.Logger
}

// NewRedisMetricsHandler возвращает ссылку на экземпляр CacheHandler
func NewRedisMetricsHandler(redisService services.RedisServiceInterface, log *logger.Logger) *RedisMetricsHandler {
	return &RedisMetricsHandler{redisService: redisService, log: log}
}

// GetStatistics получает статистику кеша
func (h *RedisMetricsHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Собираем статистику кеша
	metricsPtr, err := h.redisService.GetStatistics(r.Context())
	if err != nil {
		h.log.WithError(err).Error("Failed getting statistics")
		WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteJSONResponse(w, http.StatusOK, metricsPtr)
}
