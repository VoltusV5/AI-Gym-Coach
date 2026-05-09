package users_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
)

func (r *UsersRepository) LoadCompletedWorkouts(
	ctx context.Context,
	userID string,
) ([]WorkoutCompleteRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var raw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(completed_workouts, '[]'::jsonb)
		FROM sportapp.user_data
		WHERE user_id = $1
	`, userID).Scan(&raw)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return []WorkoutCompleteRequest{}, nil
		}
		return nil, fmt.Errorf("load completed_workouts: %w", err)
	}

	list, err := parseCompletedWorkoutsJSON(raw)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return []WorkoutCompleteRequest{}, nil
	}
	return list, nil
}
