package generation_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
	"sport_app/internal/features/mlclient"
	plans_postgres_repository "sport_app/internal/features/plans/repository/postgres"
	profile_postgres_repository "sport_app/internal/features/profile/repository/postgres"
	workouts_domain "sport_app/internal/features/workouts/domain"
	workouts_postgres_repository "sport_app/internal/features/workouts/repository/postgres"
)

type Result struct {
	PlanID       string   `json:"plan_id"`
	Version      int64    `json:"version"`
	ModelVersion string   `json:"model_version"`
	Warnings     []string `json:"warnings"`
}
type Repository struct {
	pool     core_postgres_pool.TransactionalPool
	profile  *profile_postgres_repository.Repository
	workouts *workouts_postgres_repository.Repository
}

func NewRepository(pool core_postgres_pool.TransactionalPool) *Repository {
	return &Repository{pool: pool, profile: profile_postgres_repository.NewRepository(pool), workouts: workouts_postgres_repository.NewRepository(pool)}
}
func (r *Repository) Cached(ctx context.Context, user int, key, hash string) (Result, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	var result Result
	hit, err := workouts_postgres_repository.Cached(ctx, r.pool, user, "generate", key, hash, &result)
	return result, hit, err
}
func (r *Repository) Profile(ctx context.Context, user int) (profile_postgres_repository.Profile, error) {
	return r.profile.GetProfile(ctx, user)
}
func (r *Repository) History(ctx context.Context, user int) (workouts_domain.Data, error) {
	return r.workouts.Data(ctx, user)
}
func (r *Repository) Save(ctx context.Context, user int, key, hash string, profileVersion int64, input plans_postgres_repository.Input, ml mlclient.GenerateResult) (Result, error) {
	var result Result
	err := r.workouts.Transaction(ctx, user, func(ctx context.Context, tx core_postgres_pool.Tx) error {
		hit, err := workouts_postgres_repository.Cached(ctx, tx, user, "generate", key, hash, &result)
		if err != nil || hit {
			return err
		}
		var version int64
		if err := tx.QueryRow(ctx, `SELECT version FROM sportapp.profile WHERE user_id=$1 FOR SHARE`, user).Scan(&version); err != nil {
			if errors.Is(err, core_postgres_pool.ErrNoRows) {
				return core_errors.ErrNotFound
			}
			return err
		}
		if version != profileVersion {
			return fmt.Errorf("profile changed during generation: %w", core_errors.ErrConflict)
		}
		ids := []int{}
		for _, day := range input.Days {
			for _, ex := range day.Exercises {
				ids = append(ids, ex.ExerciseID)
			}
		}
		var valid bool
		if err := tx.QueryRow(ctx, `SELECT NOT EXISTS(SELECT 1 FROM unnest($2::integer[]) x(id) WHERE
   NOT EXISTS(SELECT 1 FROM sportapp.exercises e WHERE e.id=x.id AND e.is_active)
   OR EXISTS(SELECT 1 FROM sportapp.profile_excluded_exercises ex WHERE ex.user_id=$1 AND ex.exercise_id=x.id)
   OR EXISTS(SELECT 1 FROM sportapp.exercise_equipment eq WHERE eq.exercise_id=x.id AND NOT EXISTS(SELECT 1 FROM sportapp.profile_equipment pe WHERE pe.user_id=$1 AND pe.equipment_id=eq.equipment_id)))`, user, ids).Scan(&valid); err != nil {
			return err
		}
		if !valid {
			return mlclient.NoFeasiblePlan()
		}
		plan, err := plans_postgres_repository.CreatePlanInTx(ctx, tx, user, input, "ai")
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO sportapp.plan_generations(id,user_id,plan_id,profile_version,request_id,model_version,feature_schema_version) VALUES($1,$2,$3,$4,$5,$6,'profile-v2')`, uuid.NewString(), user, plan.ID, profileVersion, ml.RequestID, ml.ModelVersion); err != nil {
			return err
		}
		result = Result{PlanID: plan.ID, Version: plan.Version, ModelVersion: ml.ModelVersion, Warnings: ml.Warnings}
		return workouts_postgres_repository.SaveResult(ctx, tx, user, "generate", key, hash, result)
	})
	return result, err
}
