ALTER TABLE sportapp.notes
    DROP CONSTRAINT IF EXISTS notes_user_id_key;

ALTER TABLE sportapp.notes
    ALTER COLUMN body TYPE TEXT;

UPDATE sportapp.notes
SET created_at = NOW()
WHERE created_at IS NULL;

UPDATE sportapp.notes
SET updated_at = NOW()
WHERE updated_at IS NULL;

ALTER TABLE sportapp.notes
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN created_at SET NOT NULL,
    ALTER COLUMN updated_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_notes_user_id ON sportapp.notes(user_id);
