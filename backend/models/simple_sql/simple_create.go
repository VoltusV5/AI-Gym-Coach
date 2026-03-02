package simplesql

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func CreateTableUsers(ctx context.Context, conn *pgx.Conn) error {
	sqlQuery := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        is_anonymous BOOLEAN NOT NULL,
        email VARCHAR(200),
        password_hash VARCHAR(200),
        oauth_provider VARCHAR(200),
        oauth_id INTEGER,
        created_at TIMESTAMP NOT NULL,
        updated_at TIMESTAMP,
        subscription_status VARCHAR(200)
    );
    `
	_, err := conn.Exec(ctx, sqlQuery)
	return err
}

func CreateTableProfile(ctx context.Context, conn *pgx.Conn) error {
	sqlQuery := `
    CREATE TABLE IF NOT EXISTS profile (
        id SERIAL PRIMARY KEY,
        user_id INTEGER NOT NULL UNIQUE,
        birth_date DATE,
        gender VARCHAR(100),
        height_cm INTEGER,
        weight_kg INTEGER,
        activity_level VARCHAR(200),
        injuries_notes VARCHAR(200),
        goal VARCHAR(200),
        fitness_level VARCHAR(200),
        training_days_map TEXT[],
        created_at TIMESTAMP NOT NULL,
        updated_at TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    );
    `
	_, err := conn.Exec(ctx, sqlQuery)
	return err
}
