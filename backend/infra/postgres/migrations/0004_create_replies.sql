CREATE TABLE IF NOT EXISTS replies (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    thread_id TEXT NOT NULL CONSTRAINT replies_thread_id_fkey REFERENCES threads (id),
    user_id TEXT NOT NULL,
    body TEXT NOT NULL,
    -- Attaching an image to a reply is optional; one file belongs to at most one reply.
    file_id TEXT CONSTRAINT replies_file_id_fkey REFERENCES files (id) CONSTRAINT replies_file_id_key UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_replies_thread_id_created_at ON replies (thread_id, created_at DESC);
