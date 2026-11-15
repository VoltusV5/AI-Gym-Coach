BEGIN;
ALTER TABLE sportapp.exercises DROP COLUMN admin_modified;
ALTER TABLE sportapp.exercises DROP COLUMN seed_key;
COMMIT;
