-- +goose Up
-- +goose StatementBegin

-- Таблица для заявок в друзья
CREATE TABLE friend_requests
(
    id          UUID PRIMARY KEY,
    sender_id   UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    receiver_id UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status      VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'rejected')),
    created_at  TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP   NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_friend_request UNIQUE (sender_id, receiver_id),

    CONSTRAINT no_self_request CHECK (sender_id != receiver_id)
);

CREATE INDEX idx_friend_requests_sender ON friendship_requests (sender_id, status);
CREATE INDEX idx_friend_requests_receiver ON friendship_requests (receiver_id, status);
CREATE INDEX idx_friend_requests_created ON friendship_requests (created_at DESC);

CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_friend_requests_updated_at
    BEFORE UPDATE ON friendship_requests
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

COMMENT ON COLUMN friendship_requests.status IS 'pending/rejected';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_friend_requests_updated_at ON friendship_requests;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS friendship_requests CASCADE;
-- +goose StatementEnd