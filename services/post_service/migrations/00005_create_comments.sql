-- +goose Up
-- +goose StatementBegin
CREATE TABLE comments
(
    id               UUID PRIMARY KEY,
    post_id          UUID        NOT NULL REFERENCES posts (id) ON DELETE CASCADE,
    author_id        UUID        NOT NULL,
    content          TEXT        NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    parent_id        UUID        NULL REFERENCES comments (id) ON DELETE CASCADE,
    reply_to_user_id UUID        NULL,
    upvotes_count    INT         NOT NULL DEFAULT 0,
    downvotes_count  INT         NOT NULL DEFAULT 0


CONSTRAINT comment_content_not_empty CHECK (char_length(content) >= 1),
    CONSTRAINT comment_content_max_length CHECK (char_length(content) <= 2000)
);
CREATE INDEX idx_comments_parent ON comments (parent_id, created_at ASC, id ASC);
CREATE INDEX idx_comments_post_created ON comments (post_id, created_at ASC, id ASC);

CREATE OR REPLACE FUNCTION update_comments_updated_at()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_comments_updated_at
    BEFORE UPDATE
    ON comments
    FOR EACH ROW
EXECUTE FUNCTION update_comments_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_comments_updated_at ON comments;
DROP FUNCTION IF EXISTS update_comments_updated_at();
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS comment_votes;
DROP INDEX IF EXISTS idx_comments_parent;

-- +goose StatementEnd
