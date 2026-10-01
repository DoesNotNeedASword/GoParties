CREATE TABLE IF NOT EXISTS parties (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    promises TEXT NOT NULL,
    image TEXT NOT NULL DEFAULT '',
    is_allowed BOOLEAN NOT NULL DEFAULT TRUE,
    ban_reason TEXT DEFAULT NULL,
    CONSTRAINT parties_ban_reason_required_when_banned
        CHECK (is_allowed OR ban_reason IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS idx_parties_name ON parties (name);

