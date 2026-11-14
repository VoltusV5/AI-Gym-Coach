package administration_postgres_repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
	"sport_app/internal/core/transport/http/v2request"
	trainers_postgres_repository "sport_app/internal/features/trainers/repository/postgres"
)

type User struct {
	ID          int      `json:"id"`
	Email       *string  `json:"email"`
	Roles       []string `json:"roles"`
	Blocked     bool     `json:"blocked"`
	IsAnonymous bool     `json:"is_anonymous"`
	CreatedAt   string   `json:"created_at"`
}
type Change struct {
	Blocked *bool     `json:"blocked"`
	Roles   *[]string `json:"roles"`
	Reason  string    `json:"reason"`
}
type Page struct {
	Items      []User  `json:"items"`
	NextCursor *string `json:"next_cursor"`
}
type Repository struct {
	pool core_postgres_pool.TransactionalPool
}

func NewRepository(pool core_postgres_pool.TransactionalPool) *Repository {
	return &Repository{pool: pool}
}

const userJSON = `jsonb_build_object('id',u.id,'email',u.email,'blocked',u.blocked,'is_anonymous',u.is_anonymous,'created_at',u.created_at,'roles',COALESCE((SELECT jsonb_agg(role ORDER BY role) FROM sportapp.user_roles WHERE user_id=u.id),'[]'::jsonb))`

func read(ctx context.Context, reader trainers_postgres_repository.Reader, id int) (User, error) {
	var raw []byte
	var u User
	err := reader.QueryRow(ctx, `SELECT `+userJSON+` FROM sportapp.users u WHERE u.id=$1`, id).Scan(&raw)
	if errors.Is(err, core_postgres_pool.ErrNoRows) {
		return u, fmt.Errorf("account missing: %w", core_errors.ErrNotFound)
	}
	if err != nil {
		return u, err
	}
	err = json.Unmarshal(raw, &u)
	return u, err
}
func (r *Repository) Session(ctx context.Context, id int) (User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	return read(ctx, r.pool, id)
}
func (r *Repository) List(ctx context.Context, actor int, f v2request.PageFilter, query string) (Page, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	p := Page{Items: []User{}}
	if err := trainers_postgres_repository.RequireRole(ctx, r.pool, actor, "admin"); err != nil {
		return p, err
	}
	rows, err := r.pool.Query(ctx, `SELECT `+userJSON+` FROM sportapp.users u WHERE u.id>COALESCE(NULLIF($1,'')::integer,0) AND ($2='' OR COALESCE(u.email,'') ILIKE '%' || $2 || '%') ORDER BY u.id LIMIT $3`, f.After, query, f.Limit+1)
	if err != nil {
		return p, err
	}
	defer rows.Close()
	for rows.Next() {
		var raw []byte
		var u User
		if err := rows.Scan(&raw); err != nil {
			return p, err
		}
		if err := json.Unmarshal(raw, &u); err != nil {
			return p, err
		}
		p.Items = append(p.Items, u)
	}
	return p, rows.Err()
}
func (r *Repository) Update(ctx context.Context, actor, target int, input Change, requestID string) (User, error) {
	var out User
	err := r.write(ctx, actor, func(ctx context.Context, tx core_postgres_pool.Tx) error {
		var anonymous bool
		if err := tx.QueryRow(ctx, `SELECT is_anonymous FROM sportapp.users WHERE id=$1 FOR UPDATE`, target).Scan(&anonymous); err != nil {
			if errors.Is(err, core_postgres_pool.ErrNoRows) {
				return core_errors.ErrNotFound
			}
			return err
		}
		before, err := read(ctx, tx, target)
		if err != nil {
			return err
		}
		if anonymous && input.Roles != nil {
			return fmt.Errorf("guest role change: %w", core_errors.ErrInvalidArgument)
		}
		if target == actor {
			return fmt.Errorf("self access change: %w", core_errors.ErrConflict)
		}
		if _, err := tx.Exec(ctx, `UPDATE sportapp.users SET blocked=COALESCE($2,blocked),block_reason=$3,version=version+1,updated_at=NOW() WHERE id=$1`, target, input.Blocked, input.Reason); err != nil {
			return err
		}
		if input.Roles != nil {
			if _, err := tx.Exec(ctx, `DELETE FROM sportapp.user_roles WHERE user_id=$1`, target); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `INSERT INTO sportapp.user_roles(user_id,role) SELECT $1,unnest($2::text[])`, target, *input.Roles); err != nil {
				return err
			}
		}

		out, err = read(ctx, tx, target)
		if err != nil {
			return err
		}
		oldState, _ := json.Marshal(map[string]any{"blocked": before.Blocked, "roles": before.Roles})
		newState, _ := json.Marshal(map[string]any{"blocked": out.Blocked, "roles": out.Roles})
		_, err = tx.Exec(ctx, `INSERT INTO sportapp.account_audit(id,actor_id,target_id,reason,request_id,before_state,after_state) VALUES($1,$2,$3,$4,$5,$6,$7)`, uuid.NewString(), actor, target, input.Reason, requestID, oldState, newState)
		return err
	})
	return out, err
}

func (r *Repository) write(ctx context.Context, actor int, fn func(context.Context, core_postgres_pool.Tx) error) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		c, stop := context.WithTimeout(context.WithoutCancel(ctx), r.pool.OpTimeout())
		defer stop()
		_ = tx.Rollback(c)
	}()
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(782431)`); err != nil {
		return err
	}
	if err := trainers_postgres_repository.RequireRole(ctx, tx, actor, "admin"); err != nil {
		return err
	}
	if err := fn(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
