CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4()
    order_id UUID NOT NULL,
    courier_id UUID REFERENCES couriers(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    text VARCHAR(255),
);

ALTER TABLE couriers
ADD (
    rating DECIMAL(3, 2) CHECK (rating >= 1 AND rating <= 5),
    total_reviews INTEGER NOT NULL DEFAULT 0
)

CREATE INDEX idx_reviews_courier_id ON reviews(courier_id)