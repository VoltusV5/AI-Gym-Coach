package users_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	core_errors "sport_app/internal/core/errors"

	"github.com/jackc/pgx/v5"
)

func (r *UsersRepository) GetUserProgramsVersion(
	ctx context.Context,
	userID string,
) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var version int64
	err := r.pool.QueryRow(ctx, `
		SELECT version FROM sportapp.user_programs WHERE user_id = $1
	`, userID).Scan(&version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf(
				"user_programs for user_id='%s': %w",
				userID,
				core_errors.ErrNotFound,
			)
		}
		return 0, fmt.Errorf("scan user_programs version: %w", err)
	}

	return version, nil
}

func (r *UsersRepository) GetUserDataVersion(
	ctx context.Context,
	userID string,
) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var version int64
	err := r.pool.QueryRow(ctx, `
		SELECT version FROM sportapp.user_data WHERE user_id = $1
	`, userID).Scan(&version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf(
				"user_data for user_id='%s': %w",
				userID,
				core_errors.ErrNotFound,
			)
		}
		return 0, fmt.Errorf("scan user_data version: %w", err)
	}

	return version, nil
}
