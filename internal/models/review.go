package models

import "github.com/google/uuid"

type Review struct {
	ID        uuid.UUID `json:"id"`
	OrderID   uuid.UUID `json:"order_id"`
	CourierID uuid.UUID `json:"courier_id"`
	Rating    int       `json:"rating"`
	Text      string    `json:"text"`
}

type CreateReviewRequest struct {
	Rating int    `json:"rating"`
	Text   string `json:"text"`
}
