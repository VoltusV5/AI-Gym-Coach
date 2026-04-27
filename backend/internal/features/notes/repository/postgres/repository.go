package notes_postgres_repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Note struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) ListNotes(ctx context.Context, userID int) ([]Note, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, title, body, created_at, updated_at, deleted_at 
		 FROM sportapp.notes 
		 WHERE user_id = $1 AND deleted_at IS NULL 
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.CreatedAt, &n.UpdatedAt, &n.DeletedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, nil
}

func (r *Repository) CreateNote(ctx context.Context, userID int, title, body string) (Note, error) {
	var n Note
	err := r.pool.QueryRow(ctx,
		`INSERT INTO sportapp.notes (user_id, title, body, created_at, updated_at) 
		 VALUES ($1, $2, $3, NOW(), NOW()) 
		 RETURNING id, user_id, title, body, created_at, updated_at, deleted_at`,
		userID, title, body,
	).Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.CreatedAt, &n.UpdatedAt, &n.DeletedAt)
	return n, err
}

func (r *Repository) UpdateNote(ctx context.Context, userID int, noteID int, title, body string) (Note, error) {
	var n Note
	err := r.pool.QueryRow(ctx,
		`UPDATE sportapp.notes 
		 SET title = $1, body = $2, updated_at = NOW() 
		 WHERE id = $3 AND user_id = $4 AND deleted_at IS NULL 
		 RETURNING id, user_id, title, body, created_at, updated_at, deleted_at`,
		title, body, noteID, userID,
	).Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.CreatedAt, &n.UpdatedAt, &n.DeletedAt)
	return n, err
}

func (r *Repository) DeleteNote(ctx context.Context, userID int, noteID int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE sportapp.notes SET deleted_at = NOW() WHERE id = $1 AND user_id = $2`,
		noteID, userID,
	)
	return err
}
