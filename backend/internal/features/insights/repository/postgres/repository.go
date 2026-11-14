package insights_postgres_repository

import (
	"context"
	"encoding/json"
	"errors"
	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
	catalog "sport_app/internal/features/catalog/repository/postgres"
	"sport_app/internal/features/mlclient"
	profile "sport_app/internal/features/profile/repository/postgres"
	domain "sport_app/internal/features/workouts/domain"
	workouts "sport_app/internal/features/workouts/repository/postgres"
	"time"
)

type Repository struct {
	pool core_postgres_pool.TransactionalPool
}

func NewRepository(pool core_postgres_pool.TransactionalPool) *Repository { return &Repository{pool} }

type Snapshot struct {
	Profile profile.Profile
	Data    domain.Data
	Cards   []catalog.ProgramCard
	Context mlclient.RankingContext
}

func (r *Repository) Snapshot(ctx context.Context, user int) (Snapshot, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	var out Snapshot
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
	out, err = snapshotInTx(ctx, tx, user)
	if err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}
func snapshotInTx(ctx context.Context, tx core_postgres_pool.Tx, user int) (Snapshot, error) {
	var out Snapshot
	var err error
	out.Profile, err = profile.ReadSnapshot(ctx, tx, user)
	if errors.Is(err, core_errors.ErrNotFound) {
		out.Profile = profile.Profile{TrainingDays: []int{}, EquipmentIDs: []string{}, ExcludedExerciseIDs: []int{}}
		err = nil
	}
	if err != nil {
		return out, err
	}
	out.Data, err = workouts.DataSnapshot(ctx, tx, user)
	if err != nil {
		return out, err
	}
	p := out.Profile
	out.Context = mlclient.RankingContext{Level: p.Level, Goal: p.Goal, SessionsPerWeek: len(p.TrainingDays), EquipmentIDs: p.EquipmentIDs, Progress: mlclient.History{ExerciseMetrics: []mlclient.HistoryMetric{}}, Behavior: mlclient.Behavior{Viewed: []string{}, Saved: []string{}, Purchased: []string{}, Ratings: []mlclient.Rating{}}}
	now := time.Now().UTC()
	start := now.AddDate(0, 0, -28)
	progress := domain.Aggregate(out.Data, "28d", start, now, now, time.UTC)
	out.Context.Progress.Completed28d = progress.WorkoutCount
	out.Context.Progress.Adherence28d = progress.Adherence
	latest := map[int]mlclient.HistoryMetric{}
	for _, w := range out.Data.Workouts {
		if w.FinishedAt.Before(start) {
			continue
		}
		for _, slot := range w.Slots {
			metric := mlclient.HistoryMetric{ExerciseID: slot.ExerciseID, MeasuredAt: w.FinishedAt}
			for _, set := range slot.Sets {
				if set.IsWarmup {
					continue
				}
				metric.VolumeKg += set.WeightKg * float64(set.Reps)
				if set.WeightKg > metric.WeightKg || (set.WeightKg == metric.WeightKg && set.Reps > metric.Reps) {
					metric.WeightKg = set.WeightKg
					metric.Reps = set.Reps
				}
			}
			latest[slot.ExerciseID] = metric
		}
	}
	for _, m := range latest {
		out.Context.Progress.ExerciseMetrics = append(out.Context.Progress.ExerciseMetrics, m)
	}
	var raw []byte
	err = tx.QueryRow(ctx, `SELECT jsonb_build_object(
 'viewed_program_ids',COALESCE((SELECT jsonb_agg(program_id ORDER BY program_id) FROM (SELECT DISTINCT program_id FROM sportapp.behavior_events WHERE user_id=$1 AND type IN ('view','recommendation_click') AND occurred_at>=NOW()-INTERVAL '90 days' ORDER BY program_id LIMIT 200) q),'[]'),
 'saved_program_ids',COALESCE((SELECT jsonb_agg(program_id ORDER BY program_id) FROM (SELECT program_id FROM sportapp.program_favorites WHERE user_id=$1 ORDER BY program_id LIMIT 200) q),'[]'),
 'purchased_program_ids',COALESCE((SELECT jsonb_agg(program_id ORDER BY program_id) FROM (SELECT DISTINCT program_id FROM sportapp.orders WHERE user_id=$1 AND status='paid' ORDER BY program_id LIMIT 200) q),'[]'),
 'ratings',COALESCE((SELECT jsonb_agg(jsonb_build_object('program_id',program_id,'overall',overall) ORDER BY program_id) FROM (SELECT program_id,overall FROM sportapp.program_reviews WHERE user_id=$1 AND deleted_at IS NULL ORDER BY program_id LIMIT 200) q),'[]'))`, user).Scan(&raw)
	if err != nil {
		return out, err
	}
	if err = json.Unmarshal(raw, &out.Context.Behavior); err != nil {
		return out, err
	}
	rows, err := tx.Query(ctx, `SELECT `+catalog.CardJSON+catalog.PublishedFrom+`
 AND EXISTS(SELECT 1 FROM sportapp.users u WHERE u.id=p.trainer_id AND NOT u.blocked)
 AND NOT EXISTS(SELECT 1 FROM sportapp.program_version_equipment e WHERE e.version_id=v.id AND NOT(e.equipment_id=ANY($1::text[])))
 AND NOT EXISTS(SELECT 1 FROM sportapp.program_exercises e JOIN sportapp.program_days d ON d.id=e.program_day_id JOIN sportapp.exercises x ON x.id=e.exercise_id WHERE d.program_version_id=v.id AND (NOT x.is_active OR x.id=ANY($2::integer[]) OR ($3 AND NOT COALESCE(x.safe_for_injuries,FALSE))))
 ORDER BY p.id LIMIT 500`, p.EquipmentIDs, p.ExcludedExerciseIDs, p.InjuriesNotes != nil && *p.InjuriesNotes)
	if err != nil {
		return out, err
	}
	out.Cards = []catalog.ProgramCard{}
	for rows.Next() {
		var raw []byte
		var card catalog.ProgramCard
		if err = rows.Scan(&raw); err != nil {
			break
		}
		if err = json.Unmarshal(raw, &card); err != nil {
			break
		}
		out.Cards = append(out.Cards, card)
	}
	rows.Close()
	if err != nil {
		return out, err
	}
	if err = rows.Err(); err != nil {
		return out, err
	}
	return out, nil
}
func (r *Repository) Visible(ctx context.Context, ids []string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	rows, err := r.pool.Query(ctx, `SELECT p.id::text,v.id::text`+catalog.PublishedFrom+` AND p.id=ANY($1::uuid[]) AND EXISTS(SELECT 1 FROM sportapp.users u WHERE u.id=p.trainer_id AND NOT u.blocked)`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var id, version string
		if err := rows.Scan(&id, &version); err != nil {
			return nil, err
		}
		out[id] = version
	}
	return out, rows.Err()
}
