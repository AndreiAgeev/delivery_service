package handler_tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/mock"

	"delivery-system/internal/handlers"
	"delivery-system/internal/kafka/kafka_mocks"
	"delivery-system/internal/logger"
	"delivery-system/internal/redis/redis_mocks"
	"delivery-system/internal/services/services_mocks"
)

// TestGetOrder выполняеи тестирование получения заказа
func TestGetOrder(t *testing.T) {
	mockOrderService := services_mocks.NewMockOrderServiceInterface(t)
	mockReviewService := services_mocks.NewMockReviewServiceInterface(t)
	mockProducer := kafka_mocks.NewMockProducerInterface(t)
	mockRedis := redis_mocks.NewMockRedisClientInterface(t)
	discardLogger := logger.NewTest()

	// Создаём хендлер
	h := handlers.NewOrderHandler(mockOrderService, mockReviewService, mockProducer, mockRedis, discardLogger)
	mux := setupTestOrderRoutes(h)

	mockRedis.
		On("Get", mock.Anything, mock.Anything, mock.Anything).
		Return(errorNotFound).Maybe().Times(len(getOrderTestCases))

	mockRedis.
		On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Maybe().Once()

	mockRedis.On("Miss").Times(len(getOrderTestCases))
	mockRedis.On("Hit").Once()

	for _, tc := range getOrderTestCases {
		tc := tc
		// Задаём ожидания для мока OrderService
		mockOrderService.
			On("GetOrder", tc.id).
			Return(tc.returnedValue, tc.returnedError)

		server := httptest.NewServer(mux)

		e := httpexpect.Default(t, server.URL)
		e.GET(fmt.Sprintf("/api/orders/%s", tc.id)).Expect().Status(tc.expectedStatusCode)

		mockOrderService.AssertExpectations(t)
		server.Close()
	}
}

// TestCreateOrder выполняет тестирование создания заказа
func TestCreateOrder(t *testing.T) {
	mockReviewService := services_mocks.NewMockReviewServiceInterface(t)
	mockProducer := kafka_mocks.NewMockProducerInterface(t)
	mockRedis := redis_mocks.NewMockRedisClientInterface(t)
	discardLogger := logger.NewTest()

	for _, tc := range createOrderTestCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mockOrderService := services_mocks.NewMockOrderServiceInterface(t)

			h := handlers.NewOrderHandler(mockOrderService, mockReviewService, mockProducer, mockRedis, discardLogger)
			mux := setupTestOrderRoutes(h)

			server := httptest.NewServer(mux)

			if tc.expectedStatusCode != http.StatusBadRequest {
				mockOrderService.On("CreateOrder", tc.payload).Return(tc.returnedValue, tc.returnedError)
				mockProducer.On("PublishOrderCreated", mock.Anything).Return(nil).Maybe()
				mockRedis.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
			}

			e := httpexpect.Default(t, server.URL)

			obj := e.POST("/api/orders").WithJSON(tc.payload).Expect().Status(tc.expectedStatusCode).JSON().Object()
			if tc.expectedStatusCode == http.StatusCreated {
				obj.Value("customer_name").String().IsEqual(tc.payload.CustomerName)
				obj.Value("customer_phone").String().IsEqual(tc.payload.CustomerPhone)
				obj.Value("pickup_address").String().IsEqual(tc.payload.PickupAddress)
				obj.Value("delivery_address").String().IsEqual(tc.payload.DeliveryAddress)
				itemsArray := obj.Value("items").Array()
				itemsArray.Length().IsEqual(len(tc.payload.Items))
				for i, item := range tc.payload.Items {
					expectedItem := itemsArray.Value(i).Object()
					expectedItem.Value("name").String().IsEqual(item.Name)
					expectedItem.Value("quantity").Number().IsEqual(item.Quantity)
					expectedItem.Value("price").Number().IsEqual(item.Price)
					expectedItem.Value("id").String().NotEmpty()
					expectedItem.Value("order_id").String().NotEmpty()
				}
				obj.Value("id").String().NotEmpty()
			}
			mockOrderService.AssertExpectations(t)
			server.Close()
		})
	}
}

