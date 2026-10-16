package users_postgres_repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
)

func (r *UsersRepository) UnlockNewAchievements(
	ctx context.Context,
	userID string,
) ([]AchievementUnlocked, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var raw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(completed_workouts, '[]'::jsonb)
		FROM sportapp.user_data
		WHERE user_id = $1
	`, userID).Scan(&raw)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return nil, fmt.Errorf(
				"user_data for user_id='%s': %w",
				userID,
				core_errors.ErrNotFound,
			)
		}
		return nil, fmt.Errorf("load completed_workouts: %w", err)
	}

	workouts, err := parseCompletedWorkoutsJSON(raw)
	if err != nil {
		return nil, fmt.Errorf("parse completed_workouts: %w", err)
	}

	profile, err := r.GetProfile(ctx, userID)
	if err != nil && !errors.Is(err, core_errors.ErrNotFound) {
		return nil, fmt.Errorf("load profile for achievements: %w", err)
	}
	if errors.Is(err, core_errors.ErrNotFound) {
		profile = Profile{}
	}

	now := time.Now()
	earned := earnedAchievementTitles(workouts, profile, now)

	rows, err := r.pool.Query(ctx, `
		SELECT achievement_id FROM sportapp.user_achievements WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list user achievements: %w", err)
	}
	defer rows.Close()

	already := map[int]struct{}{}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan user achievement: %w", err)
		}
		already[id] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user achievements: %w", err)
	}

	titleRows, err := r.pool.Query(ctx, `SELECT id, title FROM sportapp.achievements`)
	if err != nil {
		return nil, fmt.Errorf("list achievements: %w", err)
	}
	defer titleRows.Close()

	titleByID := map[int]string{}
	idByTitle := map[string]int{}
	for titleRows.Next() {
		var id int
		var title string
		if err := titleRows.Scan(&id, &title); err != nil {
			return nil, fmt.Errorf("scan achievement: %w", err)
		}
		titleByID[id] = title
		idByTitle[title] = id
	}
	if err := titleRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate achievements: %w", err)
	}

	var toInsert []int
	for title := range earned {
		id, ok := idByTitle[title]
		if !ok {
			continue
		}
		if _, has := already[id]; has {
			continue
		}
		toInsert = append(toInsert, id)
	}
	sort.Ints(toInsert)

	inserted := make([]AchievementUnlocked, 0, len(toInsert))
	for _, achID := range toInsert {
		tag, err := r.pool.Exec(ctx, `
			INSERT INTO sportapp.user_achievements (user_id, achievement_id, unlocked_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (user_id, achievement_id) DO NOTHING
		`, userID, achID)
		if err != nil {
			return nil, fmt.Errorf("insert user_achievement: %w", err)
		}
		if tag.RowsAffected() == 0 {
			continue
		}
		inserted = append(inserted, AchievementUnlocked{
			ID:    achID,
			Title: titleByID[achID],
		})
	}

	return inserted, nil
}
