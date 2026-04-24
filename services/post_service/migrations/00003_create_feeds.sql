-- +goose Up
-- +goose StatementBegin
CREATE TABLE feeds
(
    user_id     UUID        NOT NULL,
    post_id     UUID        NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    inserted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, post_id)
);

CREATE INDEX idx_feeds_user_inserted
    ON feeds (user_id, inserted_at DESC, post_id DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS feeds;
-- +goose StatementEnd