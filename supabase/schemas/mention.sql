-- Mention table
-- Stores user mentions and their associated threads

CREATE TABLE IF NOT EXISTS mention (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id           TEXT NOT NULL,
    thread_id         UUID NOT NULL REFERENCES thread(id) 
                          ON UPDATE RESTRICT ON DELETE RESTRICT,
    mention_create_at TIMESTAMPTZ NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(user_id, thread_id)
);

-- Add updated_at trigger
CREATE OR REPLACE TRIGGER set_mention_updated_at
    BEFORE UPDATE ON mention
    FOR EACH ROW
    EXECUTE FUNCTION moddatetime('updated_at');

-- Indexes
CREATE INDEX IF NOT EXISTS idx_mention_user_id ON mention(user_id);
CREATE INDEX IF NOT EXISTS idx_mention_thread_id ON mention(thread_id);
CREATE INDEX IF NOT EXISTS idx_mention_created_at ON mention(created_at);
CREATE INDEX IF NOT EXISTS idx_mention_user_thread ON mention(user_id, thread_id);
