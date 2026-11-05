BEGIN;
ALTER TABLE sportapp.programs DROP COLUMN version;
DROP TABLE sportapp.publication_audit;
DROP TABLE sportapp.trainer_verifications;
ALTER TABLE sportapp.trainer_profiles DROP COLUMN specializations;
COMMIT;
