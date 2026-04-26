package users_service

import (
	"context"
	"fmt"

	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
)

func (s *UsersService) CompleteWorkout(
	ctx context.Context,
	userID string,
	req users_postgres_repository.WorkoutCompleteRequest,
) error {
	if err := s.usersRepository.CompleteWorkout(ctx, userID, req); err != nil {
		return fmt.Errorf("complete workout: %w", err)
	}

	return nil
}
