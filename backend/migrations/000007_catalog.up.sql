BEGIN;

CREATE TABLE sportapp.user_roles (
    user_id INTEGER NOT NULL REFERENCES sportapp.users(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('user', 'trainer', 'admin')),
    PRIMARY KEY (user_id, role)
);

INSERT INTO sportapp.user_roles (user_id, role)
SELECT id, 'user' FROM sportapp.users WHERE NOT is_anonymous;

CREATE TABLE sportapp.equipment (
    id TEXT PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL
);

INSERT INTO sportapp.equipment (id, code, name) VALUES
    ('barbell', 'barbell', 'Штанга'),
    ('dumbbell', 'dumbbell', 'Гантели'),
    ('bench', 'bench', 'Скамья'),
    ('machine', 'machine', 'Силовой тренажёр'),
    ('cable', 'cable', 'Блочный тренажёр'),
    ('pull_up_bar', 'pull_up_bar', 'Турник'),
    ('parallel_bars', 'parallel_bars', 'Брусья'),
    ('kettlebell', 'kettlebell', 'Гиря');

ALTER TABLE sportapp.exercises ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE;

CREATE TABLE sportapp.exercise_equipment (
    exercise_id INTEGER NOT NULL REFERENCES sportapp.exercises(id),
    equipment_id TEXT NOT NULL REFERENCES sportapp.equipment(id),
    PRIMARY KEY (exercise_id, equipment_id)
);
CREATE INDEX exercise_equipment_equipment_idx ON sportapp.exercise_equipment(equipment_id, exercise_id);

CREATE TABLE sportapp.profile_equipment (
    user_id INTEGER NOT NULL REFERENCES sportapp.profile(user_id) ON DELETE CASCADE,
    equipment_id TEXT NOT NULL REFERENCES sportapp.equipment(id),
    PRIMARY KEY (user_id, equipment_id)
);

CREATE TABLE sportapp.trainer_profiles (
    user_id INTEGER PRIMARY KEY REFERENCES sportapp.users(id),
    display_name TEXT NOT NULL,
    photo_ref TEXT,
    bio TEXT NOT NULL DEFAULT '',
    experience_years INTEGER NOT NULL DEFAULT 0 CHECK (experience_years >= 0),
    qualifications TEXT NOT NULL DEFAULT '',
    verified_at TIMESTAMPTZ
);

CREATE TABLE sportapp.program_categories (
    id UUID PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    position INTEGER NOT NULL DEFAULT 0 CHECK (position >= 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE sportapp.programs (
    id UUID PRIMARY KEY,
    trainer_id INTEGER NOT NULL REFERENCES sportapp.trainer_profiles(user_id),
    published_version_id UUID,
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'hidden', 'archived')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE sportapp.program_versions (
    id UUID PRIMARY KEY,
    program_id UUID NOT NULL REFERENCES sportapp.programs(id),
    number INTEGER NOT NULL CHECK (number > 0),
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    goal TEXT NOT NULL,
    level TEXT NOT NULL,
    duration_weeks INTEGER NOT NULL CHECK (duration_weeks > 0),
    sessions_per_week INTEGER NOT NULL CHECK (sessions_per_week BETWEEN 1 AND 7),
    price_minor BIGINT NOT NULL CHECK (price_minor >= 0),
    currency TEXT NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    allow_public_remix BOOLEAN NOT NULL DEFAULT FALSE,
    status TEXT NOT NULL CHECK (status IN ('draft', 'in_review', 'published', 'rejected')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (program_id, number),
    UNIQUE (program_id, id)
);

ALTER TABLE sportapp.programs ADD CONSTRAINT programs_published_version_fk
    FOREIGN KEY (id, published_version_id) REFERENCES sportapp.program_versions(program_id, id);

CREATE TABLE sportapp.program_version_equipment (
    version_id UUID NOT NULL REFERENCES sportapp.program_versions(id),
    equipment_id TEXT NOT NULL REFERENCES sportapp.equipment(id),
    PRIMARY KEY (version_id, equipment_id)
);
CREATE INDEX program_version_equipment_equipment_idx ON sportapp.program_version_equipment(equipment_id, version_id);

CREATE TABLE sportapp.program_version_categories (
    version_id UUID NOT NULL REFERENCES sportapp.program_versions(id),
    category_id UUID NOT NULL REFERENCES sportapp.program_categories(id),
    PRIMARY KEY (version_id, category_id)
);

CREATE TABLE sportapp.program_days (
    id UUID PRIMARY KEY,
    program_version_id UUID NOT NULL REFERENCES sportapp.program_versions(id),
    title TEXT NOT NULL,
    position INTEGER NOT NULL CHECK (position >= 0),
    UNIQUE (program_version_id, position)
);

CREATE TABLE sportapp.program_exercises (
    id UUID PRIMARY KEY,
    program_day_id UUID NOT NULL REFERENCES sportapp.program_days(id),
    exercise_id INTEGER NOT NULL REFERENCES sportapp.exercises(id),
    position INTEGER NOT NULL CHECK (position >= 0),
    sets INTEGER NOT NULL CHECK (sets > 0),
    reps_min INTEGER NOT NULL CHECK (reps_min > 0),
    reps_max INTEGER NOT NULL CHECK (reps_max >= reps_min),
    weight_kg NUMERIC CHECK (weight_kg >= 0),
    rest_sec INTEGER NOT NULL CHECK (rest_sec >= 0),
    UNIQUE (program_day_id, position)
);

CREATE TABLE sportapp.entitlements (
    id UUID PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES sportapp.users(id),
    program_id UUID NOT NULL REFERENCES sportapp.programs(id),
    status TEXT NOT NULL CHECK (status IN ('active', 'revoked')),
    granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    order_id UUID,
    UNIQUE (user_id, program_id)
);

CREATE INDEX programs_catalog_idx ON sportapp.programs(status, id);
CREATE INDEX program_versions_filters_idx ON sportapp.program_versions(goal, level, duration_weeks);

COMMIT;
