ALTER TABLE orders
ADD (
    pickup_address TEXT NOT NULL,
    delivery_cost DECIMAL(10, 2) NOT NULL
);