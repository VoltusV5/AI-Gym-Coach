package users_service

import (
	"context"
	"fmt"

	core_errors "sport_app/internal/core/errors"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
)

// UpdateProfile applies partial profile modifications using optimistic locking to prevent concurrent overwrite issues.
//
// Parameters:
//   - ctx: Operation context.
//   - userID: Authenticated user ID.
//   - expectedVersion: Expected current version of user_data (used for optimistic locking).
//   - updates: Map of key-value fields to update (e.g. weight, height, goal).
//
// Returns:
//   - users_postgres_repository.Profile: Fresh profile state after update.
//   - error: core_errors.ErrInvalidArgument if updates map is empty, core_errors.ErrConflict if version does not match, or db error.
//
// Side effects:
//   - Updates user profile row in database and increments version number.
func (s *UsersService) UpdateProfile(
	ctx context.Context,
	userID string,
	expectedVersion int64,
	updates map[string]any,
) (users_postgres_repository.Profile, error) {
	if len(updates) == 0 {
		return users_postgres_repository.Profile{}, fmt.Errorf(
			"empty profile update: %w", core_errors.ErrInvalidArgument,
		)
	}

	if err := s.usersRepository.UpdateProfile(ctx, userID, expectedVersion, updates); err != nil {
		return users_postgres_repository.Profile{}, fmt.Errorf("update profile: %w", err)
	}

	profile, err := s.usersRepository.GetProfile(ctx, userID)
	if err != nil {
		return users_postgres_repository.Profile{}, fmt.Errorf("reload profile: %w", err)
	}

	return profile, nil
}
