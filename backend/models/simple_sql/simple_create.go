package simplesql

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func CreateTableUsers(ctx context.Context, conn *pgx.Conn) error {
	sqlQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		is_anonimous BOOLEAN NOT NULL,
		email VARCHAR(200) NOT NULL,
		password_hash VARCHAR(200) NOT NULL,
		oauth_provider VARCHAR(200),
		oauth_id INTEGER NOT NULL,
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
		user_id INTEGER NOT NULL,
		date DATE,
		gender VARCHAR(100) NOT NULL,
		height_cm INTEGER NOT NULL,
		weight_kg INTEGER NOT NULL,
		activity_level VARCHAR(200) NOT NULL,
		injuries_notes VARCHAR(200) NOT NULL,
		goal VARCHAR(200) NOT NULL,
		fitness_level VARCHAR(200) NOT NULL,
		training_days_map JSONB NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP
	);
	`
	_, err := conn.Exec(ctx, sqlQuery)
	return err
}
