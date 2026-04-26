package users_service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	core_errors "sport_app/internal/core/errors"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"

	"golang.org/x/crypto/bcrypt"
)

func (s *UsersService) LoginUser(
	ctx context.Context,
	email string,
	password string,
) (string, users_postgres_repository.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, passwordHash, err := s.usersRepository.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, core_errors.ErrNotFound) {
			return "", users_postgres_repository.User{}, fmt.Errorf(
				"invalid credentials: %w", core_errors.ErrUnauthorized,
			)
		}
		return "", users_postgres_repository.User{}, fmt.Errorf("get user by email: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return "", users_postgres_repository.User{}, fmt.Errorf(
			"invalid credentials: %w", core_errors.ErrUnauthorized,
		)
	}

	token, err := s.jwt.CreateToken(strconv.Itoa(user.ID), false)
	if err != nil {
		return "", users_postgres_repository.User{}, fmt.Errorf("create JWT token: %w", err)
	}

	return token, user, nil
}
