-- +goose Up
-- +goose StatementBegin

-- Таблица для подтверждённой дружбы
CREATE TABLE friendships
(
    id         UUID PRIMARY KEY,
    user1_id   UUID      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    user2_id   UUID      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_friendship UNIQUE (user1_id, user2_id),

    CONSTRAINT ordered_users CHECK (user1_id < user2_id)
);

CREATE INDEX idx_friendships_user1 ON friendships (user1_id);
CREATE INDEX idx_friendships_user2 ON friendships (user2_id);
CREATE INDEX idx_friendships_created ON friendships (created_at DESC);

CREATE INDEX idx_friendships_users ON friendships (user1_id, user2_id);

CREATE OR REPLACE FUNCTION cleanup_friend_request_on_friendship()
    RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM friendship_requests
    WHERE (sender_id = NEW.user1_id AND receiver_id = NEW.user2_id)
       OR (sender_id = NEW.user2_id AND receiver_id = NEW.user1_id);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cleanup_request_after_friendship
    AFTER INSERT ON friendships
    FOR EACH ROW
EXECUTE FUNCTION cleanup_friend_request_on_friendship();


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS cleanup_request_after_friendship ON friendships;
DROP FUNCTION IF EXISTS cleanup_friend_request_on_friendship();
DROP TABLE IF EXISTS friendships CASCADE;
-- +goose StatementEnd