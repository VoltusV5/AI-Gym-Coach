BEGIN;

-- The v1 reader retains its integer projection; v2 reads the full value.
ALTER TABLE sportapp.profile
    ALTER COLUMN weight_kg TYPE NUMERIC USING weight_kg::numeric,
    ADD COLUMN session_duration_min INTEGER CHECK (session_duration_min > 0),
    ADD CONSTRAINT profile_weight_v2_compat CHECK (weight_kg BETWEEN -2147483648 AND 2147483647);

CREATE TABLE sportapp.profile_excluded_exercises (
    user_id INTEGER NOT NULL REFERENCES sportapp.profile(user_id) ON DELETE CASCADE,
    exercise_id INTEGER NOT NULL REFERENCES sportapp.exercises(id),
    PRIMARY KEY (user_id, exercise_id)
);

COMMIT;
