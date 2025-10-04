-- +goose Up
CREATE TABLE subscription (
                       id BIGSERIAL PRIMARY KEY,
                       user_id UUID,
                       start_date DATE,
                       price INTEGER,
                       service_name VARCHAR(64),
                       created_at TIMESTAMP DEFAULT NOW()
);

-- +goose Down
DROP TABLE subscription;