-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS subscriptions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_name TEXT NOT NULL,
    price        INTEGER NOT NULL CHECK (price >= 0),
    user_id      UUID NOT NULL,
    start_date   DATE NOT NULL, -- Store as YYYY-MM-01
    end_date     DATE,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT check_dates_order CHECK (end_date IS NULL OR start_date <= end_date)
);

-- CREATE INDEX idx_subscriptions_user_service ON subscriptions(user_id, service_name);
-- CREATE INDEX idx_subscriptions_dates ON subscriptions(start_date, end_date);

CREATE INDEX idx_subs_perf_user_service 
    ON subscriptions (user_id, service_name, start_date, end_date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS subscriptions;
-- +goose StatementEnd
