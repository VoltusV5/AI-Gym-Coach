package users_postgres_repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	core_errors "sport_app/internal/core/errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// CreateUser создаёт зарегистрированного пользователя и начальные записи профиля, программы и user_data. Возвращает пользователя либо ошибку БД.
func (r *UsersRepository) CreateUser(
	ctx context.Context,
	email string,
	passwordHash string,
) (User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	createdAt := time.Now()

	query := `
	WITH inserted_user AS (
		INSERT INTO sportapp.users (is_anonymous, email, password_hash, subscription_status, created_at, updated_at)
		VALUES (false, $1, $2, 'free', $3, $3)
		RETURNING id, version, is_anonymous, email, subscription_status, created_at, updated_at
	),
	inserted_profile AS (
		INSERT INTO sportapp.profile (user_id, created_at)
		SELECT id, $3 FROM inserted_user
	),
	inserted_program AS (
		INSERT INTO sportapp.user_programs (user_id)
		SELECT id FROM inserted_user
	),
	inserted_data AS (
		INSERT INTO sportapp.user_data (user_id, created_at)
		SELECT id, $3 FROM inserted_user
	)
	SELECT id, version, is_anonymous, email, subscription_status, created_at, updated_at
	FROM inserted_user;
	`

	var u User
	err := r.pool.QueryRow(ctx, query, email, passwordHash, createdAt).Scan(
		&u.ID, &u.Version, &u.IsAnonymous, &u.Email,
		&u.SubscriptionStatus, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return User{}, fmt.Errorf("email '%s' already registered: %w", email, core_errors.ErrConflict)
		}
		return User{}, fmt.Errorf("create user: %w", err)
	}

	return u, nil
}
