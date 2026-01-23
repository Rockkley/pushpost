-- +goose Up
-- +goose StatementBegin
CREATE TABLE messages
(
    id          UUID PRIMARY KEY,
    sender_id   UUID      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    receiver_id UUID      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    content     TEXT      NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    read_at     TIMESTAMP,

    CONSTRAINT no_self_message CHECK (sender_id != receiver_id
) ,
    CONSTRAINT content_not_empty CHECK (char_length(content) >= 1),
    CONSTRAINT content_max_length CHECK (char_length(content) <= 10000)
);

CREATE INDEX idx_messages_conversation ON messages (sender_id, receiver_id, created_at DESC);
CREATE INDEX idx_messages_conversation_reverse ON messages (receiver_id, sender_id, created_at DESC);
CREATE INDEX idx_messages_unread ON messages (receiver_id, created_at DESC) WHERE read_at IS NULL;


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS messages CASCADE;
-- +goose StatementEnd