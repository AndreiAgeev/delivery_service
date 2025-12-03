package handler_tests

import (
	"delivery-system/internal/handlers"
	"delivery-system/internal/kafka/kafka_mocks"
	"delivery-system/internal/logger"
	"delivery-system/internal/models"
	"delivery-system/internal/redis/redis_mocks"
	"delivery-system/internal/services/services_mocks"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/mock"
)

// TestGetCourier выполняеи тестирование получения курьера
func TestGetCourier(t *testing.T) {
	mockCourierService := services_mocks.NewMockCourierServiceInterface(t)
	mockReviewService := services_mocks.NewMockReviewServiceInterface(t)
	mockProducer := kafka_mocks.NewMockProducerInterface(t)
	mockRedis := redis_mocks.NewMockRedisClientInterface(t)
	discardLogger := logger.NewTest()

	// Создаём хендлер
	h := handlers.NewCourierHandler(mockCourierService, mockReviewService, mockProducer, mockRedis, discardLogger)
	mux := setupTestCourierRoutes(h)

	mockRedis.
		On("Get", mock.Anything, mock.Anything, mock.Anything).
		Return(errorNotFound).Times(len(getOrderTestCases))
	mockRedis.
		On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Once()
	mockRedis.On("Miss").Times(len(getOrderTestCases))
	mockRedis.On("Hit").Once()

	for _, tc := range getCourierTestCases {
		tc := tc
		// Задаём ожидания для мока OrderService
		mockCourierService.
			On("GetCourier", tc.id).
			Return(tc.returnedValue, tc.returnedError)

		server := httptest.NewServer(mux)

		e := httpexpect.Default(t, server.URL)
		e.GET(fmt.Sprintf("/api/couriers/%s", tc.id)).Expect().Status(tc.expectedStatusCode)

		mockCourierService.AssertExpectations(t)
		server.Close()
	}
	mockRedis.AssertExpectations(t)
}

// TestCreateCourier выполняет тестирование создания курьера
func TestCreateCourier(t *testing.T) {
	mockReviewService := services_mocks.NewMockReviewServiceInterface(t)
	mockProducer := kafka_mocks.NewMockProducerInterface(t)
	mockRedis := redis_mocks.NewMockRedisClientInterface(t)
	discardLogger := logger.NewTest()

	mockRedis.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	for _, tc := range createCourierTestCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mockCourierService := services_mocks.NewMockCourierServiceInterface(t)

			h := handlers.NewCourierHandler(mockCourierService, mockReviewService, mockProducer, mockRedis, discardLogger)
			mux := setupTestCourierRoutes(h)

			server := httptest.NewServer(mux)

			if tc.expectedStatusCode != http.StatusBadRequest {
				mockCourierService.On("CreateCourier", tc.payload).Return(tc.returnedValue, tc.returnedError)
			}

			e := httpexpect.Default(t, server.URL)

			obj := e.POST("/api/couriers").WithJSON(tc.payload).Expect().Status(tc.expectedStatusCode).JSON().Object()
			if tc.expectedStatusCode == http.StatusCreated {
				obj.Value("id").String().NotEmpty()
				obj.Value("name").String().IsEqual(tc.payload.Name)
				obj.Value("phone").String().IsEqual(tc.payload.Phone)
			}
			mockCourierService.AssertExpectations(t)
			server.Close()
		})
	}
	mockRedis.AssertExpectations(t)
}

