package users_postgres_repository

import (
	"context"
	"encoding/json"
	"fmt"

	core_errors "sport_app/internal/core/errors"
)

func (r *UsersRepository) CompleteWorkout(
	ctx context.Context,
	userID string,
	req WorkoutCompleteRequest,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal workout data: %w", err)
	}

	query := `
	UPDATE sportapp.user_data
	SET completed_workouts = COALESCE(completed_workouts, '[]'::jsonb) || jsonb_build_array($1::jsonb),
	    updated_at = NOW(),
	    version = version + 1
	WHERE user_id = $2
	`

	tag, err := r.pool.Exec(ctx, query, data, userID)
	if err != nil {
		return fmt.Errorf("update completed_workouts: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf(
			"user_data for user_id='%s': %w",
			userID,
			core_errors.ErrNotFound,
		)
	}

	return nil
}
