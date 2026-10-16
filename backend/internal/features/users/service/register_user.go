package users_service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	users_postgres_repository "sport_app/internal/features/users/repository/postgres"

	"golang.org/x/crypto/bcrypt"
)

// RegisterUser either converts an existing anonymous guest account into a registered user,
// or directly creates a new registered account when no guest token is provided.
// Hashes the password using bcrypt and returns a signed non-guest JWT token.
//
// Parameters:
//   - ctx: Operation context.
//   - guestUserID: Optional guest user ID string (empty if direct registration).
//   - email: User email address (normalized to lowercase).
//   - password: Plaintext password.
//
// Returns:
//   - string: Signed JWT token with is_anonymous=false.
//   - users_postgres_repository.User: User account entity.
//   - error: core_errors.ErrConflict if email is already taken, or bcrypt/db/token generation error.
//
// Side effects:
//   - Mutates user row (sets email, password hash, is_anonymous=false).
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
