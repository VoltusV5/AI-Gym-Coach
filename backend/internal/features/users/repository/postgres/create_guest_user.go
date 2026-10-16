package users_postgres_repository

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

func (r *UsersRepository) CreateGuestUser(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	createdAt := time.Now()

	query := `
	WITH inserted_user AS (
		INSERT INTO sportapp.users (is_anonymous, subscription_status, created_at)
		VALUES (true, 'free', $1)
		RETURNING id
	),
	inserted_profile AS (
		INSERT INTO sportapp.profile (user_id, created_at)
		SELECT id, $1 FROM inserted_user
	),
	inserted_program AS (
		INSERT INTO sportapp.user_programs (user_id)
		SELECT id FROM inserted_user
	)
	INSERT INTO sportapp.user_data (user_id, created_at)
	SELECT id, $1 FROM inserted_user
	RETURNING user_id;
	`

	var userID int
	if err := r.pool.QueryRow(ctx, query, createdAt).Scan(&userID); err != nil {
		return "", fmt.Errorf("insert guest user: %w", err)
	}

	return strconv.Itoa(userID), nil
}
