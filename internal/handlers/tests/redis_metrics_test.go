package handler_tests

import (
	"delivery-system/internal/handlers"
	"delivery-system/internal/logger"
	"delivery-system/internal/services/services_mocks"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/mock"
)

// TestRedisGetStatistics выполняет тестирование на получение метрик Redis
func TestRedisGetStatistics(t *testing.T) {
	for _, tc := range getRedisMetricsTestCases {
		mockRedis := services_mocks.NewMockRedisServiceInterface(t)
		discardLogger := logger.NewTest()

		h := handlers.NewRedisMetricsHandler(mockRedis, discardLogger)
		mux := setupTestRedisMetricsRoute(h)

		server := httptest.NewServer(mux)

		e := httpexpect.Default(t, server.URL)
		mockRedis.On("GetStatistics", mock.Anything).
			Return(tc.returnedValue, tc.returnedError)

		obj := e.GET("/api/cache/metrics").Expect().Status(tc.expectedStatusCode).JSON().Object()
		if tc.expectedStatusCode == http.StatusOK {
			obj.Value("hit_rate").Number().IsEqual(tc.returnedValue.HitRate)
			obj.Value("miss_rate").Number().IsEqual(tc.returnedValue.MissRate)
			obj.Value("cache_size").Number().IsEqual(tc.returnedValue.CacheSize)
		}
		mockRedis.AssertExpectations(t)
		server.Close()
	}
}
