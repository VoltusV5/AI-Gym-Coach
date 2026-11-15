package insights_postgres_repository

import (
	"context"
	"errors"
	"fmt"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
	plans "sport_app/internal/features/plans/repository/postgres"
	workouts "sport_app/internal/features/workouts/repository/postgres"
)

type DashboardSnapshot struct {
	Snapshot
	ActivePlan  *plans.Plan
	Followed    []map[string]any
	Unavailable map[string]bool
}

func (r *Repository) Dashboard(ctx context.Context, user int) (DashboardSnapshot, error) {
	out := DashboardSnapshot{Followed: []map[string]any{}, Unavailable: map[string]bool{}}
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer func() {
		c, cancel := context.WithTimeout(context.WithoutCancel(ctx), r.pool.OpTimeout())
		defer cancel()
		_ = tx.Rollback(c)
	}()
	if _, err = tx.Exec(ctx, `SET TRANSACTION ISOLATION LEVEL REPEATABLE READ, READ ONLY`); err != nil {
		return out, err
	}
	milliseconds := max(1, r.pool.OpTimeout().Milliseconds()/4)
	if _, err = tx.Exec(ctx, fmt.Sprintf(`SET LOCAL statement_timeout = %d`, milliseconds)); err != nil {
		return out, err
	}
	block := func(name string, read func() error) error {
		if _, err := tx.Exec(ctx, `SAVEPOINT dashboard_block`); err != nil {
			return err
		}
		if err := read(); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if _, err := tx.Exec(ctx, `ROLLBACK TO SAVEPOINT dashboard_block`); err != nil {
				return err
			}
			out.Unavailable[name] = true
		}
		_, err := tx.Exec(ctx, `RELEASE SAVEPOINT dashboard_block`)
		return err
	}
	if err := block("recent_progress", func() error {
		data, err := workouts.DataSnapshot(ctx, tx, user)
		if err == nil {
			out.Data = data
		}
		return err
	}); err != nil {
		return out, err
	}
	if err := block("today_workout", func() error {
		var id string
		err := tx.QueryRow(ctx, `SELECT id::text FROM sportapp.training_plans WHERE owner_id=$1 AND is_active AND archived_at IS NULL`, user).Scan(&id)
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		plan, err := plans.ReadSnapshot(ctx, tx, user, id)
		if err == nil {
			out.ActivePlan = &plan
		}
		return err
	}); err != nil {
		return out, err
	}
	if err := block("followed_trainers", func() error {
		rows, err := tx.Query(ctx, `SELECT p.id::text,v.title,t.user_id,t.display_name FROM sportapp.trainer_follows f JOIN sportapp.trainer_profiles t ON t.user_id=f.trainer_id JOIN sportapp.users u ON u.id=t.user_id AND NOT u.blocked JOIN sportapp.programs p ON p.trainer_id=t.user_id AND p.status='active' JOIN sportapp.program_versions v ON v.id=p.published_version_id AND v.status='published' WHERE f.user_id=$1 ORDER BY v.created_at DESC,p.id LIMIT 5`, user)
		if err != nil {
			return err
		}
		defer rows.Close()
		items := []map[string]any{}
		for rows.Next() {
			var id, title, name string
			var trainer int
			if err := rows.Scan(&id, &title, &trainer, &name); err != nil {
				return err
			}
			items = append(items, map[string]any{"program_id": id, "title": title, "trainer_id": trainer, "display_name": name})
		}
		if err := rows.Err(); err != nil {
			return err
		}
		out.Followed = items
		return nil
	}); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}
