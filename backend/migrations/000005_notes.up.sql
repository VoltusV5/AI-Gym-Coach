CREATE TABLE sportapp.notes (
    id SERIAL PRIMARY KEY,
    version BIGINT NOT NULL DEFAULT 1,
    user_id INTEGER NOT NULL,
    title VARCHAR(120),
    body TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    FOREIGN KEY (user_id) REFERENCES sportapp.users(id) ON DELETE CASCADE
);

CREATE INDEX idx_notes_user_id ON sportapp.notes(user_id);
