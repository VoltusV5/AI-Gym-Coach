package plans_postgres_repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type Repository struct {
	pool core_postgres_pool.TransactionalPool
}

func NewRepository(pool core_postgres_pool.TransactionalPool) *Repository {
	return &Repository{pool: pool}
}

const planJSON = `jsonb_build_object(
	'id', p.id, 'version', p.version, 'title', p.title, 'source_type', p.source_type,
	'source_program_version_id', p.source_program_version_id, 'is_active', p.is_active,
	'days', COALESCE((SELECT jsonb_agg(jsonb_build_object(
		'id', d.id, 'title', d.title, 'position', d.position,
		'exercises', COALESCE((SELECT jsonb_agg(jsonb_build_object(
			'exercise_id', e.exercise_id, 'position', e.position, 'sets', e.sets,
			'reps_min', e.reps_min, 'reps_max', e.reps_max, 'weight_kg', e.weight_kg, 'rest_sec', e.rest_sec)
			ORDER BY e.position) FROM sportapp.plan_exercises e WHERE e.plan_day_id = d.id), '[]'::jsonb))
		ORDER BY d.position) FROM sportapp.plan_days d WHERE d.plan_id = p.id), '[]'::jsonb),
	'schedule', (SELECT jsonb_build_object('time_zone', s.time_zone, 'start_date', s.start_date,
		'entries', COALESCE((SELECT jsonb_agg(jsonb_build_object('weekday', en.weekday, 'plan_day_id', en.plan_day_id)
			ORDER BY en.weekday) FROM sportapp.plan_schedule_entries en WHERE en.plan_id = p.id), '[]'::jsonb))
		FROM sportapp.plan_schedules s WHERE s.plan_id = p.id))`

type planReader interface {
	QueryRow(context.Context, string, ...any) core_postgres_pool.Row
}

func readPlan(ctx context.Context, reader planReader, userID int, planID string) (Plan, error) {
	var raw []byte
	var plan Plan
	err := reader.QueryRow(ctx, `SELECT `+planJSON+`, p.created_at FROM sportapp.training_plans p
		WHERE p.id = $1 AND p.owner_id = $2 AND p.archived_at IS NULL`, planID, userID).Scan(&raw, &plan.CreatedAt)
	if errors.Is(err, core_postgres_pool.ErrNoRows) {
		return Plan{}, fmt.Errorf("plan not found: %w", core_errors.ErrNotFound)
	}
	if err != nil {
		return Plan{}, fmt.Errorf("read plan: %w", err)
	}
	if err := json.Unmarshal(raw, &plan); err != nil {
		return Plan{}, fmt.Errorf("decode plan: %w", err)
	}
	return plan, nil
}

func (r *Repository) GetPlan(ctx context.Context, userID int, planID string) (Plan, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	return readPlan(ctx, r.pool, userID, planID)
}

func (r *Repository) ListPlans(ctx context.Context, userID int, filter Filter) ([]Plan, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	var after any
	if filter.AfterID != "" {
		after = filter.AfterID
	}
	rows, err := r.pool.Query(ctx, `SELECT `+planJSON+`, p.created_at FROM sportapp.training_plans p
		WHERE p.owner_id = $1 AND p.archived_at IS NULL
		AND ($2::uuid IS NULL OR (p.created_at, p.id) < ($3::timestamptz, $2::uuid))
		ORDER BY p.created_at DESC, p.id DESC LIMIT $4`, userID, after, filter.AfterCreatedAt, filter.Limit+1)
	if err != nil {
		return nil, fmt.Errorf("list plans: %w", err)
	}
	defer rows.Close()
	items := make([]Plan, 0, filter.Limit+1)
	for rows.Next() {
		var raw []byte
		var item Plan
		if err := rows.Scan(&raw, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan plan: %w", err)
		}
		if err := json.Unmarshal(raw, &item); err != nil {
			return nil, fmt.Errorf("decode plan: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read plans: %w", err)
	}
	return items, nil
}

func (r *Repository) write(ctx context.Context, userID int, operation func(context.Context, core_postgres_pool.Tx) (Plan, error)) (Plan, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Plan{}, fmt.Errorf("begin plan change: %w", err)
	}
	defer func() {
		rollbackCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), r.pool.OpTimeout())
		defer cancel()
		_ = tx.Rollback(rollbackCtx)
	}()
	var id int
	err = tx.QueryRow(ctx, `SELECT id FROM sportapp.users WHERE id = $1 FOR UPDATE`, userID).Scan(&id)
	if errors.Is(err, core_postgres_pool.ErrNoRows) {
		return Plan{}, fmt.Errorf("plan owner missing: %w", core_errors.ErrUnauthorized)
	}
	if err != nil {
		return Plan{}, fmt.Errorf("lock plan owner: %w", err)
	}
	plan, err := operation(ctx, tx)
	if err != nil {
		var pe *pgconn.PgError
		if errors.As(err, &pe) && (pe.Code == "23503" || pe.Code == "23505" || pe.Code == "23514") {
			return Plan{}, fmt.Errorf("invalid plan references: %w", core_errors.ErrInvalidArgument)
		}
		return Plan{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Plan{}, fmt.Errorf("commit plan change: %w", err)
	}
	return plan, nil
}

func saveRevision(ctx context.Context, tx core_postgres_pool.Tx, plan Plan) error {
	raw, err := json.Marshal(plan)
	if err != nil {
		return fmt.Errorf("encode plan revision: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO sportapp.plan_revisions (id, plan_id, number, snapshot)
		VALUES ($1, $2, $3, $4)`, uuid.NewString(), plan.ID, plan.Version, raw); err != nil {
		return fmt.Errorf("save plan revision: %w", err)
	}
	return nil
}
