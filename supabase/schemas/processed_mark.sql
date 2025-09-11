-- ProcessedMark table
-- Tracks processed business actions to avoid duplicate responses

CREATE TABLE IF NOT EXISTS processed_mark (
    key        TEXT NOT NULL,
    type       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (key, type)
);

-- Add updated_at trigger
CREATE OR REPLACE TRIGGER set_processed_mark_updated_at
    BEFORE UPDATE ON processed_mark
    FOR EACH ROW
    EXECUTE FUNCTION moddatetime('updated_at');

-- Indexes
CREATE INDEX IF NOT EXISTS idx_processed_mark_type ON processed_mark(type);
CREATE INDEX IF NOT EXISTS idx_processed_mark_created_at ON processed_mark(created_at);
