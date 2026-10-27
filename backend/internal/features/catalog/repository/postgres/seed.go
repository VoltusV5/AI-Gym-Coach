package catalog_postgres_repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"

	"github.com/google/uuid"
)

type seedProgram struct {
	ID          string
	Title       string
	Description string
	Weeks       int
	Days        int
	Equipment   []string
	Exercises   []string
}

var seedPrograms = []seedProgram{
	{
		ID:          "18000000-0000-4000-8000-000000000001",
		Title:       "Базовая программа с гантелями",
		Description: "Три тренировочных дня с гантелями и скамьёй. Начальная нагрузка подбирается индивидуально; программа не заменяет консультацию тренера.",
		Weeks:       6, Days: 3, Equipment: []string{"dumbbell", "bench"},
		Exercises: []string{"Болгарские выпады с гантелями", "Жим гантелей на горизонтальной скамье", "Тяга гантели к поясу", "Румынская тяга с гантелями", "Махи с гантелями в стороны"},
	},
	{
		ID:          "18000000-0000-4000-8000-000000000002",
		Title:       "Базовая программа в зале",
		Description: "Два тренировочных дня на силовых и блочных тренажёрах. Начальная нагрузка подбирается индивидуально; программа не заменяет консультацию тренера.",
		Weeks:       4, Days: 2, Equipment: []string{"machine", "cable"},
		Exercises: []string{"Жим ногами лёжа", "Жим в хамере на середину груди", "Горизонтальная тяга с узкой ручкой", "Сгибания ног в тренажёре сидя", "Жим сидя на плечи в тренажёре"},
	},
}

func (r *Repository) SeedCatalog(ctx context.Context, trainerID int, displayName string) error {
	if trainerID <= 0 || strings.TrimSpace(displayName) == "" {
		return fmt.Errorf("trainer id and display name required: %w", core_errors.ErrInvalidArgument)
	}
	pool, ok := r.pool.(core_postgres_pool.TransactionalPool)
	if !ok {
		return fmt.Errorf("catalog seed requires transactions")
	}
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin catalog seed: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(20261027)`); err != nil {
		return fmt.Errorf("lock catalog seed: %w", err)
	}
	var anonymous bool
	if err := tx.QueryRow(ctx, `SELECT is_anonymous FROM sportapp.users WHERE id = $1 FOR SHARE`, trainerID).Scan(&anonymous); err != nil {
		return fmt.Errorf("read seed author: %w", err)
	}
	if anonymous {
		return fmt.Errorf("seed author must be registered: %w", core_errors.ErrForbidden)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO sportapp.trainer_profiles (user_id, display_name) VALUES ($1, $2)
		ON CONFLICT DO NOTHING`, trainerID, strings.TrimSpace(displayName)); err != nil {
		return fmt.Errorf("seed author profile: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO sportapp.user_roles (user_id, role) VALUES ($1, 'user'), ($1, 'trainer')
		ON CONFLICT DO NOTHING`, trainerID); err != nil {
		return fmt.Errorf("seed author roles: %w", err)
	}
	for _, program := range seedPrograms {
		if err := seedCatalogProgram(ctx, tx, trainerID, program); err != nil {
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit catalog seed: %w", err)
	}
	return nil
}

func seedCatalogProgram(ctx context.Context, tx core_postgres_pool.Tx, trainerID int, program seedProgram) error {
	var author int
	err := tx.QueryRow(ctx, `SELECT trainer_id FROM sportapp.programs WHERE id = $1`, program.ID).Scan(&author)
	if err == nil {
		if author != trainerID {
			return fmt.Errorf("seed program belongs to another author: %w", core_errors.ErrConflict)
		}
		return nil
	}
	if !errors.Is(err, core_postgres_pool.ErrNoRows) {
		return fmt.Errorf("check seed program: %w", err)
	}
	versionID := uuid.NewString()
	if _, err := tx.Exec(ctx, `INSERT INTO sportapp.programs (id, trainer_id) VALUES ($1, $2)`, program.ID, trainerID); err != nil {
		return fmt.Errorf("seed program: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO sportapp.program_versions
		(id, program_id, number, title, description, goal, level, duration_weeks, sessions_per_week, price_minor, currency, allow_public_remix, status)
		VALUES ($1, $2, 1, $3, $4, 'Набрать мышцы', 'новичок', $5, $6, 0, 'RUB', TRUE, 'published')`,
		versionID, program.ID, program.Title, program.Description, program.Weeks, program.Days); err != nil {
		return fmt.Errorf("seed version: %w", err)
	}
	for _, id := range program.Equipment {
		if _, err := tx.Exec(ctx, `INSERT INTO sportapp.program_version_equipment (version_id, equipment_id) VALUES ($1, $2)`, versionID, id); err != nil {
			return fmt.Errorf("seed version equipment: %w", err)
		}
	}
	for day := 0; day < program.Days; day++ {
		dayID := uuid.NewString()
		if _, err := tx.Exec(ctx, `INSERT INTO sportapp.program_days (id, program_version_id, title, position) VALUES ($1, $2, $3, $4)`,
			dayID, versionID, fmt.Sprintf("День %d", day+1), day); err != nil {
			return fmt.Errorf("seed day: %w", err)
		}
		for position, name := range program.Exercises {
			var exerciseID int
			if err := tx.QueryRow(ctx, `SELECT e.id FROM sportapp.exercises e
				WHERE e.exercises_name = $1 AND e.is_active AND NOT EXISTS (
					SELECT 1 FROM sportapp.exercise_equipment ee WHERE ee.exercise_id = e.id
					AND NOT (ee.equipment_id = ANY($2::text[]))) ORDER BY e.id LIMIT 1`, name, program.Equipment).Scan(&exerciseID); err != nil {
				return fmt.Errorf("find seed exercise %s: %w", name, err)
			}
			if _, err := tx.Exec(ctx, `INSERT INTO sportapp.program_exercises
				(id, program_day_id, exercise_id, position, sets, reps_min, reps_max, weight_kg, rest_sec)
				VALUES ($1, $2, $3, $4, 3, 8, 12, NULL, 90)`, uuid.NewString(), dayID, exerciseID, position); err != nil {
				return fmt.Errorf("seed prescription: %w", err)
			}
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE sportapp.programs SET status = 'active', published_version_id = $2 WHERE id = $1`, program.ID, versionID); err != nil {
		return fmt.Errorf("publish seed program: %w", err)
	}
	return nil
}
