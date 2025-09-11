-- Thread table
-- Stores thread information and scraping status

CREATE TABLE IF NOT EXISTS thread (
    id                        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    summary                   TEXT NOT NULL,
    cid                       TEXT NOT NULL,
    num_tweets               INTEGER NOT NULL DEFAULT 0,
    
    -- Thread status tracking
    status                   thread_status NOT NULL DEFAULT 'pending',
    
    -- Retry tracking for cron jobs (internal use only)
    retry_count              INTEGER NOT NULL DEFAULT 0,
    
    -- Optimistic locking
    version                  INTEGER NOT NULL DEFAULT 1,
    
    -- Thread author information (centralized here instead of duplicating in mentions)
    author_id                TEXT,
    author_name              TEXT,
    author_screen_name       TEXT,
    author_profile_image_url TEXT,
    
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add updated_at trigger
CREATE OR REPLACE TRIGGER set_thread_updated_at
    BEFORE UPDATE ON thread
    FOR EACH ROW
    EXECUTE FUNCTION moddatetime('updated_at');

-- Indexes
CREATE INDEX IF NOT EXISTS idx_thread_status ON thread(status);
CREATE INDEX IF NOT EXISTS idx_thread_author_id ON thread(author_id);
CREATE INDEX IF NOT EXISTS idx_thread_created_at ON thread(created_at);
CREATE INDEX IF NOT EXISTS idx_thread_updated_at ON thread(updated_at);
CREATE INDEX IF NOT EXISTS idx_thread_retry_count ON thread(retry_count);
