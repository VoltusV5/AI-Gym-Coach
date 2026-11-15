BEGIN;
DROP TRIGGER ensure_user_role ON sportapp.users;
DROP FUNCTION sportapp.ensure_user_role();
DROP TABLE sportapp.account_audit;
ALTER TABLE sportapp.users DROP COLUMN block_reason, DROP COLUMN blocked;
COMMIT;
