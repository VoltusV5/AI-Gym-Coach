BEGIN;

DROP TABLE sportapp.entitlements;
DROP TABLE sportapp.program_exercises;
DROP TABLE sportapp.program_days;
DROP TABLE sportapp.program_version_categories;
DROP TABLE sportapp.program_version_equipment;
ALTER TABLE sportapp.programs DROP CONSTRAINT programs_published_version_fk;
DROP TABLE sportapp.program_versions;
DROP TABLE sportapp.programs;
DROP TABLE sportapp.program_categories;
DROP TABLE sportapp.trainer_profiles;
DROP TABLE sportapp.profile_equipment;
DROP TABLE sportapp.exercise_equipment;
ALTER TABLE sportapp.exercises DROP COLUMN is_active;
DROP TABLE sportapp.equipment;
DROP TABLE sportapp.user_roles;

COMMIT;
