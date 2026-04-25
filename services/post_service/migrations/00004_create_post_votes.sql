-- +goose Up
-- +goose StatementBegin
CREATE TABLE post_votes
(
    post_id    UUID        NOT NULL REFERENCES posts (id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL,
    value      SMALLINT    NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT post_votes_value_check CHECK (value IN (-1, 1)),
    PRIMARY KEY (post_id, user_id)
);

CREATE INDEX idx_post_votes_user_id ON post_votes (user_id);

CREATE OR REPLACE FUNCTION update_post_votes_updated_at()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_post_votes_updated_at
    BEFORE UPDATE ON post_votes
    FOR EACH ROW
EXECUTE FUNCTION update_post_votes_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER  IF EXISTS trg_post_votes_updated_at ON post_votes;
DROP FUNCTION IF EXISTS update_post_votes_updated_at();
DROP TABLE    IF EXISTS post_votes;


-- +goose StatementEnd
