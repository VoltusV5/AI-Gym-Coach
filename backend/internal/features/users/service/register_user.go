package users_service

import (
	"context"
	"fmt"
	"strings"

	users_postgres_repository "sport_app/internal/features/users/repository/postgres"

	"golang.org/x/crypto/bcrypt"
)

func (s *UsersService) RegisterUser(
	ctx context.Context,
	guestUserID string,
	email string,
	password string,
) (string, users_postgres_repository.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", users_postgres_repository.User{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.usersRepository.PromoteGuestToUser(ctx, guestUserID, email, string(hash))
	if err != nil {
		return "", users_postgres_repository.User{}, fmt.Errorf("promote guest: %w", err)
	}

	token, err := s.jwt.CreateToken(guestUserID, false)
	if err != nil {
		return "", users_postgres_repository.User{}, fmt.Errorf("create JWT token: %w", err)
	}

	return token, user, nil
}
