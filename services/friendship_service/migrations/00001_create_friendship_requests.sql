-- +goose Up
-- +goose StatementBegin

CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE friendship_requests
(
    id          UUID        PRIMARY KEY,
    sender_id   UUID        NOT NULL,
    receiver_id UUID        NOT NULL,
    status      VARCHAR(20) NOT NULL
        CHECK (status IN ('pending', 'accepted', 'rejected', 'cancelled')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT friendship_requests_no_self
        CHECK (sender_id != receiver_id)
);

CREATE UNIQUE INDEX idx_friendship_requests_one_pending
    ON friendship_requests (sender_id, receiver_id)
    WHERE status = 'pending';

CREATE INDEX idx_friendship_requests_receiver_pending
    ON friendship_requests (receiver_id, created_at DESC)
    WHERE status = 'pending';

CREATE INDEX idx_friendship_requests_sender_pending
    ON friendship_requests (sender_id, created_at DESC)
    WHERE status = 'pending';

CREATE INDEX idx_friendship_requests_cooldown
    ON friendship_requests (sender_id, receiver_id, updated_at DESC)
    WHERE status IN ('rejected', 'cancelled');

CREATE TRIGGER trg_friendship_requests_updated_at
    BEFORE UPDATE ON friendship_requests
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TRIGGER  IF EXISTS trg_friendship_requests_updated_at ON friendship_requests;
DROP TABLE    IF EXISTS friendship_requests CASCADE;
DROP FUNCTION IF EXISTS update_updated_at_column();
-- +goose StatementEnd