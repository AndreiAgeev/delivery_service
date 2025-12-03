package handler_tests

import (
	"delivery-system/internal/handlers"
	"net/http"
	"strings"
)

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// setupTestOrderRoutes настраивает HTTP-маршруты для функционала заказов
func setupTestOrderRoutes(h *handlers.OrderHandler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/orders", corsMiddleware(handleOrdersRoute(h)))
	mux.HandleFunc("/api/orders/", corsMiddleware(handleOrderRoute(h)))

	return mux
}

// handleOrdersRoute обрабатывает маршруты для коллекции заказов
func handleOrdersRoute(handler *handlers.OrderHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetOrders(w, r)
		case http.MethodPost:
			handler.CreateOrder(w, r)
		default:
			handlers.WriteErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	}
}

// handleOrderRoute обрабатывает маршруты для отдельного заказа
func handleOrderRoute(handler *handlers.OrderHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/status") {
			// Обновление статуса заказа
			if r.Method == http.MethodPut {
				handler.UpdateOrderStatus(w, r)
			} else {
				handlers.WriteErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
		} else if strings.HasSuffix(r.URL.Path, "/review") {
			if r.Method == http.MethodPost {
				handler.CreateReview(w, r)
			} else {
				handlers.WriteErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
		} else {
			// Получение заказа по ID
			if r.Method == http.MethodGet {
				handler.GetOrder(w, r)
			} else {
				handlers.WriteErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
		}
	}
}

// setupTestCourierRoutes настраивает HTTP-маршруты для функционала курьеров
func setupTestCourierRoutes(h *handlers.CourierHandler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/couriers", corsMiddleware(handleCouriersRoute(h)))
	mux.HandleFunc("/api/couriers/", corsMiddleware(handleCourierRoute(h)))
	mux.HandleFunc("/api/couriers/available", corsMiddleware(h.GetAvailableCouriers))

	return mux
}

// handleCouriersRoute обрабатывает маршруты для коллекции курьеров
func handleCouriersRoute(handler *handlers.CourierHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetCouriers(w, r)
		case http.MethodPost:
			handler.CreateCourier(w, r)
		default:
			handlers.WriteErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	}
}

// handleCourierRoute обрабатывает маршруты для отдельного курьера
func handleCourierRoute(handler *handlers.CourierHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/status") {
			// Обновление статуса курьера
			if r.Method == http.MethodPut {
				handler.UpdateCourierStatus(w, r)
			} else {
				handlers.WriteErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
		} else if strings.HasSuffix(r.URL.Path, "/assign") {
			// Назначение заказа курьеру
			if r.Method == http.MethodPost {
				handler.AssignOrderToCourier(w, r)
			} else {
				handlers.WriteErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
		} else if strings.HasSuffix(r.URL.Path, "/reviews") {
			if r.Method == http.MethodGet {
				handler.GetCourierReviews(w, r)
			} else {
				handlers.WriteErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			}

		} else {
			// Получение курьера по ID
			if r.Method == http.MethodGet {
				handler.GetCourier(w, r)
			} else {
				handlers.WriteErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
		}
	}
}

// setupTestKafkaMetricsRoute настраивает HTTP-маршрут для функционала получения статистики Kafka
func setupTestKafkaMetricsRoute(h *handlers.KafkaMetricsHandler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/kafka/stats", corsMiddleware(h.GetStatistics))

	return mux
}

// setupTestRedisMetricsRoute настраивает HTTP-маршрут для функционала получения статистики Redis
func setupTestRedisMetricsRoute(h *handlers.RedisMetricsHandler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/cache/metrics", corsMiddleware(h.GetStatistics))

	return mux
}
