-- +goose Up
-- +goose StatementBegin
CREATE TABLE profiles
(
    user_id       UUID PRIMARY KEY,
    username      VARCHAR(30)  NOT NULL,
    display_name  VARCHAR(60),
    first_name    VARCHAR(60),
    last_name     VARCHAR(60),
    birth_date    DATE,
    avatar_url    TEXT,
    bio           VARCHAR(500),
    telegram_link VARCHAR(255),
    is_private    BOOLEAN      NOT NULL DEFAULT FALSE,
    version       BIGINT       NOT NULL DEFAULT 1,
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMP,

    CONSTRAINT profiles_username_length CHECK (char_length(username) BETWEEN 3 AND 30),
    CONSTRAINT profiles_display_name_length CHECK (display_name IS NULL OR char_length(display_name) BETWEEN 1 AND 60),
    CONSTRAINT profiles_first_name_length CHECK (first_name IS NULL OR char_length(first_name) BETWEEN 1 AND 60),
    CONSTRAINT profiles_last_name_length CHECK (last_name IS NULL OR char_length(last_name) BETWEEN 1 AND 60),
    CONSTRAINT profiles_bio_length CHECK (bio IS NULL OR char_length(bio) <= 500),
    CONSTRAINT profiles_birth_date_not_future CHECK (birth_date IS NULL OR birth_date <= CURRENT_DATE)
);

CREATE UNIQUE INDEX idx_profiles_username_unique ON profiles (LOWER(username)) WHERE deleted_at IS NULL;
CREATE INDEX idx_profiles_created_at ON profiles (created_at DESC);
CREATE INDEX idx_profiles_active ON profiles (user_id) WHERE deleted_at IS NULL;

CREATE OR REPLACE FUNCTION update_profiles_updated_at_column()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_profiles_updated_at
    BEFORE UPDATE ON profiles
    FOR EACH ROW
EXECUTE FUNCTION update_profiles_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_profiles_updated_at ON profiles;
DROP FUNCTION IF EXISTS update_profiles_updated_at_column();
DROP TABLE IF EXISTS profiles CASCADE;
-- +goose StatementEnd
