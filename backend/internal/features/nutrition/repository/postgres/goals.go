package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
)

// GetGoal retrieves the nutrition goal for a user.
func (r *Repository) GetGoal(
	ctx context.Context,
	userID int,
) (NutritionGoal, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT protein_g, fat_g, carbs_g, calories, updated_at
		FROM sportapp.nutrition_goals
		WHERE user_id = $1`

	var g NutritionGoal
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&g.ProteinG, &g.FatG, &g.CarbsG, &g.Calories, &g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return NutritionGoal{}, fmt.Errorf("nutrition goal not found: %w", core_errors.ErrNotFound)
		}
		return NutritionGoal{}, fmt.Errorf("query nutrition goal: %w", err)
	}

	return g, nil
}

// UpsertGoal inserts or updates the nutrition goal for a user.
func (r *Repository) UpsertGoal(
	ctx context.Context,
	userID int,
	goal NutritionGoal,
) (NutritionGoal, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO sportapp.nutrition_goals (user_id, protein_g, fat_g, carbs_g, calories)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE
		SET protein_g = EXCLUDED.protein_g,
		    fat_g = EXCLUDED.fat_g,
		    carbs_g = EXCLUDED.carbs_g,
		    calories = EXCLUDED.calories,
		    updated_at = NOW()
		RETURNING updated_at`

	var updatedAt time.Time
	err := r.pool.QueryRow(
		ctx, query,
		userID, goal.ProteinG, goal.FatG, goal.CarbsG, goal.Calories,
	).Scan(&updatedAt)
	if err != nil {
		return NutritionGoal{}, fmt.Errorf("upsert nutrition goal: %w", err)
	}

	goal.UpdatedAt = updatedAt
	return goal, nil
}

// GetUserProfile retrieves profile parameters used in nutrition calculations.
func (r *Repository) GetUserProfile(
	ctx context.Context,
	userID int,
) (UserProfileData, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT age, gender, height_cm, weight_kg, activity_level, goal
		FROM sportapp.profile
		WHERE user_id = $1`

	var p UserProfileData
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&p.Age, &p.Gender, &p.HeightCm, &p.WeightKg, &p.ActivityLevel, &p.Goal,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return UserProfileData{}, nil
		}
		return UserProfileData{}, fmt.Errorf("query profile for goal: %w", err)
	}

	return p, nil
}

// GetLatestWeightKg retrieves the latest logged weight for a user.
func (r *Repository) GetLatestWeightKg(
	ctx context.Context,
	userID int,
) (*float64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT weight_kg
		FROM sportapp.nutrition_weight_logs
		WHERE user_id = $1
		ORDER BY logged_on DESC
		LIMIT 1`

	var w *float64
	err := r.pool.QueryRow(ctx, query, userID).Scan(&w)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query latest weight: %w", err)
	}

	return w, nil
}

// GetProfileWeightKg retrieves the baseline weight recorded in profile.
func (r *Repository) GetProfileWeightKg(
	ctx context.Context,
	userID int,
) (*int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `SELECT weight_kg FROM sportapp.profile WHERE user_id = $1`
	var w *int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&w)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query profile weight: %w", err)
	}

	return w, nil
}
