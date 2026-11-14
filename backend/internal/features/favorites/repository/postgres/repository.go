package favorites_postgres_repository

import (
	"context"
	"encoding/json"
	"errors"
	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
	"sport_app/internal/core/transport/http/v2request"
	catalog_postgres_repository "sport_app/internal/features/catalog/repository/postgres"
)

type Repository struct {
	pool core_postgres_pool.TransactionalPool
}

func NewRepository(pool core_postgres_pool.TransactionalPool) *Repository { return &Repository{pool} }
func member(ctx context.Context, reader interface {
	QueryRow(context.Context, string, ...any) core_postgres_pool.Row
}, user int) error {
	var valid bool
	err := reader.QueryRow(ctx, `SELECT NOT is_anonymous AND NOT blocked FROM sportapp.users WHERE id=$1`, user).Scan(&valid)
	if err != nil {
		return err
	}
	if !valid {
		return core_errors.ErrForbidden
	}
	return nil
}
func (r *Repository) Set(ctx context.Context, user int, program string, add bool) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	if err := member(ctx, r.pool, user); err != nil {
		return err
	}
	if !add {
		_, err := r.pool.Exec(ctx, `DELETE FROM sportapp.program_favorites WHERE user_id=$1 AND program_id=$2`, user, program)
		return err
	}
	var id string
	err := r.pool.QueryRow(ctx, `WITH available AS(SELECT p.id FROM sportapp.programs p JOIN sportapp.program_versions v ON v.id=p.published_version_id WHERE p.id=$2 AND p.status='active' AND v.status='published'),saved AS(INSERT INTO sportapp.program_favorites(user_id,program_id) SELECT $1,id FROM available ON CONFLICT DO NOTHING RETURNING program_id) SELECT id::text FROM available`, user, program).Scan(&id)
	if errors.Is(err, core_postgres_pool.ErrNoRows) {
		return core_errors.ErrNotFound
	}
	return err
}
func (r *Repository) List(ctx context.Context, user int, f v2request.PageFilter) (catalog_postgres_repository.Page[catalog_postgres_repository.ProgramCard], error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	page := catalog_postgres_repository.Page[catalog_postgres_repository.ProgramCard]{Items: []catalog_postgres_repository.ProgramCard{}}
	if err := member(ctx, r.pool, user); err != nil {
		return page, err
	}
	rows, err := r.pool.Query(ctx, `SELECT `+catalog_postgres_repository.CardJSON+catalog_postgres_repository.PublishedFrom+` AND EXISTS(SELECT 1 FROM sportapp.program_favorites f WHERE f.program_id=p.id AND f.user_id=$1) AND ($2='' OR p.id>NULLIF($2,'')::uuid) ORDER BY p.id LIMIT $3`, user, f.After, f.Limit+1)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		var raw []byte
		var p catalog_postgres_repository.ProgramCard
		if err := rows.Scan(&raw); err != nil {
			return page, err
		}
		if err := json.Unmarshal(raw, &p); err != nil {
			return page, err
		}
		page.Items = append(page.Items, p)
	}
	return page, rows.Err()
}
