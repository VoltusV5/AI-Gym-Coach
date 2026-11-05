package profile_postgres_repository

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
)

// Repository хранит профиль в общей с v1 таблице и связанные ограничения в таблицах v2.
type Repository struct {
	pool core_postgres_pool.TransactionalPool
}

// NewRepository создаёт репозиторий с поддержкой атомарного обновления анкеты и списков.
func NewRepository(pool core_postgres_pool.TransactionalPool) *Repository {
	return &Repository{pool: pool}
}

type profileReader interface {
	QueryRow(context.Context, string, ...any) core_postgres_pool.Row
}

var dayCodes = []string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}

var legacyLevels = map[string]string{
	"beginner": "Новичок", "intermediate": "Любитель", "advanced": "Продвинутый",
}

// readProfile читает анкету и её списки одним SQL-снимком, при lock блокирует строку до завершения транзакции.
// Возвращает ErrNotFound для отсутствующей анкеты либо ошибку БД.
func readProfile(ctx context.Context, reader profileReader, userID int, lock bool) (Profile, error) {
	query := `SELECT p.version, p.age, p.height_cm, p.weight_kg::double precision,
		p.gender, p.activity_level, p.injuries_notes, p.goal, p.fitness_level,
		p.training_days_map, p.session_duration_min,
		ARRAY(SELECT equipment_id FROM sportapp.profile_equipment WHERE user_id = p.user_id ORDER BY equipment_id),
		ARRAY(SELECT exercise_id FROM sportapp.profile_excluded_exercises WHERE user_id = p.user_id ORDER BY exercise_id)
		FROM sportapp.profile p WHERE p.user_id = $1`
	if lock {
		query += " FOR UPDATE OF p"
	}
	var profile Profile
	var days []string
	err := reader.QueryRow(ctx, query, userID).Scan(&profile.Version, &profile.Age, &profile.HeightCm,
		&profile.WeightKg, &profile.Gender, &profile.ActivityLevel, &profile.InjuriesNotes,
		&profile.Goal, &profile.Level, &days, &profile.SessionDurationMin,
		&profile.EquipmentIDs, &profile.ExcludedExerciseIDs)
	if errors.Is(err, core_postgres_pool.ErrNoRows) {
		return Profile{}, fmt.Errorf("profile not found: %w", core_errors.ErrNotFound)
	}
	if err != nil {
		return Profile{}, fmt.Errorf("read profile: %w", err)
	}
	if profile.Level != nil {
		level := strings.ToLower(strings.TrimSpace(*profile.Level))
		profile.Level = nil
		for canonical, legacy := range legacyLevels {
			if level == canonical || level == strings.ToLower(legacy) {
				profile.Level = &canonical
				break
			}
		}
	}
	profile.TrainingDays = []int{}
	russianDays := []string{"пн", "вт", "ср", "чт", "пт", "сб", "вс"}
	for _, day := range days {
		day = strings.ToLower(strings.TrimSpace(day))
		for i, code := range dayCodes {
			if day == code || day == russianDays[i] {
				profile.TrainingDays = append(profile.TrainingDays, i+1)
			}
		}
	}
	slices.Sort(profile.TrainingDays)
	profile.TrainingDays = slices.Compact(profile.TrainingDays)
	return profile, nil
}

// GetProfile возвращает текущую анкету пользователя userID, не читая чужие данные.
func (r *Repository) GetProfile(ctx context.Context, userID int) (Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	return readProfile(ctx, r.pool, userID, false)
}

