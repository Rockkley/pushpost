-- +goose Up
-- +goose StatementBegin
CREATE TABLE telegram_bindings
(
    user_id UUID PRIMARY KEY,
    chat_id BIGINT NOT NULL UNIQUE,
    username TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS telegram_bindings;
-- +goose StatementEnd
