package users_postgres_repository

import (
	"context"
	"fmt"
	"time"

	core_errors "sport_app/internal/core/errors"
)

func (r *UsersRepository) UpdatePassword(
	ctx context.Context,
	userID string,
	expectedVersion int64,
	newPasswordHash string,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	UPDATE sportapp.users
	SET password_hash = $1,
	    version = version + 1,
	    updated_at = $2
	WHERE id = $3 AND version = $4 AND is_anonymous = FALSE
	`

	tag, err := r.pool.Exec(ctx, query, newPasswordHash, time.Now(), userID, expectedVersion)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf(
			"password for user_id='%s' concurrently accessed or user not found: %w",
			userID,
			core_errors.ErrConflict,
		)
	}

	return nil
}