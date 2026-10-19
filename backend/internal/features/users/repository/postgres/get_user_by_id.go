package users_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
)

func (r *UsersRepository) GetUserByID(
	ctx context.Context,
	userID string,
) (User, string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, version, is_anonymous, email, password_hash,
	       subscription_status, created_at, updated_at
	FROM sportapp.users
	WHERE id = $1 AND is_anonymous = FALSE;
	`

	var (
		u            User
		passwordHash *string
	)
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&u.ID, &u.Version, &u.IsAnonymous, &u.Email, &passwordHash,
		&u.SubscriptionStatus, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return User{}, "", fmt.Errorf(
				"user with id='%s': %w", userID, core_errors.ErrNotFound,
			)
		}
		return User{}, "", fmt.Errorf("scan user by id: %w", err)
	}

	if passwordHash == nil {
		return User{}, "", fmt.Errorf(
			"user with id='%s' has no password set: %w",
			userID, core_errors.ErrUnauthorized,
		)
	}

	return u, *passwordHash, nil
}
