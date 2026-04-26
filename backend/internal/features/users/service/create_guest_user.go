package users_service

import (
	"context"
	"fmt"
)

func (s *UsersService) CreateGuestUser(ctx context.Context) (string, error) {
	userID, err := s.usersRepository.CreateGuestUser(ctx)
	if err != nil {
		return "", fmt.Errorf("create guest user: %w", err)
	}

	token, err := s.jwt.CreateToken(userID, true)
	if err != nil {
		return "", fmt.Errorf("create JWT token: %w", err)
	}

	return token, nil
}
