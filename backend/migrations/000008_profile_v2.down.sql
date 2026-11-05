BEGIN;

-- Never silently round weights introduced through v2 during rollback.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM sportapp.profile
               WHERE weight_kg <> trunc(weight_kg)
                  OR weight_kg NOT BETWEEN -2147483648 AND 2147483647) THEN
        RAISE EXCEPTION 'Profile weights cannot be represented by v1 integer storage'
            USING HINT = 'Export fractional weights and explicitly resolve them before rollback';
    END IF;
END $$;

DROP TABLE sportapp.profile_excluded_exercises;
ALTER TABLE sportapp.profile
    DROP COLUMN session_duration_min,
    DROP CONSTRAINT profile_weight_v2_compat,
    ALTER COLUMN weight_kg TYPE INTEGER USING weight_kg::integer;

COMMIT;
