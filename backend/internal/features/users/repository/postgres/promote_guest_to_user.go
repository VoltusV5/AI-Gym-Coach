package users_postgres_repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"

	"github.com/jackc/pgx/v5/pgconn"
)

func (r *UsersRepository) PromoteGuestToUser(
	ctx context.Context,
	userID string,
	email string,
	passwordHash string,
) (User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	UPDATE sportapp.users
	SET email = $1,
	    password_hash = $2,
	    is_anonymous = FALSE,
	    version = version + 1,
	    updated_at = $3
	WHERE id = $4 AND is_anonymous = TRUE
	RETURNING id, version, is_anonymous, email, subscription_status, created_at, updated_at;
	`

	var u User
	err := r.pool.QueryRow(ctx, query, email, passwordHash, time.Now(), userID).Scan(
		&u.ID, &u.Version, &u.IsAnonymous, &u.Email,
		&u.SubscriptionStatus, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return User{}, fmt.Errorf(
				"email '%s' already registered: %w", email, core_errors.ErrConflict,
			)
		}
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return User{}, fmt.Errorf(
				"user_id='%s' is not a guest or does not exist: %w",
				userID, core_errors.ErrConflict,
			)
		}
		return User{}, fmt.Errorf("promote guest user_id='%s': %w", userID, err)
	}

	return u, nil
}
