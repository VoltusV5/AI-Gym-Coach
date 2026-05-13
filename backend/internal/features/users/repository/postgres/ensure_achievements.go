package users_postgres_repository

import (
	"context"
	"fmt"
)

func (r *UsersRepository) EnsureAchievements(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	achievements := []struct {
		Title       string
		Description string
		Category    *string
	}{
		{"first_workout", "Первая завершённая тренировка", nil},
		{"five_workouts", "Завершение 5 тренировок", nil},
		{"ten_workouts", "Завершение более 10 тренировок", nil},
		{"twenty_five_workouts", "25 тренировок", nil},
		{"week_warrior", "3 и более тренировок в неделю", nil},
		{"consistent_month", "12 тренировок за месяц", nil},
		{"comeback", "Возвращение к тренировкам после паузы", nil},
		{"early_bird", "Утренняя тренировка", nil},
		{"night_owl", "Тренировка в позднее время", nil},
		{"volume_session_5k", "В одной тренировке работа с суммарным весом, превышающим 5000кг", nil},
		{"volume_session_10k", "В одной тренировке работа с суммарным весом, превышающим 10000кг", nil},
		{"double_digit_sets", "Более 20 подходов за одну тренировку", nil},
		{"profile_ready", "Полностью заполненный профиль", nil},
	}

	for _, ach := range achievements {
		_, err := r.pool.Exec(ctx, `
			INSERT INTO sportapp.achievements (title, description, category)
			VALUES ($1, $2, $3)
			ON CONFLICT (title) DO UPDATE
			SET description = EXCLUDED.description, category = EXCLUDED.category
		`, ach.Title, ach.Description, ach.Category)
		if err != nil {
			return fmt.Errorf("ensure achievement %s: %w", ach.Title, err)
		}
	}

	return nil
}