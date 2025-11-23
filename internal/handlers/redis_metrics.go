package handlers

import (
	"delivery-system/internal/logger"
	"delivery-system/internal/services"
	"net/http"
)

// CacheHandler - хендлер для статистике по кешу
type CacheHandler struct {
	redisService *services.RedisService
	log          *logger.Logger
}

// NewCacheHandler возвращает ссылку на экземпляр CacheHandler
func NewCacheHandler(redisService *services.RedisService, log *logger.Logger) *CacheHandler {
	return &CacheHandler{redisService: redisService, log: log}
}

// GetStatistics получает статистику кеша
func (h *CacheHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Собираем статистику кеша
	metricsPtr, err := h.redisService.GetStatistics(r.Context())
	if err != nil {
		h.log.WithError(err).Error("Failed getting statistics")
		writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, metricsPtr)
}
