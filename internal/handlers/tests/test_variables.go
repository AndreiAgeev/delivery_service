package handler_tests

import (
	"context"
	"delivery-system/internal/models"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Обычные переменные
var orderID = uuid.New()
var courierID = uuid.New()
var itemID = uuid.New()
var reviewID = uuid.New()
var deliveredTime = time.Now().Add(5 * time.Hour)
var statusOrderCreated = models.OrderStatusCreated
var courierRating1 = 5.0
var courierRating2 = 4.0
var courierStatusOffline = models.CourierStatusOffline

// Экземпляры моделей приложения
// // Заказы
var order1 = &models.Order{
	ID:              orderID,
	CustomerName:    "test_name_1",
	CustomerPhone:   "71111111111",
	PickupAddress:   "pickup_location_1",
	DeliveryAddress: "delivery_location_1",
	Items: []models.OrderItem{
		{ID: itemID, OrderID: orderID, Name: "test_item_1", Quantity: 2, Price: 50.0},
	},
	TotalAmount:  float64(2) * 50.0,
	DeliveryCost: 4510.0,
	Status:       models.OrderStatusCreated,
	CreatedAt:    time.Now(),
	UpdatedAt:    time.Now(),
}
var order2 = &models.Order{
	ID:              uuid.New(),
	CustomerName:    "test_name_2",
	CustomerPhone:   "72222222222",
	PickupAddress:   "pickup_location_2",
	DeliveryAddress: "delivery_location_2",
	Items: []models.OrderItem{
		{ID: itemID, OrderID: orderID, Name: "test_item_1", Quantity: 2, Price: 50.0},
		{ID: uuid.New(), OrderID: orderID, Name: "test_item_2", Quantity: 1, Price: 1000.0},
	},
	TotalAmount:  float64(2)*50.0 + 1000.0,
	DeliveryCost: 123.0,
	Status:       models.OrderStatusDelivered,
	CourierID:    &courierID,
	CreatedAt:    time.Now(),
	UpdatedAt:    time.Now(),
	DeliveredAt:  &deliveredTime,
}
var order3 = &models.Order{
	ID:              uuid.New(),
	CustomerName:    "test_name_3",
	CustomerPhone:   "73333333333",
	PickupAddress:   "pickup_location_3",
	DeliveryAddress: "delivery_location_3",
	Items: []models.OrderItem{
		{ID: itemID, OrderID: orderID, Name: "test_item_1", Quantity: 2, Price: 50.0},
	},
	TotalAmount:  float64(2) * 50.0,
	DeliveryCost: 234.0,
	Status:       models.OrderStatusCreated,
	CreatedAt:    time.Now(),
	UpdatedAt:    time.Now(),
}

// // Курьеры
var courier1 = &models.Courier{
	ID:           courierID,
	Name:         "test_name_1",
	Phone:        "71111111111",
	Status:       models.CourierStatusAvailable,
	Rating:       &courierRating1,
	TotalReviews: 1,
	CreatedAt:    time.Now(),
	UpdatedAt:    time.Now(),
}
var courier2 = &models.Courier{
	ID:           uuid.New(),
	Name:         "test_name_2",
	Phone:        "72222222222",
	Status:       models.CourierStatusOffline,
	Rating:       &courierRating2,
	TotalReviews: 1,
	CreatedAt:    time.Now(),
	UpdatedAt:    time.Now(),
}
var courier3 = &models.Courier{
	ID:           uuid.New(),
	Name:         "test_name_3",
	Phone:        "73333333333",
	Status:       models.CourierStatusAvailable,
	Rating:       nil,
	TotalReviews: 0,
	CreatedAt:    time.Now(),
	UpdatedAt:    time.Now(),
}

// // Отзывы
var review = &models.Review{
	ID:        reviewID,
	OrderID:   order2.ID,
	CourierID: courier1.ID,
	Rating:    createReviewRequest.Rating,
	Text:      createReviewRequest.Text,
}

// // Запросы
var createOrderRequest = models.CreateOrderRequest{
	CustomerName:    order1.CustomerName,
	CustomerPhone:   order1.CustomerPhone,
	PickupAddress:   order1.PickupAddress,
	DeliveryAddress: order1.DeliveryAddress,
	Items: []models.CreateOrderItemRequest{
		{Name: order1.Items[0].Name, Quantity: order1.Items[0].Quantity, Price: order1.Items[0].Price},
	},
}
var updateOrderRequest = models.UpdateOrderStatusRequest{
	Status: models.OrderStatusAccepted, CourierID: &courierID,
}
var createReviewRequest = models.CreateReviewRequest{Rating: 4, Text: "text_review"}
var createCourierRequest = models.CreateCourierRequest{
	Name:  courier1.Name,
	Phone: courier1.Phone,
}
var updateCourierStatusRequest = models.UpdateCourierStatusRequest{Status: models.CourierStatusOffline}
var assignOrderRequest = assignOrderRequestType{OrderID: order1.ID}

// // Метрики
var kafkaMetrics = &models.KafkaMetricsResponse{
	TotalLag: 123,
	Statistics: []models.KafkaTopicMetricsResponse{
		{"test_topic_1", 1000, 20, "10 ms"},
		{"test_topic_2", 1500, 100, "20 ms"},
	},
}
var redisMetrics = &models.RedisMetricsResponse{
	HitRate:   95.0,
	MissRate:  5.0,
	CacheSize: 1000,
}

// Ошибки
var errorNotFound = errors.New("not found")
var errorInternalServerError = errors.New("internal Server Error")

// Модели
type assignOrderRequestType struct {
	OrderID uuid.UUID `json:"order_id"`
}

// Тесткейсы для /api/orders
var createOrderTestCases = []struct {
	name               string
	payload            *models.CreateOrderRequest
	returnedValue      *models.Order
	returnedError      error
	expectedStatusCode int
}{
	{
		"test_created",
		&createOrderRequest,
		order1,
		nil,
		http.StatusCreated,
	},
	{
		"test_bad_request",
		nil,
		nil,
		nil,
		http.StatusBadRequest,
	},
	{
		"test_server_error",
		&createOrderRequest,
		nil,
		errorInternalServerError,
		http.StatusInternalServerError,
	},
	{
		"validate_customer_name",
		&models.CreateOrderRequest{},
		nil,
		nil,
		http.StatusBadRequest,
	},
	{
		"validate_customer_phone",
		&models.CreateOrderRequest{
			CustomerName: "test_name",
		},
		nil,
		nil,
		http.StatusBadRequest,
	},
	{
		"validate_pickup_address",
		&models.CreateOrderRequest{
			CustomerName:  "test_name",
			CustomerPhone: "79999999999",
		},
		nil,
		nil,
		http.StatusBadRequest,
	},
	{
		"validate_delivery_address",
		&models.CreateOrderRequest{
			CustomerName:  "test_name",
			CustomerPhone: "79999999999",
			PickupAddress: "pickup_location",
		},
		nil,
		nil,
		http.StatusBadRequest,
	},
	{
		"validate_order_items_absent",
		&models.CreateOrderRequest{
			CustomerName:    "test_name",
			CustomerPhone:   "79999999999",
			PickupAddress:   "pickup_location",
			DeliveryAddress: "delivery_location",
		},
		nil,
		nil,
		http.StatusBadRequest,
	},
	{
		"validate_order_item_name",
		&models.CreateOrderRequest{
			CustomerName:    "test_name",
			CustomerPhone:   "79999999999",
			PickupAddress:   "pickup_location",
			DeliveryAddress: "delivery_location",
			Items:           []models.CreateOrderItemRequest{{Name: ""}},
		},
		nil,
		nil,
		http.StatusBadRequest,
	},
	{
		"validate_order_item_quantity",
		&models.CreateOrderRequest{
			CustomerName:    "test_name",
			CustomerPhone:   "79999999999",
			PickupAddress:   "pickup_location",
			DeliveryAddress: "delivery_location",
			Items:           []models.CreateOrderItemRequest{{Name: "test", Quantity: 0}},
		},
		nil,
		nil,
		http.StatusBadRequest,
	},
	{
		"validate_order_item_price",
		&models.CreateOrderRequest{
			CustomerName:    "test_name",
			CustomerPhone:   "79999999999",
			PickupAddress:   "pickup_location",
			DeliveryAddress: "delivery_location",
			Items:           []models.CreateOrderItemRequest{{Name: "test", Quantity: 2, Price: -1}},
		},
		nil,
		nil,
		http.StatusBadRequest,
	},
}

var getOrderTestCases = []struct {
	name               string
	id                 uuid.UUID
	returnedValue      *models.Order
	returnedError      error
	expectedStatusCode int
}{
	{"test_ok", orderID, order1, nil, http.StatusOK},
	{"test_not_found", uuid.New(), nil, errorNotFound, http.StatusNotFound},
	{"test_server_error", uuid.New(), nil, errorInternalServerError, http.StatusInternalServerError},
}

var updateOrderStatusTestCases = []struct {
	name               string
	id                 uuid.UUID
	payload            *models.UpdateOrderStatusRequest
	returnedError      error
	expectedStatusCode int
}{
	{"test_ok", orderID, &updateOrderRequest, nil, http.StatusOK},
	{"test_not_found", uuid.New(), &updateOrderRequest, errorNotFound, http.StatusNotFound},
	{"test_server_error", uuid.New(), &updateOrderRequest, errorInternalServerError, http.StatusInternalServerError},
}

var getOrdersTestCases = []struct {
	name               string
	status             *models.OrderStatus
	courierID          *uuid.UUID
	limit              int
	offset             int
	queryParam         string
	returnedValue      []*models.Order
	returnedError      error
	expectedStatusCode int
}{
	{
		"test_w/o_filters",
		nil,
		nil,
		50,
		0,
		"",
		[]*models.Order{order1, order2, order3},
		nil,
		http.StatusOK,
	},
	{
		"test_with_status_filter",
		&statusOrderCreated,
		nil,
		50,
		0,
		"status",
		[]*models.Order{order1, order3},
		nil,
		http.StatusOK,
	},
	{
		"test_with_courier_filter",
		nil,
		&courierID,
		50,
		0,
		"courier_id",
		[]*models.Order{order2},
		nil,
		http.StatusOK,
	},
	{
		"test_with_courier_filter_bad_request",
		nil,
		nil,
		100,
		0,
		"courier_id",
		nil,
		nil,
		http.StatusBadRequest,
	},
	{
		"test_with_limit_filter",
		nil,
		nil,
		1,
		0,
		"limit",
		[]*models.Order{order1},
		nil,
		http.StatusOK,
	},
	{
		"test_with_offset_filter",
		nil,
		nil,
		50,
		1,
		"offset",
		[]*models.Order{order2, order3},
		nil,
		http.StatusOK,
	},
	{
		"test_server_error",
		nil,
		nil,
		50,
		0,
		"",
		nil,
		errorInternalServerError,
		http.StatusInternalServerError,
	},
}

var createOrderReviewTestCases = []struct {
	name               string
	payload            *models.CreateReviewRequest
	order              *models.Order
	orderID            string
	returnedValue      *models.Review
	returnedError      error
	expectedStatusCode int
}{
	{
		"test_created",
		&createReviewRequest,
		order2,
		order2.ID.String(),
		review,
		nil,
		http.StatusCreated,
	},
	{
		"test_bad_request",
		nil,
		order2,
		order2.ID.String(),
		nil,
		nil,
		http.StatusBadRequest,
	},
	{
		"test_server_error",
		&createReviewRequest,
		order2,
		order2.ID.String(),
		nil,
		errorInternalServerError,
		http.StatusInternalServerError,
	},
	{
		"test_invalid_order_id",
		&createReviewRequest,
		nil,
		"1",
		nil,
		nil,
		http.StatusBadRequest,
	},
	{
		"validate_rating",
		&models.CreateReviewRequest{Rating: -1},
		order2,
		order2.ID.String(),
		nil,
		nil,
		http.StatusBadRequest,
	},
	{
		"validate_text",
		&models.CreateReviewRequest{Rating: 4, Text: strings.Repeat("s", 300)},
		order2,
		order2.ID.String(),
		nil,
		nil,
		http.StatusBadRequest,
	},
}

// Тесткейсы для /api/couriers
var createCourierTestCases = []struct {
	name               string
	payload            *models.CreateCourierRequest
	returnedValue      *models.Courier
	returnedError      error
	expectedStatusCode int
}{
	{
		"test_created",
		&createCourierRequest,
		courier1,
		nil,
		http.StatusCreated,
	},
	{
		"test_bad_request",
		nil,
		nil,
		nil,
		http.StatusBadRequest,
	},
	{
		"test_server_error",
		&createCourierRequest,
		nil,
		errorInternalServerError,
		http.StatusInternalServerError,
	},
	{
		"validate_courier_name",
		&models.CreateCourierRequest{Name: ""},
		nil,
		nil,
		http.StatusBadRequest,
	},
	{
		"validate_courier_phone",
		&models.CreateCourierRequest{Name: courier1.Name, Phone: ""},
		nil,
		nil,
		http.StatusBadRequest,
	},
}

var getCourierTestCases = []struct {
	name               string
	id                 uuid.UUID
	returnedValue      *models.Courier
	returnedError      error
	expectedStatusCode int
}{
	{"test_ok", courierID, courier1, nil, http.StatusOK},
	{"test_not_found", uuid.New(), nil, errorNotFound, http.StatusNotFound},
	{"test_server_error", uuid.New(), nil, errorInternalServerError, http.StatusInternalServerError},
}

var updateCourierStatusTestCases = []struct {
	name               string
	id                 uuid.UUID
	payload            *models.UpdateCourierStatusRequest
	returnedError      error
	expectedStatusCode int
}{
	{"test_ok", orderID, &updateCourierStatusRequest, nil, http.StatusOK},
	{"test_not_found", uuid.New(), &updateCourierStatusRequest, errorNotFound, http.StatusNotFound},
	{"test_server_error", uuid.New(), &updateCourierStatusRequest, errorInternalServerError, http.StatusInternalServerError},
}

var getCouriersTestCases = []struct {
	name               string
	status             *models.CourierStatus
	ratingSort         bool
	limit              int
	offset             int
	queryParam         string
	returnedValue      []*models.Courier
	returnedError      error
	expectedStatusCode int
}{
	{
		"test_w/o_filters",
		nil,
		false,
		50,
		0,
		"",
		[]*models.Courier{courier1, courier2, courier3},
		nil,
		http.StatusOK,
	},
	{
		"test_with_status_filter",
		&courierStatusOffline,
		false,
		50,
		0,
		"status",
		[]*models.Courier{courier2},
		nil,
		http.StatusOK,
	},
	{
		"test_with_rating_sort_filter",
		nil,
		true,
		50,
		0,
		"rating_sort",
		[]*models.Courier{courier1, courier2, courier3},
		nil,
		http.StatusOK,
	},
	{
		"test_with_limit_filter",
		nil,
		false,
		1,
		0,
		"limit",
		[]*models.Courier{courier1},
		nil,
		http.StatusOK,
	},
	{
		"test_with_offset_filter",
		nil,
		false,
		50,
		1,
		"offset",
		[]*models.Courier{courier2, courier3},
		nil,
		http.StatusOK,
	},
	{
		"test_server_error",
		nil,
		false,
		50,
		0,
		"",
		nil,
		errorInternalServerError,
		http.StatusInternalServerError,
	},
}

var getAvailableCouriersTestCases = []struct {
	name               string
	returnedValue      []*models.Courier
	returnedError      error
	expectedStatusCode int
}{
	{"test_ok", []*models.Courier{courier1, courier3}, nil, http.StatusOK},
	{"test_server_error", nil, errorInternalServerError, http.StatusInternalServerError},
}

var assignOrderToCourierTestCases = []struct {
	name               string
	payload            *assignOrderRequestType
	courierID          uuid.UUID
	returnedError      error
	expectedStatusCode int
}{
	{
		"test_ok",
		&assignOrderRequest,
		courier1.ID,
		nil,
		http.StatusOK,
	},
	{
		"test_courier_not_found",
		&assignOrderRequest,
		uuid.New(),
		errorNotFound,
		http.StatusNotFound,
	},
	{
		"test_courier_not_available",
		&assignOrderRequest,
		courier2.ID,
		errors.New("not available"),
		http.StatusBadRequest,
	},
	{
		"test_server_error",
		&assignOrderRequest,
		courier1.ID,
		errorInternalServerError,
		http.StatusInternalServerError,
	},
	{
		"test_bad_courier_id",
		nil,
		uuid.Nil,
		nil,
		http.StatusBadRequest,
	},
	{
		"test_bad_request",
		nil,
		courier1.ID,
		nil,
		http.StatusBadRequest,
	},
	{
		"test_bad_order_id",
		&assignOrderRequestType{OrderID: uuid.Nil},
		courier1.ID,
		nil,
		http.StatusBadRequest,
	},
}

var getCourierReviewsTestCases = []struct {
	name               string
	courierID          uuid.UUID
	returnedValue      []*models.Review
	returnedError      error
	expectedStatusCode int
}{
	{
		"test_ok",
		review.CourierID,
		[]*models.Review{review},
		nil,
		http.StatusOK,
	},
	{
		"test_server_error",
		courier2.ID,
		nil,
		errorInternalServerError,
		http.StatusInternalServerError,
	},
}

// Тесткейсы для /api/cache/metrics
var getRedisMetricsTestCases = []struct {
	name               string
	context            context.Context
	returnedValue      *models.RedisMetricsResponse
	returnedError      error
	expectedStatusCode int
}{
	{"test_ok", context.Background(), redisMetrics, nil, http.StatusOK},
	{"test_server_error", context.Background(), nil, errorInternalServerError, http.StatusInternalServerError},
}

// Тесткейсы для extractUUIDFromPath
var extractUUIDFromPathTestCases = []struct {
	name     string
	path     string
	prefix   string
	hasError bool
}{
	{"test_ok", fmt.Sprintf("/api/%s/status", uuid.New().String()), "/api/", false},
	{"test_invalid_path", "/test/case", "/api/", true},
	{"test_missing_uuid", "/api/status", "/api/", true},
	{"test_invalid_uuid", fmt.Sprintf("/api/%d/status", 1234), "/api/", true},
}
