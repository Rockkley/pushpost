-- +goose Up
-- +goose StatementBegin
CREATE TABLE outbox_events
(
    id             UUID PRIMARY KEY,
    aggregate_type TEXT      NOT NULL,
    aggregate_id   TEXT      NOT NULL,
    event_type     TEXT      NOT NULL,
    payload        JSONB     NOT NULL,
    status         TEXT      NOT NULL DEFAULT 'pending',
    created_at     TIMESTAMP NOT NULL DEFAULT now(),
    processed_at   TIMESTAMP
);

CREATE INDEX idx_outbox_status_created
    ON outbox_events (status, created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS outbox_events CASCADE;

-- +goose StatementEnd