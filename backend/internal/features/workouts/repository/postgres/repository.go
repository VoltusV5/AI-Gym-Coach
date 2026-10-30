package workouts_postgres_repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
	plans_postgres_repository "sport_app/internal/features/plans/repository/postgres"
	workouts_domain "sport_app/internal/features/workouts/domain"
)

type Repository struct {
	pool core_postgres_pool.TransactionalPool
}

func NewRepository(pool core_postgres_pool.TransactionalPool) *Repository {
	return &Repository{pool: pool}
}

type Reader interface {
	Query(context.Context, string, ...any) (core_postgres_pool.Rows, error)
	QueryRow(context.Context, string, ...any) core_postgres_pool.Row
}

func (r *Repository) Transaction(ctx context.Context, user int, fn func(context.Context, core_postgres_pool.Tx) error) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		c, cancel := context.WithTimeout(context.WithoutCancel(ctx), r.pool.OpTimeout())
		defer cancel()
		_ = tx.Rollback(c)
	}()
	var id int
	if err := tx.QueryRow(ctx, `SELECT id FROM sportapp.users WHERE id=$1 FOR UPDATE`, user).Scan(&id); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return core_errors.ErrUnauthorized
		}
		return err
	}
	if err := fn(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func Cached(ctx context.Context, reader Reader, user int, operation, key, hash string, dest any) (bool, error) {
	var storedHash string
	var raw []byte
	err := reader.QueryRow(ctx, `SELECT request_hash,response FROM sportapp.operation_results WHERE user_id=$1 AND operation=$2 AND key=$3`, user, operation, key).Scan(&storedHash, &raw)
	if errors.Is(err, core_postgres_pool.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if storedHash != hash {
		return false, fmt.Errorf("idempotency payload differs: %w", core_errors.ErrConflict)
	}
	return true, json.Unmarshal(raw, dest)
}
func SaveResult(ctx context.Context, tx core_postgres_pool.Tx, user int, operation, key, hash string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO sportapp.operation_results(user_id,operation,key,request_hash,response) VALUES($1,$2,$3,$4,$5)`, user, operation, key, hash, raw)
	return err
}
func readWorkouts(ctx context.Context, reader Reader, user int) ([]workouts_domain.Workout, error) {
	rows, err := reader.Query(ctx, `SELECT id::text,payload FROM sportapp.workout_sessions WHERE user_id=$1 ORDER BY finished_at,id`, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []workouts_domain.Workout{}
	for rows.Next() {
		var w workouts_domain.Workout
		var raw []byte
		if err := rows.Scan(&w.ID, &raw); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &w.Input); err != nil {
			return nil, err
		}
		items = append(items, w)
	}
	return items, rows.Err()
}
func (r *Repository) Complete(ctx context.Context, user int, key, hash string, input workouts_domain.Input) (workouts_domain.Result, error) {
	var result workouts_domain.Result
	err := r.Transaction(ctx, user, func(ctx context.Context, tx core_postgres_pool.Tx) error {
		hit, err := Cached(ctx, tx, user, "workout", key, hash, &result)
		if err != nil || hit {
			return err
		}
		var oldHash string
		var raw []byte
		err = tx.QueryRow(ctx, `SELECT request_hash,result FROM sportapp.workout_sessions WHERE user_id=$1 AND client_session_id=$2`, user, input.ClientSessionID).Scan(&oldHash, &raw)
		if err == nil {
			if oldHash != hash {
				return core_errors.ErrConflict
			}
			if err := json.Unmarshal(raw, &result); err != nil {
				return err
			}
			return SaveResult(ctx, tx, user, "workout", key, hash, result)
		}
		if !errors.Is(err, core_postgres_pool.ErrNoRows) {
			return err
		}
		var revisionID string
		var snapshot []byte
		err = tx.QueryRow(ctx, `SELECT r.id::text,r.snapshot FROM sportapp.plan_revisions r JOIN sportapp.training_plans p ON p.id=r.plan_id WHERE p.owner_id=$1 AND p.id=$2 AND r.number=$3`, user, input.PlanID, input.PlanVersion).Scan(&revisionID, &snapshot)
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return core_errors.ErrNotFound
		}
		if err != nil {
			return err
		}
		var plan plans_postgres_repository.Plan
		if err := json.Unmarshal(snapshot, &plan); err != nil {
			return err
		}
		var day *plans_postgres_repository.Day
		for i := range plan.Days {
			if plan.Days[i].ID == input.PlanDayID {
				day = &plan.Days[i]
				break
			}
		}
		if day == nil {
			return core_errors.ErrInvalidArgument
		}
		positions := map[int]int{}
		for _, ex := range day.Exercises {
			positions[ex.Position] = ex.ExerciseID
		}
		for _, slot := range input.Slots {
			if positions[slot.Position] != slot.ExerciseID {
				return fmt.Errorf("exercise outside revision: %w", core_errors.ErrInvalidArgument)
			}
		}
		previous, err := readWorkouts(ctx, tx, user)
		if err != nil {
			return err
		}
		id := uuid.NewString()
		result = workouts_domain.Result{SessionID: id, Metrics: workouts_domain.Measure(input), Highlights: workouts_domain.Records(workouts_domain.Workout{ID: id, Input: input}, previous), NewAchievements: []any{}}
		payload, err := json.Marshal(input)
		if err != nil {
			return err
		}
		raw, err = json.Marshal(result)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO sportapp.workout_sessions(id,user_id,client_session_id,plan_id,revision_id,source_program_version_id,started_at,finished_at,request_hash,payload,result) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, id, user, input.ClientSessionID, input.PlanID, revisionID, plan.SourceProgramVersionID, input.StartedAt, input.FinishedAt, hash, payload, raw); err != nil {
			return err
		}
		return SaveResult(ctx, tx, user, "workout", key, hash, result)
	})
	return result, err
}
func (r *Repository) Get(ctx context.Context, user int, id string) (workouts_domain.Workout, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	var w workouts_domain.Workout
	var raw []byte
	err := r.pool.QueryRow(ctx, `SELECT id::text,payload FROM sportapp.workout_sessions WHERE user_id=$1 AND id=$2`, user, id).Scan(&w.ID, &raw)
	if errors.Is(err, core_postgres_pool.ErrNoRows) {
		return w, core_errors.ErrNotFound
	}
	if err != nil {
		return w, err
	}
	return w, json.Unmarshal(raw, &w.Input)
}
func (r *Repository) List(ctx context.Context, user int, f workouts_domain.Filter) ([]workouts_domain.Workout, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	rows, err := r.pool.Query(ctx, `SELECT id::text,payload FROM sportapp.workout_sessions WHERE user_id=$1
 AND ($2::timestamptz IS NULL OR finished_at>=$2) AND ($3::timestamptz IS NULL OR finished_at<$3)
 AND ($4='' OR plan_id=NULLIF($4,'')::uuid)
 AND ($5='' OR (finished_at,id)<($6::timestamptz,NULLIF($5,'')::uuid)) ORDER BY finished_at DESC,id DESC LIMIT $7`, user, optionalTime(f.From), optionalTime(f.To), f.PlanID, f.AfterID, f.AfterTime, f.Limit+1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []workouts_domain.Workout{}
	for rows.Next() {
		var w workouts_domain.Workout
		var raw []byte
		if err := rows.Scan(&w.ID, &raw); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &w.Input); err != nil {
			return nil, err
		}
		items = append(items, w)
	}
	return items, rows.Err()
}
func optionalTime(t interface{ IsZero() bool }) any {
	if t.IsZero() {
		return nil
	}
	return t
}
func (r *Repository) Weight(ctx context.Context, user int, w workouts_domain.Weight) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	_, err := r.pool.Exec(ctx, `INSERT INTO sportapp.nutrition_weight_logs(user_id,logged_on,weight_kg) VALUES($1,$2::date,$3)
 ON CONFLICT(user_id,logged_on) DO UPDATE SET weight_kg=EXCLUDED.weight_kg`, user, w.Day, w.WeightKg)
	return err
}
func (r *Repository) Data(ctx context.Context, user int) (workouts_domain.Data, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	var data workouts_domain.Data
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return data, err
	}
	defer func() {
		c, cancel := context.WithTimeout(context.WithoutCancel(ctx), r.pool.OpTimeout())
		defer cancel()
		_ = tx.Rollback(c)
	}()
	if _, err := tx.Exec(ctx, `SET TRANSACTION ISOLATION LEVEL REPEATABLE READ READ ONLY`); err != nil {
		return data, err
	}
	data.Workouts, err = readWorkouts(ctx, tx, user)
	if err != nil {
		return data, err
	}
	rows, err := tx.Query(ctx, `SELECT logged_on::text,weight_kg FROM sportapp.nutrition_weight_logs WHERE user_id=$1 ORDER BY logged_on`, user)
	if err != nil {
		return data, err
	}
	for rows.Next() {
		var w workouts_domain.Weight
		if err := rows.Scan(&w.Day, &w.WeightKg); err != nil {
			rows.Close()
			return data, err
		}
		data.Weights = append(data.Weights, w)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return data, err
	}
	rows, err = tx.Query(ctx, `SELECT r.snapshot,r.created_at FROM sportapp.plan_revisions r JOIN sportapp.training_plans p ON p.id=r.plan_id WHERE p.owner_id=$1 ORDER BY r.created_at,r.number`, user)
	if err != nil {
		return data, err
	}
	for rows.Next() {
		var rev workouts_domain.Revision
		var raw []byte
		if err := rows.Scan(&raw, &rev.At); err != nil {
			rows.Close()
			return data, err
		}
		if err := json.Unmarshal(raw, &rev.Plan); err != nil {
			rows.Close()
			return data, err
		}
		data.Revisions = append(data.Revisions, rev)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return data, err
	}
	return data, tx.Commit(ctx)
}
