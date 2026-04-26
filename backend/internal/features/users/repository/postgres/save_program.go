package users_postgres_repository

import (
	"context"
	"fmt"
	"time"

	core_errors "sport_app/internal/core/errors"
	"sport_app/internal/features/mlclient"
)

func (r *UsersRepository) SaveProgram(
	ctx context.Context,
	userID string,
	expectedVersion int64,
	isActive bool,
	planTemplate mlclient.Plan,
	planExercises EPlanNoWeight,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	UPDATE sportapp.user_programs
	SET started_at = $1,
	    planned_end_at = $2,
	    is_active = $3,
	    plan_template = $4,
	    plan_exercises = $5,
	    version = version + 1
	WHERE user_id = $6 AND version = $7
	`

	start := time.Now()
	end := start.AddDate(0, 6, 0)

	tag, err := r.pool.Exec(ctx, query, start, end, isActive, planTemplate, planExercises, userID, expectedVersion)
	if err != nil {
		return fmt.Errorf("update user_programs: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf(
			"user_programs for user_id='%s' concurrently accessed or not found: %w",
			userID,
			core_errors.ErrConflict,
		)
	}

	return nil
}
