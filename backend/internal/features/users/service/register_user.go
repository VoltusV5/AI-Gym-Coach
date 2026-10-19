package users_service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	users_postgres_repository "sport_app/internal/features/users/repository/postgres"

	"golang.org/x/crypto/bcrypt"
)

// RegisterUser преобразует гостя guestUserID в пользователя либо создаёт учётную запись, если guestUserID пуст.
// Хеширует password через bcrypt, сохраняет email и хеш и возвращает JWT, пользователя и ошибку.
// При занятом email возвращает ErrConflict; также возвращает ошибки БД, хеширования и подписи токена.
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

	if guestUserID == "" {
		user, err := s.usersRepository.CreateUser(ctx, email, string(hash))
		if err != nil {
			return "", users_postgres_repository.User{}, fmt.Errorf("create user: %w", err)
		}

		token, err := s.jwt.CreateToken(strconv.Itoa(user.ID), false)
		if err != nil {
			return "", users_postgres_repository.User{}, fmt.Errorf("create JWT token: %w", err)
		}

		return token, user, nil
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
