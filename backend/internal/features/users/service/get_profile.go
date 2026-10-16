package users_service

import (
	"context"
	"fmt"

	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
)

func (s *UsersService) GetProfile(
	ctx context.Context,
	userID string,
) (users_postgres_repository.Profile, error) {
	profile, err := s.usersRepository.GetProfile(ctx, userID)
	if err != nil {
		return users_postgres_repository.Profile{}, fmt.Errorf("get profile from repository: %w", err)
	}

	return profile, nil
}
