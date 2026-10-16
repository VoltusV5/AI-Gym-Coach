package users_service

import (
	"context"
	"errors"
	"fmt"

	core_errors "sport_app/internal/core/errors"

	"golang.org/x/crypto/bcrypt"
)

func (s *UsersService) ChangeUserPassword(
	ctx context.Context,
	userID string,
	current_password string,
	new_password string,
) error {
	user, currentHash, err := s.usersRepository.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, core_errors.ErrNotFound) {
			return fmt.Errorf("invalid credentials: %w", core_errors.ErrUnauthorized)
		}
		return fmt.Errorf("get user by id: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(current_password)); err != nil {
		return fmt.Errorf("invalid credentials: %w", core_errors.ErrUnauthorized)
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(new_password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	if err := s.usersRepository.UpdatePassword(ctx, userID, user.Version, string(newHash)); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}
