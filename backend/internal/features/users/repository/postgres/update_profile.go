package users_postgres_repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	core_errors "sport_app/internal/core/errors"
)

var allowedProfileFields = map[string]bool{
	"age":               true,
	"gender":            true,
	"height_cm":         true,
	"weight_kg":         true,
	"activity_level":    true,
	"injuries_notes":    true,
	"goal":              true,
	"fitness_level":     true,
	"training_days_map": true,
}

func (r *UsersRepository) UpdateProfile(
	ctx context.Context,
	userID string,
	expectedVersion int64,
	updates map[string]any,
) error {
	if len(updates) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	setParts := make([]string, 0, len(updates)+2)
	args := make([]any, 0, len(updates)+3)
	idx := 1

	for field, value := range updates {
		if !allowedProfileFields[field] {
			continue
		}
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, idx))
		args = append(args, value)
		idx++
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", idx))
	args = append(args, time.Now())
	idx++

	setParts = append(setParts, "version = version + 1")

	args = append(args, userID, expectedVersion)
	query := fmt.Sprintf(`
	UPDATE sportapp.profile
	SET %s
	WHERE user_id = $%d AND version = $%d
	`, strings.Join(setParts, ", "), idx, idx+1)

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update profile: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf(
			"profile for user_id='%s' concurrently accessed or not found: %w",
			userID,
			core_errors.ErrConflict,
		)
	}

	return nil
}
