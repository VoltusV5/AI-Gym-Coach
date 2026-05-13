package users_service

import (
	"context"
	"fmt"
)

func (s *UsersService) EnsureExercisesSeeded(ctx context.Context) error {
	if err := s.usersRepository.EnsureExercisesSeeded(ctx); err != nil {
		return fmt.Errorf("ensure exercises seeded: %w", err)
	}

	return nil
}