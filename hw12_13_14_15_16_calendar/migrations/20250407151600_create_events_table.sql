-- +goose Up
-- +goose StatementBegin
CREATE TABLE events (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    date_time TIMESTAMP WITH TIME ZONE NOT NULL,
    duration BIGINT NOT NULL, -- stored as nanoseconds
    description TEXT,
    user_id VARCHAR(36) NOT NULL,
    notify_before BIGINT, -- stored as nanoseconds, nullable
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_events_date_time ON events(date_time);
CREATE INDEX idx_events_user_id ON events(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS events;
-- +goose StatementEnd