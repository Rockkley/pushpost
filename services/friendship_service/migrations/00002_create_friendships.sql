-- +goose Up
-- +goose StatementBegin

CREATE TABLE friendships
(
    id         UUID        PRIMARY KEY,
    user1_id   UUID        NOT NULL,
    user2_id   UUID        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT friendships_unique_pair
        UNIQUE (user1_id, user2_id),

    CONSTRAINT friendships_ordered_users
        CHECK (user1_id < user2_id)
);

CREATE INDEX idx_friendships_user1
    ON friendships (user1_id, created_at DESC);

CREATE INDEX idx_friendships_user2
    ON friendships (user2_id, created_at DESC);

CREATE INDEX idx_friendships_pair
    ON friendships (user1_id, user2_id);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS friendships CASCADE;
-- +goose StatementEnd