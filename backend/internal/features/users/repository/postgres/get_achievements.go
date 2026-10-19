package users_postgres_repository

import (
	"context"
	"fmt"
)

func (r *UsersRepository) GetAchievements(
	ctx context.Context,
	userID string,
) ([]UserAchievement, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	rows, err := r.pool.Query(ctx, `
		SELECT a.id, a.title, a.description, a.category, ua.unlocked_at
		FROM sportapp.user_achievements ua
		INNER JOIN sportapp.achievements a ON a.id = ua.achievement_id
		WHERE ua.user_id = $1
		ORDER BY ua.unlocked_at ASC, a.id ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("query user achievements: %w", err)
	}
	defer rows.Close()

	var list []UserAchievement
	for rows.Next() {
		var ua UserAchievement
		if err := rows.Scan(
			&ua.ID,
			&ua.Title,
			&ua.Description,
			&ua.Category,
			&ua.UnlockedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user achievement: %w", err)
		}
		list = append(list, ua)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user achievements: %w", err)
	}

	return list, nil
}