// TestUpdateOrderStatus выполняет тестировани обновление статуса заказа
func TestUpdateOrderStatus(t *testing.T) {
	// Создаём моки сервисов
	mockOrderService := services_mocks.NewMockOrderServiceInterface(t)
	mockReviewService := services_mocks.NewMockReviewServiceInterface(t)
	mockProducer := kafka_mocks.NewMockProducerInterface(t)
	mockRedis := redis_mocks.NewMockRedisClientInterface(t)
	discardLogger := logger.NewTest()

	// Создаём хендлер
	h := handlers.NewOrderHandler(mockOrderService, mockReviewService, mockProducer, mockRedis, discardLogger)
	mux := setupTestOrderRoutes(h)

	// Задаём ожидания для моков OrderService, Kafka Producer, Redis Client
	mockProducer.
		On("PublishOrderStatusChanged", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Maybe().Once()
	mockRedis.
		On("Delete", mock.Anything, mock.Anything).
		Return(nil).Maybe().Once()
	mockOrderService.
		On("GetOrder", mock.AnythingOfType("uuid.UUID")).
		Return(order1, nil)

	for _, tc := range updateOrderStatusTestCases {
		tc := tc
		mockOrderService.
			On("UpdateOrderStatus", tc.id, mock.AnythingOfType("*models.UpdateOrderStatusRequest")).
			Return(tc.returnedError)

		server := httptest.NewServer(mux)

		e := httpexpect.Default(t, server.URL)
		e.PUT(fmt.Sprintf("/api/orders/%s/status", tc.id)).WithJSON(tc.payload).Expect().Status(tc.expectedStatusCode)

		server.Close()
	}
	mockProducer.AssertExpectations(t)
	mockRedis.AssertExpectations(t)
	mockOrderService.AssertExpectations(t)
}

// TestGetOrders выполняет тестирование получения списка заказов
func TestGetOrders(t *testing.T) {
	mockReviewService := services_mocks.NewMockReviewServiceInterface(t)
	mockProducer := kafka_mocks.NewMockProducerInterface(t)
	mockRedis := redis_mocks.NewMockRedisClientInterface(t)
	discardLogger := logger.NewTest()

	for _, tc := range getOrdersTestCases {
		mockOrderService := services_mocks.NewMockOrderServiceInterface(t)

		h := handlers.NewOrderHandler(mockOrderService, mockReviewService, mockProducer, mockRedis, discardLogger)
		mux := setupTestOrderRoutes(h)

		tc := tc
		if tc.expectedStatusCode != http.StatusBadRequest {
			mockOrderService.
				On("GetOrders", tc.status, tc.courierID, tc.limit, tc.offset).
				Return(tc.returnedValue, tc.returnedError)
		}

		server := httptest.NewServer(mux)

		e := httpexpect.Default(t, server.URL)
		req := e.GET("/api/orders")
		if tc.expectedStatusCode != http.StatusOK {
			switch tc.queryParam {
			case "courier_id":
				req.WithQuery(tc.queryParam, 1).Expect().Status(tc.expectedStatusCode)
			default:
				req.Expect().Status(tc.expectedStatusCode)
			}
		} else {
			var ordersList *httpexpect.Array

			switch tc.queryParam {
			case "status":
				ordersList = req.WithQuery(tc.queryParam, *tc.status).Expect().Status(tc.expectedStatusCode).JSON().Array()
				for _, order := range ordersList.Iter() {
					status := string(*tc.status)
					order.Object().Value("status").String().IsEqual(status)
				}
			case "courier_id":
				ordersList = req.WithQuery(tc.queryParam, *tc.courierID).Expect().Status(tc.expectedStatusCode).JSON().Array()
				for _, order := range ordersList.Iter() {
					id := tc.courierID.String()
					order.Object().Value("courier_id").String().IsEqual(id)
				}
			case "limit":
				ordersList = req.WithQuery(tc.queryParam, tc.limit).Expect().Status(tc.expectedStatusCode).JSON().Array()
			case "offset":
				ordersList = req.WithQuery(tc.queryParam, tc.offset).Expect().Status(tc.expectedStatusCode).JSON().Array()
			default:
				ordersList = req.Expect().Status(tc.expectedStatusCode).JSON().Array()
			}

			ordersList.Length().IsEqual(len(tc.returnedValue))
		}

		mockOrderService.AssertExpectations(t)
		server.Close()
	}
}

// TestCreateReview выполняет тестирование создания отзыва
func TestCreateReview(t *testing.T) {
	mockProducer := kafka_mocks.NewMockProducerInterface(t)
	mockRedis := redis_mocks.NewMockRedisClientInterface(t)
	discardLogger := logger.NewTest()

	mockRedis.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(errorNotFound).Twice()
	mockRedis.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	for _, tc := range createOrderReviewTestCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mockOrderService := services_mocks.NewMockOrderServiceInterface(t)
			mockReviewService := services_mocks.NewMockReviewServiceInterface(t)

			h := handlers.NewOrderHandler(mockOrderService, mockReviewService, mockProducer, mockRedis, discardLogger)
			mux := setupTestOrderRoutes(h)

			server := httptest.NewServer(mux)

			if tc.expectedStatusCode != http.StatusBadRequest {
				mockOrderService.On("GetOrder", tc.order.ID).Return(tc.order, nil)
				mockReviewService.On("CreateReview", tc.payload, tc.order).Return(tc.returnedValue, tc.returnedError)
				if tc.expectedStatusCode == http.StatusCreated {
					mockReviewService.On("RecalculateRating", mock.Anything).Return(nil)
				}
			}

			e := httpexpect.Default(t, server.URL)

			obj := e.POST(fmt.Sprintf("/api/orders/%s/review", tc.orderID)).
				WithJSON(tc.payload).Expect().Status(tc.expectedStatusCode).
				JSON().Object()
			if tc.expectedStatusCode == http.StatusCreated {
				obj.Value("id").String().NotEmpty()
				obj.Value("order_id").String().IsEqual(tc.returnedValue.OrderID.String())
				obj.Value("courier_id").String().IsEqual(tc.returnedValue.CourierID.String())
				obj.Value("rating").Number().IsEqual(tc.payload.Rating)
				obj.Value("text").String().IsEqual(tc.payload.Text)
			}
			mockReviewService.AssertExpectations(t)
			server.Close()
		})
	}
}
