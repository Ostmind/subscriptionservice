-- +goose Up
ALTER TABLE subscription
    ADD CONSTRAINT subscription_constrain UNIQUE (user_id, start_date, price, service_name);


-- +goose Down
ALTER TABLE subscription DROP CONSTRAINT subscription_constrain;
