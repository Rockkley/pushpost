-- +goose Up
-- +goose StatementBegin
CREATE TABLE relationships
(
    id         UUID PRIMARY KEY,
    user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    friend_id  UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status     VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'accepted')),
    created_at TIMESTAMP   NOT NULL,
    updated_at TIMESTAMP   NOT NULL,

    CONSTRAINT unique_relationship UNIQUE (user_id, friend_id),
    CONSTRAINT no_self_friendship CHECK (user_id != friend_id)
);

CREATE INDEX idx_relationships_user_status ON relationships (user_id, status);
CREATE INDEX idx_relationships_friend_status ON relationships (friend_id, status);

CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_relationships_updated_at
    BEFORE UPDATE
    ON relationships
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_relationships_updated_at ON relationships;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS relationships CASCADE;
-- +goose StatementEnd