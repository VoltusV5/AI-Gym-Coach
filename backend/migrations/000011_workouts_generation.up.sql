BEGIN;
CREATE TABLE sportapp.operation_results (
 user_id INTEGER NOT NULL REFERENCES sportapp.users(id),
 operation TEXT NOT NULL,
 key TEXT NOT NULL,
 request_hash TEXT NOT NULL,
 response JSONB NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 PRIMARY KEY(user_id,operation,key)
);
CREATE TABLE sportapp.workout_sessions (
 id UUID PRIMARY KEY,
 user_id INTEGER NOT NULL REFERENCES sportapp.users(id),
 client_session_id UUID NOT NULL,
 plan_id UUID NOT NULL REFERENCES sportapp.training_plans(id),
 revision_id UUID NOT NULL REFERENCES sportapp.plan_revisions(id),
 source_program_version_id UUID REFERENCES sportapp.program_versions(id),
 started_at TIMESTAMPTZ NOT NULL,
 finished_at TIMESTAMPTZ NOT NULL CHECK(finished_at>=started_at),
 request_hash TEXT NOT NULL,
 payload JSONB NOT NULL,
 result JSONB NOT NULL,
 UNIQUE(user_id,client_session_id)
);
CREATE INDEX workout_sessions_owner_time_idx ON sportapp.workout_sessions(user_id,finished_at DESC,id DESC);
CREATE TABLE sportapp.plan_generations (
 id UUID PRIMARY KEY,
 user_id INTEGER NOT NULL REFERENCES sportapp.users(id),
 plan_id UUID NOT NULL REFERENCES sportapp.training_plans(id),
 profile_version BIGINT NOT NULL,
 request_id TEXT NOT NULL,
 model_version TEXT NOT NULL,
 feature_schema_version TEXT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMIT;
