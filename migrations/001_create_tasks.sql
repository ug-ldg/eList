CREATE TABLE tasks (
    id         SERIAL      PRIMARY KEY,
    title      TEXT        NOT NULL,
    parent_id  INTEGER     REFERENCES tasks(id) ON DELETE CASCADE,
    status     TEXT        NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
