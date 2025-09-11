-- BotCookie table
-- Stores Twitter bot session cookies in database

CREATE TABLE IF NOT EXISTS bot_cookie (
    id           SERIAL PRIMARY KEY,
    email        TEXT NOT NULL,
    username     TEXT NOT NULL,
    cookies_data JSONB,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,
    
    UNIQUE(email, username)
);

-- Add updated_at trigger
CREATE OR REPLACE TRIGGER set_bot_cookie_updated_at
    BEFORE UPDATE ON bot_cookie
    FOR EACH ROW
    EXECUTE FUNCTION moddatetime('updated_at');

-- Indexes
CREATE INDEX IF NOT EXISTS idx_bot_cookie_email_username ON bot_cookie(email, username);
CREATE INDEX IF NOT EXISTS idx_bot_cookie_deleted_at ON bot_cookie(deleted_at) WHERE deleted_at IS NOT NULL;
