package insights_postgres_repository

import (
	"context"
	"errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
	plans "sport_app/internal/features/plans/repository/postgres"
)

type DashboardSnapshot struct {
	Snapshot
	ActivePlan *plans.Plan
	Followed   []map[string]any
}

func (r *Repository) Dashboard(ctx context.Context, user int) (DashboardSnapshot, error) {
	out := DashboardSnapshot{Followed: []map[string]any{}}
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
	snapshot, err := snapshotInTx(ctx, tx, user)
	if err != nil {
		return out, err
	}
	out.Snapshot = snapshot
	var id string
	err = tx.QueryRow(ctx, `SELECT id::text FROM sportapp.training_plans WHERE owner_id=$1 AND is_active AND archived_at IS NULL`, user).Scan(&id)
	if err != nil && !errors.Is(err, core_postgres_pool.ErrNoRows) {
		return out, err
	}
	if err == nil {
		plan, err := plans.ReadSnapshot(ctx, tx, user, id)
		if err != nil {
			return out, err
		}
		out.ActivePlan = &plan
	}
	rows, err := tx.Query(ctx, `SELECT p.id::text,v.title,t.user_id,t.display_name FROM sportapp.trainer_follows f JOIN sportapp.trainer_profiles t ON t.user_id=f.trainer_id JOIN sportapp.users u ON u.id=t.user_id AND NOT u.blocked JOIN sportapp.programs p ON p.trainer_id=t.user_id AND p.status='active' JOIN sportapp.program_versions v ON v.id=p.published_version_id AND v.status='published' WHERE f.user_id=$1 ORDER BY v.created_at DESC,p.id LIMIT 5`, user)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, title, name string
		var trainer int
		if err := rows.Scan(&id, &title, &trainer, &name); err != nil {
			return out, err
		}
		out.Followed = append(out.Followed, map[string]any{"program_id": id, "title": title, "trainer_id": trainer, "display_name": name})
	}
	if err := rows.Err(); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}
