package services

import (
	"context"
	"delivery-system/internal/models"

	"github.com/google/uuid"
)

type GeolocationServiceInterface interface {
	GetCoordinates(address string) (float64, float64, error)
	MakeRoute(coordinates [][2]float64) (float64, error)
	CacheResults(coordinates [][2]float64, distance float64, order *models.Order)
}

type OrderServiceInterface interface {
	CreateOrder(req *models.CreateOrderRequest) (*models.Order, error)
	GetOrder(orderID uuid.UUID) (*models.Order, error)
	UpdateOrderStatus(orderID uuid.UUID, req *models.UpdateOrderStatusRequest) error
	GetOrders(status *models.OrderStatus, courierID *uuid.UUID, limit, offset int) ([]*models.Order, error)
}

type ReviewServiceInterface interface {
	CreateReview(req *models.CreateReviewRequest, order *models.Order) (*models.Review, error)
	GetReviews(courierID uuid.UUID) ([]*models.Review, error)
	RecalculateRating(courierID uuid.UUID) error
}

type CourierServiceInterface interface {
	CreateCourier(req *models.CreateCourierRequest) (*models.Courier, error)
	GetCourier(courierID uuid.UUID) (*models.Courier, error)
	UpdateCourierStatus(courierID uuid.UUID, req *models.UpdateCourierStatusRequest) error
	GetCouriers(status *models.CourierStatus, limit, offset int, ratingSort bool) ([]*models.Courier, error)
	GetAvailableCouriers() ([]*models.Courier, error)
	AssignOrderToCourier(orderID, courierID uuid.UUID) error
}

type KafkaMetricsServiceInterface interface {
	GetStatistics() *models.KafkaMetricsResponse
}

type RedisServiceInterface interface {
	GetStatistics(ctx context.Context) (*models.RedisMetricsResponse, error)
}
