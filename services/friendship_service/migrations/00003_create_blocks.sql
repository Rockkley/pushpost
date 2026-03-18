-- +goose Up
CREATE TABLE blocks (
    user_id     UUID NOT NULL,
    target_id   UUID NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, target_id)
);

ALTER TABLE blocks
    ADD CONSTRAINT chk_block_not_self CHECK (user_id <> target_id);

CREATE INDEX idx_blocks_user_id
    ON blocks(user_id);

CREATE INDEX idx_blocks_target_id
    ON blocks(target_id);

-- +goose Down
DROP TABLE IF EXISTS blocks;