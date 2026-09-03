CREATE TABLE IF NOT EXISTS threads (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    body TEXT NOT NULL,
    image_path TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION set_threads_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS threads_set_updated_at ON threads;
CREATE TRIGGER threads_set_updated_at
    BEFORE UPDATE ON threads
    FOR EACH ROW
    EXECUTE FUNCTION set_threads_updated_at();
