-- +goose Up
-- +goose StatementBegin
CREATE TABLE posts
(
    id         UUID        PRIMARY KEY,
    author_id  UUID        NOT NULL,
    content    TEXT        NOT NULL,
    version    INT         NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    

    CONSTRAINT post_content_not_empty  CHECK (char_length(content) >= 1),
    CONSTRAINT post_content_max_length CHECK (char_length(content) <= 5000)
);

CREATE INDEX idx_posts_author_created ON posts (author_id, created_at DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_posts_created ON posts (created_at DESC)
    WHERE deleted_at IS NULL;

CREATE OR REPLACE FUNCTION update_posts_updated_at()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_posts_updated_at
    BEFORE UPDATE ON posts
    FOR EACH ROW
EXECUTE FUNCTION update_posts_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER  IF EXISTS trg_posts_updated_at ON posts;
DROP FUNCTION IF EXISTS update_posts_updated_at();
DROP TABLE    IF EXISTS posts CASCADE;
-- +goose StatementEnd