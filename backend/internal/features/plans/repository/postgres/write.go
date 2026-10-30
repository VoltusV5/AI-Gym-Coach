package plans_postgres_repository

import (
	"context"
	"fmt"

	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"

	"github.com/google/uuid"
)

func replaceContents(ctx context.Context, tx core_postgres_pool.Tx, planID string, input Input) error {
	ids := []int{}
	dayIDs := []string{}
	for _, day := range input.Days {
		dayIDs = append(dayIDs, day.ID)
		for _, ex := range day.Exercises {
			ids = append(ids, ex.ExerciseID)
		}
	}
	var valid bool
	if err := tx.QueryRow(ctx, `SELECT NOT EXISTS (SELECT 1 FROM sportapp.plan_days
		WHERE id = ANY($1::uuid[]) AND plan_id <> $2)`, dayIDs, planID).Scan(&valid); err != nil {
		return fmt.Errorf("check plan days: %w", err)
	}
	if !valid {
		return fmt.Errorf("day belongs to another plan: %w", core_errors.ErrInvalidArgument)
	}
	rows, err := tx.Query(ctx, `SELECT id FROM sportapp.exercises WHERE id = ANY($1::integer[]) AND is_active FOR SHARE`, ids)
	if err != nil {
		return fmt.Errorf("check plan exercises: %w", err)
	}
	active := map[int]bool{}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return fmt.Errorf("scan exercise reference: %w", err)
		}
		active[id] = true
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return fmt.Errorf("read exercise references: %w", err)
	}
	for _, id := range ids {
		if !active[id] {
			return fmt.Errorf("unknown or inactive exercise: %w", core_errors.ErrInvalidArgument)
		}
	}
	for _, query := range []string{
		`DELETE FROM sportapp.plan_schedule_entries WHERE plan_id = $1`,
		`DELETE FROM sportapp.plan_exercises WHERE plan_day_id IN (SELECT id FROM sportapp.plan_days WHERE plan_id = $1)`,
		`DELETE FROM sportapp.plan_days WHERE plan_id = $1`,
	} {
		if _, err := tx.Exec(ctx, query, planID); err != nil {
			return fmt.Errorf("replace plan contents: %w", err)
		}
	}
	for _, day := range input.Days {
		if _, err := tx.Exec(ctx, `INSERT INTO sportapp.plan_days (id, plan_id, title, position) VALUES ($1, $2, $3, $4)`,
			day.ID, planID, day.Title, day.Position); err != nil {
			return fmt.Errorf("save plan day: %w", err)
		}
		for _, ex := range day.Exercises {
			if _, err := tx.Exec(ctx, `INSERT INTO sportapp.plan_exercises
				(id, plan_day_id, exercise_id, position, sets, reps_min, reps_max, weight_kg, rest_sec)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, uuid.NewString(), day.ID, ex.ExerciseID,
				ex.Position, ex.Sets, ex.RepsMin, ex.RepsMax, ex.WeightKg, ex.RestSec); err != nil {
				return fmt.Errorf("save prescription: %w", err)
			}
		}
	}
	if _, err := tx.Exec(ctx, `INSERT INTO sportapp.plan_schedules (plan_id, time_zone, start_date) VALUES ($1,$2,$3::date)
		ON CONFLICT (plan_id) DO UPDATE SET time_zone = EXCLUDED.time_zone, start_date = EXCLUDED.start_date`,
		planID, input.Schedule.TimeZone, input.Schedule.StartDate); err != nil {
		return fmt.Errorf("save schedule: %w", err)
	}
	for _, entry := range input.Schedule.Entries {
		if _, err := tx.Exec(ctx, `INSERT INTO sportapp.plan_schedule_entries (plan_id, weekday, plan_day_id) VALUES ($1,$2,$3)`,
			planID, entry.Weekday, entry.PlanDayID); err != nil {
			return fmt.Errorf("save scheduled day: %w", err)
		}
	}
	return nil
}

func (r *Repository) CreatePlan(ctx context.Context, userID int, input Input) (Plan, error) {
	return r.write(ctx, userID, func(ctx context.Context, tx core_postgres_pool.Tx) (Plan, error) {
		return CreatePlanInTx(ctx, tx, userID, input, "manual")
	})
}

func (r *Repository) UpdatePlan(ctx context.Context, userID int, planID string, input Input) (Plan, error) {
	return r.write(ctx, userID, func(ctx context.Context, tx core_postgres_pool.Tx) (Plan, error) {
		current, err := readPlan(ctx, tx, userID, planID)
		if err != nil {
			return Plan{}, err
		}
		if current.Version != input.Version {
			return Plan{}, fmt.Errorf("plan version changed: %w", core_errors.ErrConflict)
		}
		if err := replaceContents(ctx, tx, planID, input); err != nil {
			return Plan{}, err
		}
		if _, err := tx.Exec(ctx, `UPDATE sportapp.training_plans SET title = $3, version = version + 1
			WHERE id = $1 AND owner_id = $2`, planID, userID, input.Title); err != nil {
			return Plan{}, fmt.Errorf("update plan: %w", err)
		}
		plan, err := readPlan(ctx, tx, userID, planID)
		if err != nil {
			return Plan{}, err
		}
		return plan, saveRevision(ctx, tx, plan)
	})
}

func (r *Repository) ActivatePlan(ctx context.Context, userID int, planID string, version int64) (Plan, error) {
	return r.write(ctx, userID, func(ctx context.Context, tx core_postgres_pool.Tx) (Plan, error) {
		plan, err := readPlan(ctx, tx, userID, planID)
		if err != nil {
			return Plan{}, err
		}
		if plan.Version != version {
			return Plan{}, fmt.Errorf("plan version changed: %w", core_errors.ErrConflict)
		}
		var previousID *string
		if err := tx.QueryRow(ctx, `SELECT (SELECT id FROM sportapp.training_plans
			WHERE owner_id = $1 AND is_active AND archived_at IS NULL AND id <> $2)`, userID, planID).Scan(&previousID); err != nil {
			return Plan{}, fmt.Errorf("read previous active plan: %w", err)
		}
		if previousID != nil {
			if _, err := tx.Exec(ctx, `UPDATE sportapp.training_plans SET is_active = FALSE, version = version + 1
				WHERE id = $1 AND owner_id = $2`, *previousID, userID); err != nil {
				return Plan{}, fmt.Errorf("deactivate plan: %w", err)
			}
			previous, err := readPlan(ctx, tx, userID, *previousID)
			if err != nil {
				return Plan{}, err
			}
			if err := saveRevision(ctx, tx, previous); err != nil {
				return Plan{}, err
			}
		}
		if _, err := tx.Exec(ctx, `UPDATE sportapp.training_plans SET is_active = TRUE, version = version + 1
			WHERE id = $1 AND owner_id = $2`, planID, userID); err != nil {
			return Plan{}, fmt.Errorf("activate plan: %w", err)
		}
		plan, err = readPlan(ctx, tx, userID, planID)
		if err != nil {
			return Plan{}, err
		}
		return plan, saveRevision(ctx, tx, plan)
	})
}

func (r *Repository) ArchivePlan(ctx context.Context, userID int, planID string) error {
	_, err := r.write(ctx, userID, func(ctx context.Context, tx core_postgres_pool.Tx) (Plan, error) {
		plan, err := readPlan(ctx, tx, userID, planID)
		if err != nil {
			return Plan{}, err
		}
		plan.IsActive = false
		plan.Version++
		if _, err := tx.Exec(ctx, `UPDATE sportapp.training_plans SET archived_at = NOW(), is_active = FALSE, version = version + 1
			WHERE id = $1 AND owner_id = $2`, planID, userID); err != nil {
			return Plan{}, fmt.Errorf("archive plan: %w", err)
		}
		return plan, saveRevision(ctx, tx, plan)
	})
	return err
}

func CreatePlanInTx(ctx context.Context, tx core_postgres_pool.Tx, userID int, input Input, source string) (Plan, error) {
	id := uuid.NewString()
	if _, err := tx.Exec(ctx, `INSERT INTO sportapp.training_plans (id, owner_id, title, source_type) VALUES ($1,$2,$3,$4)`, id, userID, input.Title, source); err != nil {
		return Plan{}, fmt.Errorf("create plan: %w", err)
	}
	if err := replaceContents(ctx, tx, id, input); err != nil {
		return Plan{}, err
	}
	plan, err := readPlan(ctx, tx, userID, id)
	if err != nil {
		return Plan{}, err
	}
	return plan, saveRevision(ctx, tx, plan)
}
