package kafka

import (
	"delivery-system/internal/models"

	"github.com/google/uuid"
)

type ProducerInterface interface {
	Close() error
	PublishOrderCreated(order *models.Order) error
	PublishOrderStatusChanged(orderID uuid.UUID, oldStatus, newStatus models.OrderStatus, courierID *uuid.UUID) error
	PublishCourierAssigned(orderID, courierID uuid.UUID) error
	PublishCourierStatusChanged(courierID uuid.UUID, oldStatus, newStatus models.CourierStatus) error
	PublishLocationUpdated(courierID uuid.UUID, lat, lon float64) error
}
