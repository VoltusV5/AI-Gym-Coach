package users_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	core_errors "sport_app/internal/core/errors"

	"github.com/jackc/pgx/v5"
)

func (r *UsersRepository) GetUserByEmail(
	ctx context.Context,
	email string,
) (User, string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, version, is_anonymous, email, password_hash,
	       subscription_status, created_at, updated_at
	FROM sportapp.users
	WHERE LOWER(email) = LOWER($1) AND is_anonymous = FALSE
	`

	var (
		u            User
		passwordHash *string
	)
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Version, &u.IsAnonymous, &u.Email, &passwordHash,
		&u.SubscriptionStatus, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, "", fmt.Errorf(
				"user with email='%s': %w", email, core_errors.ErrNotFound,
			)
		}
		return User{}, "", fmt.Errorf("scan user by email: %w", err)
	}

	if passwordHash == nil {
		return User{}, "", fmt.Errorf(
			"user with email='%s' has no password set: %w",
			email, core_errors.ErrUnauthorized,
		)
	}

	return u, *passwordHash, nil
}
