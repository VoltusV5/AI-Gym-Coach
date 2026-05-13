package users_service

import (
	"context"
	"fmt"
)

func (s *UsersService) EnsureAchievements(ctx context.Context) error {
	if err := s.usersRepository.EnsureAchievements(ctx); err != nil {
		return fmt.Errorf("ensure achievements: %w", err)
	}

	return nil
}