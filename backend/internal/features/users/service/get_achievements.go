package users_service

import (
	"context"
	"fmt"

	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
)

func (s *UsersService) GetAchievements(
	ctx context.Context,
	userID string,
) ([]users_postgres_repository.UserAchievement, error) {
	list, err := s.usersRepository.GetAchievements(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get achievements: %w", err)
	}

	return list, nil
}