package users_postgres_repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
)

type Note struct {
	ID        int        `json:"id"`
	Version   int64      `json:"version"`
	UserID    int        `json:"user_id"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func (r *UsersRepository) GetListNotes(
	ctx context.Context,
	userID string,
) ([]Note, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	rows, err := r.pool.Query(ctx,
		`SELECT id, version, user_id, title, body, created_at, updated_at, deleted_at
		 FROM sportapp.notes
		 WHERE user_id = $1 AND deleted_at IS NULL
		 ORDER BY created_at DESC;`,
		userID,
	)
	if err != nil {
		return []Note{}, fmt.Errorf("insert note: %w", err)
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Version, &n.UserID, &n.Title, &n.Body, &n.CreatedAt, &n.UpdatedAt, &n.DeletedAt); err != nil {
			return []Note{}, fmt.Errorf("scan get notes: %w", err)
		}
		notes = append(notes, n)
	}
	return notes, nil
}

func (r *UsersRepository) GetNoteByID(
	ctx context.Context,
	userID string,
	noteID int,
) (Note, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var n Note
	err := r.pool.QueryRow(ctx,
		`SELECT id, version, user_id, title, body, created_at, updated_at, deleted_at
		 FROM sportapp.notes
		 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;`,
		noteID, userID,
	).Scan(&n.ID, &n.Version, &n.UserID, &n.Title, &n.Body, &n.CreatedAt, &n.UpdatedAt, &n.DeletedAt)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return Note{}, fmt.Errorf(
				"note id='%d' for user_id='%s' not found: %w",
				noteID,
				userID,
				core_errors.ErrNotFound,
			)
		}
		return Note{}, fmt.Errorf("scan get note by id: %w", err)
	}

	return n, nil
}

func (r *UsersRepository) CreateNotesUser(
	ctx context.Context,
	userID string,
	title string,
	body string,
) (Note, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var n Note
	if err := r.pool.QueryRow(ctx,
		`INSERT INTO sportapp.notes (user_id, title, body, created_at, updated_at)
		 VALUES ($1, $2, $3, NOW(), NOW())
		 RETURNING id, version, user_id, title, body, created_at, updated_at, deleted_at;`,
		userID, title, body,
	).Scan(&n.ID, &n.Version, &n.UserID, &n.Title, &n.Body, &n.CreatedAt, &n.UpdatedAt, &n.DeletedAt); err != nil {
		return Note{}, fmt.Errorf("scan create notes: %w", err)
	}

	return n, nil
}

func (r *UsersRepository) UpdateNotesUser(
	ctx context.Context,
	userID string,
	noteID int,
	expectedVersion int64,
	title string,
	body string,
) (Note, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var n Note
	if err := r.pool.QueryRow(ctx,
		`UPDATE sportapp.notes
		 SET title = $1, body = $2, updated_at = NOW(), version = version + 1
		 WHERE id = $3 AND user_id = $4 AND version = $5 AND deleted_at IS NULL
		 RETURNING id, version, user_id, title, body, created_at, updated_at, deleted_at;`,
		title, body, noteID, userID, expectedVersion,
	).Scan(&n.ID, &n.Version, &n.UserID, &n.Title, &n.Body, &n.CreatedAt, &n.UpdatedAt, &n.DeletedAt); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return Note{}, fmt.Errorf(
				"note id='%d' for user_id='%s' concurrently accessed or not found: %w",
				noteID,
				userID,
				core_errors.ErrConflict,
			)
		}
		return Note{}, fmt.Errorf("scan update notes: %w", err)
	}

	return n, nil
}

func (r *UsersRepository) DeleteNotesUser(
	ctx context.Context,
	userID string,
	noteID int,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	_, err := r.pool.Exec(ctx,
		`UPDATE sportapp.notes SET deleted_at = NOW() WHERE id = $1 AND user_id = $2;`,
		noteID, userID,
	)
	if err != nil {
		return fmt.Errorf("scan delete notes: %w", err)
	}

	return nil
}