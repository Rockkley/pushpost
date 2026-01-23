-- +goose Up
-- +goose StatementBegin
CREATE TABLE users
(
    id            UUID PRIMARY KEY,
    username      VARCHAR(30)  NOT NULL,
    email         VARCHAR(255) NOT NULL,
    password_hash CHAR(60)     NOT NULL,
    created_at    TIMESTAMP    NOT NULL,
    deleted_at    TIMESTAMP,

    CONSTRAINT username_length CHECK (char_length(username) >= 3),
    CONSTRAINT email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

CREATE UNIQUE INDEX idx_users_username_unique ON users (LOWER(username)) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_users_email_unique ON users (LOWER(email)) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at ON users (created_at DESC);
CREATE INDEX idx_users_active ON users (id) WHERE deleted_at IS NULL;

COMMENT ON COLUMN users.username IS 'display name, case-insensitive unique';
COMMENT ON COLUMN users.email IS 'case-insensitive unique';
COMMENT ON COLUMN users.password_hash IS '60 chars';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users CASCADE;
-- +goose StatementEnd