package simplesql

import (
	"context"
	"sport_app/mlclient"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InsertRowsUsers(
	ctx context.Context,
	conn *pgxpool.Pool,
	is_anonymous bool,
	subscription_status string,
) (string, error) {
	created_at := time.Now()
	var userID string
	sqlQuery := `
	WITH inserted_row AS (
		INSERT INTO users (is_anonymous, subscription_status, created_at)
		VALUES ($1, $2, $3)
		RETURNING id
	),
	inserted_rows AS (
		INSERT INTO profile (user_id, created_at)
		SELECT id, $4 FROM inserted_row
	),
	inserted_rrows AS (
		INSERT INTO user_programs (user_id)
		SELECT id FROM inserted_row
	)
	INSERT INTO user_data (user_id, created_at)
	SELECT id, $5 FROM inserted_row
	RETURNING user_id;
	`
	err := conn.QueryRow(ctx, sqlQuery, is_anonymous, subscription_status, created_at, created_at, created_at).Scan(&userID)
	return userID, err
}

func InsertRowsPrograms(
	userID string,
	is_active bool,
	plan_template mlclient.Plan,
	plan_exercises EPlanNoWeight,
	ctx context.Context,
	conn *pgxpool.Pool,
) error {
	sqlQuery := `
	UPDATE user_programs
	SET started_at = $1, planned_end_at = $2, is_active = $3, plan_template = $4, plan_exercises = $5
	WHERE user_id = $6;
	`
	start := time.Now()
	end := start.AddDate(0, 6, 0)

	_, err := conn.Exec(ctx, sqlQuery, start, end, is_active, plan_template, plan_exercises, userID)
	return err
}

func InsertRowsData(ctx context.Context, conn *pgxpool.Pool, userID string, working_weights []byte) error {
	sqlQuery := `
	UPDATE user_data
	SET working_weights = $1, updated_at = $2
	WHERE user_id = $3;
	`
	update := time.Now()
	_, err := conn.Exec(ctx, sqlQuery, working_weights, update, userID)
	return err
}