// TestUpdateCourierStatus выполняет тестировани обновление статуса курьера
func TestUpdateCourierStatus(t *testing.T) {
	mockCourierService := services_mocks.NewMockCourierServiceInterface(t)
	mockReviewService := services_mocks.NewMockReviewServiceInterface(t)
	mockProducer := kafka_mocks.NewMockProducerInterface(t)
	mockRedis := redis_mocks.NewMockRedisClientInterface(t)
	discardLogger := logger.NewTest()

	h := handlers.NewCourierHandler(mockCourierService, mockReviewService, mockProducer, mockRedis, discardLogger)
	mux := setupTestCourierRoutes(h)

	mockProducer.
		On("PublishCourierStatusChanged", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Maybe().Once()
	mockRedis.
		On("Delete", mock.Anything, mock.Anything).
		Return(nil).Maybe().Once()
	mockCourierService.
		On("GetCourier", mock.AnythingOfType("uuid.UUID")).
		Return(courier1, nil)
	for _, tc := range updateCourierStatusTestCases {
		tc := tc
		mockCourierService.
			On("UpdateCourierStatus", tc.id, mock.AnythingOfType("*models.UpdateCourierStatusRequest")).
			Return(tc.returnedError)

		server := httptest.NewServer(mux)

		e := httpexpect.Default(t, server.URL)
		e.PUT(fmt.Sprintf("/api/couriers/%s/status", tc.id)).WithJSON(tc.payload).Expect().Status(tc.expectedStatusCode)

		server.Close()
	}
	mockProducer.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
	mockCourierService.AssertExpectations(t)
}

// TestGetCouriers выполняет тестирование получения списка курьеров
func TestGetCouriers(t *testing.T) {
	mockReviewService := services_mocks.NewMockReviewServiceInterface(t)
	mockProducer := kafka_mocks.NewMockProducerInterface(t)
	mockRedis := redis_mocks.NewMockRedisClientInterface(t)
	discardLogger := logger.NewTest()

	for _, tc := range getCouriersTestCases {
		mockCourierService := services_mocks.NewMockCourierServiceInterface(t)

		h := handlers.NewCourierHandler(mockCourierService, mockReviewService, mockProducer, mockRedis, discardLogger)
		mux := setupTestCourierRoutes(h)

		tc := tc
		if tc.expectedStatusCode != http.StatusBadRequest {
			mockCourierService.
				On("GetCouriers", tc.status, tc.limit, tc.offset, tc.ratingSort).
				Return(tc.returnedValue, tc.returnedError)
		}

		server := httptest.NewServer(mux)

		e := httpexpect.Default(t, server.URL)
		req := e.GET("/api/couriers")
		if tc.expectedStatusCode != http.StatusOK {
			switch tc.queryParam {
			case "courier_id":
				req.WithQuery(tc.queryParam, 1).Expect().Status(tc.expectedStatusCode)
			default:
				req.Expect().Status(tc.expectedStatusCode)
			}
		} else {
			var couriersList *httpexpect.Array

			switch tc.queryParam {
			case "status":
				couriersList = req.WithQuery(tc.queryParam, *tc.status).Expect().Status(tc.expectedStatusCode).JSON().Array()
				for _, courier := range couriersList.Iter() {
					status := string(*tc.status)
					courier.Object().Value("status").String().IsEqual(status)
				}
			case "rating_sort":
				// Проверяем сортировку по рейтингу
				couriersList = req.WithQuery(tc.queryParam, tc.ratingSort).Expect().Status(tc.expectedStatusCode).JSON().Array()

				// Создаём переменные для записи пред. рейтинга и объекта курьера
				var prevRating float64
				var courierModel models.Courier

				// Проходимся по полученному списку
				for i, courier := range couriersList.Iter() {
					// Анмаршалим в courierModel полученный объект
					courier.Object().Decode(&courierModel)
					if i == 0 {
						if courierModel.Rating == nil {
							prevRating = 1
						} else {
							prevRating = *courierModel.Rating
						}
						continue
					}

					if prevRating == 1 && courierModel.Rating != nil {
						t.Fatal("Sorting by rating failed")
					} else {
						if prevRating == 1 && *courierModel.Rating == 1 {
							continue
						} else if prevRating > 1 && courierModel.Rating == nil {
							prevRating = 1
						} else {
							prevRating = *courierModel.Rating
						}
					}
				}
			case "limit":
				couriersList = req.WithQuery(tc.queryParam, tc.limit).Expect().Status(tc.expectedStatusCode).JSON().Array()
			case "offset":
				couriersList = req.WithQuery(tc.queryParam, tc.offset).Expect().Status(tc.expectedStatusCode).JSON().Array()
			default:
				couriersList = req.Expect().Status(tc.expectedStatusCode).JSON().Array()
			}

			couriersList.Length().IsEqual(len(tc.returnedValue))
		}

		mockCourierService.AssertExpectations(t)
		server.Close()
	}
}

// TestGetAvailableCouriers выполняет тестирование получения списка доступных курьеров
func TestGetAvailableCouriers(t *testing.T) {
	mockReviewService := services_mocks.NewMockReviewServiceInterface(t)
	mockProducer := kafka_mocks.NewMockProducerInterface(t)
	mockRedis := redis_mocks.NewMockRedisClientInterface(t)
	discardLogger := logger.NewTest()

	for _, tc := range getAvailableCouriersTestCases {
		mockCourierService := services_mocks.NewMockCourierServiceInterface(t)

		mockCourierService.On("GetAvailableCouriers").Return(tc.returnedValue, tc.returnedError)

		h := handlers.NewCourierHandler(mockCourierService, mockReviewService, mockProducer, mockRedis, discardLogger)
		mux := setupTestCourierRoutes(h)

		server := httptest.NewServer(mux)

		e := httpexpect.Default(t, server.URL)

		json := e.GET("/api/couriers/available").Expect().Status(tc.expectedStatusCode).JSON()
		if tc.expectedStatusCode == http.StatusOK {
			json.Array().Length().IsEqual(len(tc.returnedValue))
		}
		mockCourierService.AssertExpectations(t)
		server.Close()
	}
}

// TestAssignOrderToCourier выполняет тестирование назначения заказа курьеру
func TestAssignOrderToCourier(t *testing.T) {
	mockReviewService := services_mocks.NewMockReviewServiceInterface(t)
	mockProducer := kafka_mocks.NewMockProducerInterface(t)
	mockRedis := redis_mocks.NewMockRedisClientInterface(t)
	discardLogger := logger.NewTest()

	mockProducer.On("PublishCourierAssigned", mock.Anything, mock.Anything).
		Return(nil).Once()
	mockRedis.On("Delete", mock.Anything, mock.Anything).Return(nil).Twice()

	for _, tc := range assignOrderToCourierTestCases {
		mockCourierService := services_mocks.NewMockCourierServiceInterface(t)

		if !strings.Contains(tc.name, "_bad_") {
			mockCourierService.On("AssignOrderToCourier", tc.payload.OrderID, tc.courierID).Return(tc.returnedError)
		}

		h := handlers.NewCourierHandler(mockCourierService, mockReviewService, mockProducer, mockRedis, discardLogger)
		mux := setupTestCourierRoutes(h)

		server := httptest.NewServer(mux)

		e := httpexpect.Default(t, server.URL)
		e.POST(fmt.Sprintf("/api/couriers/%s/assign", tc.courierID.String())).WithJSON(tc.payload).
			Expect().Status(tc.expectedStatusCode)

		mockCourierService.AssertExpectations(t)
		server.Close()
	}
	mockProducer.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
}

// TestGetCourierReviews выполняет тестирование получения списка отзывов на курьера
func TestGetCourierReviews(t *testing.T) {
	mockCourierService := services_mocks.NewMockCourierServiceInterface(t)
	mockReviewService := services_mocks.NewMockReviewServiceInterface(t)
	mockProducer := kafka_mocks.NewMockProducerInterface(t)
	mockRedis := redis_mocks.NewMockRedisClientInterface(t)
	discardLogger := logger.NewTest()

	h := handlers.NewCourierHandler(mockCourierService, mockReviewService, mockProducer, mockRedis, discardLogger)
	mux := setupTestCourierRoutes(h)

	server := httptest.NewServer(mux)

	for _, tc := range getCourierReviewsTestCases {
		mockReviewService.On("GetReviews", tc.courierID).Return(tc.returnedValue, tc.returnedError)

		e := httpexpect.Default(t, server.URL)

		json := e.GET(fmt.Sprintf("/api/couriers/%s/reviews", tc.courierID.String())).
			Expect().Status(tc.expectedStatusCode).JSON()

		if tc.expectedStatusCode == http.StatusOK {
			json.Array().Length().IsEqual(len(tc.returnedValue))
		}
	}
	mockReviewService.AssertExpectations(t)
	server.Close()
}
