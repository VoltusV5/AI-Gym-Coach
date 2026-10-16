package users_service

import (
	"context"
	"fmt"
)

// CreateGuestUser provisions a temporary anonymous user profile and mints a guest JWT token (is_anonymous=true).
//
// Parameters:
//   - ctx: Operation context.
//
// Returns:
//   - string: Guest JWT token.
//   - error: Non-nil if database user generation or token creation fails.
//
// Side effects:
//   - Inserts anonymous user into users table.
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
