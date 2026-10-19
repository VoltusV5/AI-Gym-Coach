package users_service

import (
	"context"
	"fmt"

	core_errors "sport_app/internal/core/errors"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
)

// UpdateProfile частично обновляет профиль userID полями updates с проверкой expectedVersion.
// Возвращает обновлённый профиль, ErrInvalidArgument при пустом updates, ErrConflict при конфликте версии либо ошибку БД.
// Сохраняет изменения и увеличивает версию профиля.
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
