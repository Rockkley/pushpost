-- +goose Up
-- +goose StatementBegin
CREATE TABLE comment_votes
(
    comment_id UUID NOT NULL REFERENCES comments (id) ON DELETE CASCADE,
    user_id    UUID NOT NULL,
    value      SMALLINT NOT NULL CHECK (value IN (-1, 1)),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (comment_id, user_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS comment_votes
-- +goose StatementEnd;