CREATE TABLE users (
    id          SERIAL PRIMARY KEY,
    provider    TEXT NOT NULL,
    provider_id TEXT NOT NULL,
    email       TEXT NOT NULL,
    name        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider, provider_id)
);

ALTER TABLE tasks
    ADD COLUMN user_id INT REFERENCES users(id) ON DELETE CASCADE;