BEGIN;

CREATE TABLE sportapp.training_plans (
    id UUID PRIMARY KEY,
    owner_id INTEGER NOT NULL REFERENCES sportapp.users(id),
    title TEXT NOT NULL CHECK (btrim(title) <> ''),
    version BIGINT NOT NULL DEFAULT 1 CHECK (version > 0),
    source_type TEXT NOT NULL CHECK (source_type IN ('manual', 'ai', 'fork', 'legacy_ai')),
    source_program_version_id UUID REFERENCES sportapp.program_versions(id),
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    archived_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (archived_at IS NULL OR NOT is_active)
);

CREATE UNIQUE INDEX training_plans_active_owner_uniq
    ON sportapp.training_plans(owner_id) WHERE is_active AND archived_at IS NULL;
CREATE INDEX training_plans_owner_page_idx
    ON sportapp.training_plans(owner_id, created_at DESC, id DESC) WHERE archived_at IS NULL;

CREATE TABLE sportapp.plan_days (
    id UUID PRIMARY KEY,
    plan_id UUID NOT NULL REFERENCES sportapp.training_plans(id),
    title TEXT NOT NULL CHECK (btrim(title) <> ''),
    position INTEGER NOT NULL CHECK (position >= 0),
    UNIQUE (plan_id, position),
    UNIQUE (plan_id, id)
);

CREATE TABLE sportapp.plan_exercises (
    id UUID PRIMARY KEY,
    plan_day_id UUID NOT NULL REFERENCES sportapp.plan_days(id),
    exercise_id INTEGER NOT NULL REFERENCES sportapp.exercises(id),
    position INTEGER NOT NULL CHECK (position >= 0),
    sets INTEGER NOT NULL CHECK (sets > 0),
    reps_min INTEGER NOT NULL CHECK (reps_min > 0),
    reps_max INTEGER NOT NULL CHECK (reps_max >= reps_min),
    weight_kg NUMERIC CHECK (weight_kg >= 0 AND weight_kg < 'Infinity'::numeric),
    rest_sec INTEGER NOT NULL CHECK (rest_sec >= 0),
    UNIQUE (plan_day_id, position)
);

CREATE TABLE sportapp.plan_schedules (
    plan_id UUID PRIMARY KEY REFERENCES sportapp.training_plans(id),
    time_zone TEXT NOT NULL,
    start_date DATE NOT NULL
);

CREATE TABLE sportapp.plan_schedule_entries (
    plan_id UUID NOT NULL REFERENCES sportapp.plan_schedules(plan_id),
    weekday INTEGER NOT NULL CHECK (weekday BETWEEN 1 AND 7),
    plan_day_id UUID NOT NULL,
    PRIMARY KEY (plan_id, weekday),
    FOREIGN KEY (plan_id, plan_day_id) REFERENCES sportapp.plan_days(plan_id, id)
);

CREATE TABLE sportapp.plan_revisions (
    id UUID PRIMARY KEY,
    plan_id UUID NOT NULL REFERENCES sportapp.training_plans(id),
    number BIGINT NOT NULL CHECK (number > 0),
    snapshot JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (plan_id, number)
);

COMMIT;
