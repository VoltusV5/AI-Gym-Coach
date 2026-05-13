package users_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
)

func (r *UsersRepository) GetProfile(
	ctx context.Context,
	userID string,
) (Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, version, user_id, age, gender, height_cm, weight_kg,
	       activity_level, injuries_notes, goal, fitness_level,
	       training_days_map, created_at, updated_at
	FROM sportapp.profile
	WHERE user_id = $1;
	`

	var p Profile
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&p.ID, &p.Version, &p.UserID, &p.Age, &p.Gender, &p.HeightCm, &p.WeightKg,
		&p.ActivityLevel, &p.InjuriesNotes, &p.Goal, &p.FitnessLevel,
		&p.TrainingDaysMap, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return Profile{}, fmt.Errorf(
				"profile for user_id='%s': %w",
				userID,
				core_errors.ErrNotFound,
			)
		}
		return Profile{}, fmt.Errorf("scan profile: %w", err)
	}

	return p, nil
}