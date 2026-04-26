CREATE SCHEMA sportapp;

CREATE TABLE sportapp.users (
    id SERIAL PRIMARY KEY,
    version BIGINT NOT NULL DEFAULT 1,
    is_anonymous BOOLEAN NOT NULL,
    email VARCHAR(200) CHECK (
        email ~* '^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$'
        AND
        char_length(email) BETWEEN 5 AND 200
    ),
    password_hash VARCHAR(200),
    oauth_provider VARCHAR(200),
    oauth_id INTEGER,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ,
    subscription_status VARCHAR(200)
);

CREATE UNIQUE INDEX users_email_lower_uniq
    ON sportapp.users (LOWER(email))
    WHERE email IS NOT NULL;

CREATE TABLE sportapp.profile (
    id SERIAL PRIMARY KEY,
    version BIGINT NOT NULL DEFAULT 1,
    user_id INTEGER NOT NULL UNIQUE,
    age INTEGER,
    gender VARCHAR(100),
    height_cm INTEGER,
    weight_kg INTEGER,
    activity_level VARCHAR(200),
    injuries_notes BOOLEAN,
    goal VARCHAR(200),
    fitness_level VARCHAR(200),
    training_days_map TEXT[],
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ,
    FOREIGN KEY (user_id) REFERENCES sportapp.users(id) ON DELETE CASCADE    
);

CREATE TABLE sportapp.user_data (
    id SERIAL PRIMARY KEY,
    version BIGINT NOT NULL DEFAULT 1,
    user_id INTEGER NOT NULL UNIQUE,
    working_weights JSONB,
    completed_workouts JSONB,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ,
    FOREIGN KEY (user_id) REFERENCES sportapp.users(id) ON DELETE CASCADE
);

CREATE TABLE sportapp.exercises (
    id SERIAL PRIMARY KEY,
    exercises_name VARCHAR(100),
    muscular_group VARCHAR(100),
    muscular_subgroup VARCHAR(100),
    working_weights INTEGER,
    safe_for_injuries BOOLEAN,
    equipment VARCHAR(100),
    video_url VARCHAR(100),
    image_url VARCHAR(100)
);

CREATE TABLE sportapp.user_programs (
    id SERIAL PRIMARY KEY,
    version BIGINT NOT NULL DEFAULT 1,
    user_id INTEGER NOT NULL UNIQUE,
    started_at TIMESTAMP,
    planned_end_at TIMESTAMPTZ,
    is_active BOOLEAN,
    plan_template JSONB,
    plan_exercises JSONB,
    FOREIGN KEY (user_id) REFERENCES sportapp.users(id) ON DELETE CASCADE
);

CREATE TABLE sportapp.notes (
    id SERIAL PRIMARY KEY,
    version BIGINT NOT NULL DEFAULT 1,
    user_id INTEGER NOT NULL UNIQUE,
    title VARCHAR(120),
    body VARCHAR,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    FOREIGN KEY (user_id) REFERENCES sportapp.users(id) ON DELETE CASCADE
);