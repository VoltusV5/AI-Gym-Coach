package users_postgres_repository

import (
	"context"
	"fmt"
)

func (r *UsersRepository) EnsureAchievements(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var n int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.achievements`).Scan(&n); err != nil {
		return fmt.Errorf("count achievements: %w", err)
	}
	if n > 0 {
		return nil
	}

	return r.insertAchievements(ctx)
}

func (r *UsersRepository) insertAchievements(ctx context.Context) error {
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
		{"night_owl", "Тренировка в позднее время", nil},
		{"volume_session_5k", "В одной тренировке работа с суммарным весом, превышающим 5000кг", nil},
		{"volume_session_10k", "В одной тренировке работа с суммарным весом, превышающим 10000кг", nil},
		{"double_digit_sets", "Более 20 подходов за одну тренировку", nil},
		{"profile_ready", "Полностью заполненный профиль", nil},
	}

	query := `INSERT INTO sportapp.achievements (title, description, category) VALUES `
	args := make([]any, 0, len(achievements)*3)
	for i, ach := range achievements {
		if i > 0 {
			query += ","
		}
		query += fmt.Sprintf("($%d, $%d, $%d)",
			i*3+1, i*3+2, i*3+3)
		args = append(args, ach.Title, ach.Description, ach.Category)
	}

	if _, err := r.pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("seed achievements: %w", err)
	}

	return nil
}
