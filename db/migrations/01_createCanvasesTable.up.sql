CREATE TABLE IF NOT EXISTS canvases (
    canvas_id     UUID PRIMARY KEY,
    name          TEXT NOT NULL,
    width         INTEGER NOT NULL,
    height        INTEGER NOT NULL,
    owner_id      UUID NOT NULL,
    members_ids   UUID[] DEFAULT '{}',
    privacy       TEXT CHECK (privacy IN ('public', 'private', 'friends')),
    image         VARCHAR(256) DEFAULT '',
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);