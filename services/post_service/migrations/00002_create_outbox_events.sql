-- +goose Up
-- +goose StatementBegin

CREATE TABLE outbox_events
(
    id             UUID        PRIMARY KEY,
    aggregate_id   TEXT        NOT NULL,
    aggregate_type TEXT        NOT NULL,
    event_type     TEXT        NOT NULL,
    payload        JSONB       NOT NULL,
    status         TEXT        NOT NULL DEFAULT 'pending',
    attempts       INT         NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT outbox_status_check
        CHECK (status IN ('pending', 'processing', 'processed'))
);

CREATE INDEX idx_outbox_pending
    ON outbox_events (created_at ASC)
    WHERE status = 'pending';

CREATE INDEX idx_outbox_processing
    ON outbox_events (updated_at ASC)
    WHERE status = 'processing';

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS outbox_events;
-- +goose StatementEnd