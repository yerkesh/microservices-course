-- +goose Up
CREATE TABLE payments (
    transaction_uuid UUID PRIMARY KEY,
    order_uuid UUID NOT NULL,
    payment_method VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS payments;
