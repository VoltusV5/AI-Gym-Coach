package simplesql

import (
	"context"
	"sport_app/mlclient"
	"strconv"
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
	var userID int
	sqlQuery := `
	WITH inserted_row AS (
		INSERT INTO sportapp.users (is_anonymous, subscription_status, created_at)
		VALUES ($1, $2, $3)
		RETURNING id
	),
	inserted_rows AS (
		INSERT INTO sportapp.profile (user_id, created_at)
		SELECT id, $4 FROM inserted_row
	),
	inserted_rrows AS (
		INSERT INTO sportapp.user_programs (user_id)
		SELECT id FROM inserted_row
	)
	INSERT INTO sportapp.user_data (user_id, created_at)
	SELECT id, $5 FROM inserted_row
	RETURNING user_id;
	`
	err := conn.QueryRow(ctx, sqlQuery, is_anonymous, subscription_status, created_at, created_at, created_at).Scan(&userID)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(userID), nil
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
	UPDATE sportapp.user_programs
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
	UPDATE sportapp.user_data
	SET working_weights = $1, updated_at = $2
	WHERE user_id = $3;
	`
	update := time.Now()
	_, err := conn.Exec(ctx, sqlQuery, working_weights, update, userID)
	return err
}
