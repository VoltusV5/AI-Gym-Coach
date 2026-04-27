DROP INDEX IF EXISTS sportapp.idx_notes_user_id;

ALTER TABLE sportapp.notes
    ALTER COLUMN created_at DROP NOT NULL,
    ALTER COLUMN created_at DROP DEFAULT,
    ALTER COLUMN updated_at DROP NOT NULL,
    ALTER COLUMN updated_at DROP DEFAULT;

ALTER TABLE sportapp.notes
    ALTER COLUMN body TYPE VARCHAR;

ALTER TABLE sportapp.notes
    ADD CONSTRAINT notes_user_id_key UNIQUE (user_id);