// UpdateProfile атомарно обновляет валидированную анкету userID и её списки, увеличивая version.
// Проверяет существование оборудования и активность упражнений внутри транзакции.
// При устаревшей version возвращает ErrConflict без записи; при неизвестном ID — ErrInvalidArgument.
func (r *Repository) UpdateProfile(ctx context.Context, userID int, patch Patch) (Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Profile{}, fmt.Errorf("begin profile update: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	current, err := readProfile(ctx, tx, userID, true)
	if err != nil {
		return Profile{}, err
	}
	if current.Version != patch.Version {
		return Profile{}, fmt.Errorf("profile version changed: %w", core_errors.ErrConflict)
	}
	if err := validateReferences(ctx, tx, patch); err != nil {
		return Profile{}, err
	}
	sets, args := profileAssignments(patch)
	args = append(args, userID, patch.Version)
	query := fmt.Sprintf(`UPDATE sportapp.profile SET %s, version = version + 1, updated_at = NOW()
		WHERE user_id = $%d AND version = $%d`, strings.Join(sets, ", "), len(args)-1, len(args))
	if _, err := tx.Exec(ctx, query, args...); err != nil {
		return Profile{}, fmt.Errorf("update profile: %w", err)
	}
	if patch.EquipmentIDs != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM sportapp.profile_equipment WHERE user_id = $1`, userID); err != nil {
			return Profile{}, fmt.Errorf("clear profile equipment: %w", err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO sportapp.profile_equipment (user_id, equipment_id)
			SELECT $1, unnest($2::text[])`, userID, *patch.EquipmentIDs); err != nil {
			return Profile{}, fmt.Errorf("save profile equipment: %w", err)
		}
	}
	if patch.ExcludedExerciseIDs != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM sportapp.profile_excluded_exercises WHERE user_id = $1`, userID); err != nil {
			return Profile{}, fmt.Errorf("clear excluded exercises: %w", err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO sportapp.profile_excluded_exercises (user_id, exercise_id)
			SELECT $1, unnest($2::integer[])`, userID, *patch.ExcludedExerciseIDs); err != nil {
			return Profile{}, fmt.Errorf("save excluded exercises: %w", err)
		}
	}
	profile, err := readProfile(ctx, tx, userID, false)
	if err != nil {
		return Profile{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Profile{}, fmt.Errorf("commit profile update: %w", err)
	}
	return profile, nil
}

// validateReferences проверяет все переданные ID до изменения данных; FK защищают оборудование от удаления.
func validateReferences(ctx context.Context, tx core_postgres_pool.Tx, patch Patch) error {
	if patch.EquipmentIDs != nil {
		var valid bool
		if err := tx.QueryRow(ctx, `SELECT NOT EXISTS (SELECT 1 FROM unnest($1::text[]) AS ids(id)
			WHERE NOT EXISTS (SELECT 1 FROM sportapp.equipment WHERE id = ids.id))`, *patch.EquipmentIDs).Scan(&valid); err != nil {
			return fmt.Errorf("check profile equipment: %w", err)
		}
		if !valid {
			return fmt.Errorf("unknown equipment id: %w", core_errors.ErrInvalidArgument)
		}
	}
	if patch.ExcludedExerciseIDs != nil {
		rows, err := tx.Query(ctx, `SELECT id FROM sportapp.exercises
			WHERE id = ANY($1::integer[]) AND is_active FOR SHARE`, *patch.ExcludedExerciseIDs)
		if err != nil {
			return fmt.Errorf("check excluded exercises: %w", err)
		}
		defer rows.Close()
		count := 0
		for rows.Next() {
			count++
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("read excluded exercises: %w", err)
		}
		if count != len(*patch.ExcludedExerciseIDs) {
			return fmt.Errorf("unknown or inactive exercise: %w", core_errors.ErrInvalidArgument)
		}
	}
	return nil
}

// profileAssignments строит параметризованный SET только из фиксированного перечня полей.
// level и ISO-дни преобразуются в формат v1; пустой SET допускает обновление одних связанных списков.
func profileAssignments(patch Patch) ([]string, []any) {
	sets := []string{"user_id = user_id"}
	args := []any{}
	add := func(field string, value any) {
		args = append(args, value)
		sets = append(sets, fmt.Sprintf("%s = $%d", field, len(args)))
	}
	if patch.Age != nil {
		add("age", *patch.Age)
	}
	if patch.HeightCm != nil {
		add("height_cm", *patch.HeightCm)
	}
	if patch.WeightKg != nil {
		add("weight_kg", *patch.WeightKg)
	}
	if patch.Gender != nil {
		add("gender", *patch.Gender)
	}
	if patch.ActivityLevel != nil {
		add("activity_level", *patch.ActivityLevel)
	}
	if patch.InjuriesNotes != nil {
		add("injuries_notes", *patch.InjuriesNotes)
	}
	if patch.Goal != nil {
		add("goal", *patch.Goal)
	}
	if patch.Level != nil {
		add("fitness_level", legacyLevels[*patch.Level])
	}
	if patch.SessionDurationMin != nil {
		add("session_duration_min", *patch.SessionDurationMin)
	}
	if patch.TrainingDays != nil {
		days := []string{}
		for _, day := range *patch.TrainingDays {
			days = append(days, dayCodes[day-1])
		}
		add("training_days_map", days)
	}
	return sets, args
}
