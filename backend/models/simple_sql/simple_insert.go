package simplesql

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

func InsertRowsUsers(
	ctx context.Context,
	conn *pgx.Conn,
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
	)
	INSERT INTO profile (user_id, created_at)
	SELECT id, $4 FROM inserted_row
	RETURNING user_id;
	`
	err := conn.QueryRow(ctx, sqlQuery, is_anonymous, subscription_status, created_at, created_at).Scan(&userID)
	return userID, err
}
