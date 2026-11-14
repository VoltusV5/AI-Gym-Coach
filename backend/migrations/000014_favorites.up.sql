BEGIN;
CREATE TABLE sportapp.program_favorites (
 user_id INTEGER NOT NULL REFERENCES sportapp.users(id),
 program_id UUID NOT NULL REFERENCES sportapp.programs(id),
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 PRIMARY KEY(user_id,program_id)
);
COMMIT;
