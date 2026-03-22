package simplesql

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateTableUsers(ctx context.Context, conn *pgxpool.Pool) error {
	sqlQuery := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        is_anonymous BOOLEAN NOT NULL,
        email VARCHAR(200),
        password_hash VARCHAR(200),
        oauth_provider VARCHAR(200),
        oauth_id INTEGER,
        created_at TIMESTAMPTZ NOT NULL,
        updated_at TIMESTAMPTZ,
        subscription_status VARCHAR(200)
    );
    `
	_, err := conn.Exec(ctx, sqlQuery)
	return err
}

func CreateTableProfile(ctx context.Context, conn *pgxpool.Pool) error {
	sqlQuery := `
    CREATE TABLE IF NOT EXISTS profile (
        id SERIAL PRIMARY KEY,
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
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    );
    `
	_, err := conn.Exec(ctx, sqlQuery)
	return err
}

func CreateTableData(ctx context.Context, conn *pgxpool.Pool) error {
	sqlQuery := `
    CREATE TABLE IF NOT EXISTS user_data (
        id SERIAL PRIMARY KEY,
        user_id INTEGER NOT NULL UNIQUE,
        working_weights JSONB,
        completed_workouts JSONB,
        created_at TIMESTAMPTZ NOT NULL,
        updated_at TIMESTAMPTZ,
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    );
    `

	_, err := conn.Exec(ctx, sqlQuery)
	return err
}

func CreateTableExercises(ctx context.Context, conn *pgxpool.Pool) error {
	sqlQuery := `
    CREATE TABLE IF NOT EXISTS exercises (
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
    `

	_, err := conn.Exec(ctx, sqlQuery)
	return err
}

func CreateTableProgram(ctx context.Context, conn *pgxpool.Pool) error {
	sqlQuery := `
    CREATE TABLE IF NOT EXISTS user_programs (
        id SERIAL PRIMARY KEY,
        user_id INTEGER NOT NULL UNIQUE,
        started_at TIMESTAMP,
        planned_end_at TIMESTAMPTZ,
        is_active BOOLEAN,
        plan_template JSONB,
        plan_exercises JSONB,
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    );
    `

	_, err := conn.Exec(ctx, sqlQuery)
	return err
}
