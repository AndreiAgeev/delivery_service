package services

import (
	"fmt"

	"delivery-system/internal/database"
	"delivery-system/internal/logger"
	"delivery-system/internal/models"

	"github.com/google/uuid"
)

// ReviewService - сервис для работы с отзывами
type ReviewService struct {
	db  *database.DB
	log *logger.Logger
}

// NewReviewService создаёт экземпляр объекта ReviewService
func NewReviewService(db *database.DB, log *logger.Logger) *ReviewService {
	return &ReviewService{
		db:  db,
		log: log,
	}
}

// CreateReview создаёт новый отзыв на курьера
func (s *ReviewService) CreateReview(req *models.CreateReviewRequest, order *models.Order) (*models.Review, error) {
	review := &models.Review{
		ID:        uuid.New(),
		OrderID:   order.ID,
		CourierID: *order.CourierID,
		Rating:    req.Rating,
		Text:      req.Text,
	}

	query := `
		INSERT INTO reviews(id, order_id, courier_id, rating, text)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := s.db.Exec(query, review.ID, order.ID, *order.CourierID, req.Rating, req.Text)
	if err != nil {
		s.log.WithError(err).Error("Failed to create review")
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	s.log.WithFields(map[string]interface{}{
		"id":         review.ID,
		"order_id":   order.ID,
		"courier_id": *order.CourierID,
		"rating":     req.Rating,
		"text":       req.Text,
	}).Info("Review created successfully")

	return review, nil
}

// GetReviews возвращает список отзывов на курьера
func (s *ReviewService) GetReviews(courierID uuid.UUID) ([]*models.Review, error) {
	query := `
		SELECT id, order_id, courier_id, rating, text
		FROM reviews WHERE courier_id = $1
	`
	rows, err := s.db.Query(query, courierID)
	if err != nil {
		s.log.WithError(err).Error("Failed to get courier")
		return nil, fmt.Errorf("failed to get courier: %w", err)
	}
	defer rows.Close()

	var reviews []*models.Review
	for rows.Next() {
		var review models.Review
		if err = rows.Scan(&review.ID, &review.OrderID, &review.CourierID, &review.Rating, &review.Text); err != nil {
			s.log.WithError(err).Error("Failed to scan reviews")
			return nil, fmt.Errorf("failed to scan reviews: %w", err)
		}
		reviews = append(reviews, &review)
	}

	return reviews, nil
}

// RecalculateRating расчитывает рейтинг курьера в SQL
func (s *ReviewService) RecalculateRating(courierID uuid.UUID) error {
	query := `
		UPDATE couriers SET
		    rating = (SELECT AVG(rating) FROM reviews WHERE courier_id = couriers.id),
		    total_reviews = (SELECT COUNT(*) FROM reviews WHERE courier_id = couriers.id)
		WHERE couriers.id = $1
	`
	result, err := s.db.Exec(query, courierID)
	if err != nil {
		return fmt.Errorf("failed to update courier rating: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("courier not found")
	}

	s.log.Info("Courier rating updated")

	return nil
}
