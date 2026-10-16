package postgres

import (
	"context"
	"fmt"
	"time"

	core_errors "sport_app/internal/core/errors"
)

// ListEntries fetches nutrition entries for a user within a time range with stable ordering.
func (r *Repository) ListEntries(
	ctx context.Context,
	userID int,
	startUTC time.Time,
	endUTC time.Time,
	limit int,
	offset int,
) ([]NutritionEntry, int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM sportapp.nutrition_entries
		WHERE user_id = $1 AND consumed_at >= $2 AND consumed_at < $3`
	if err := r.pool.QueryRow(ctx, countQuery, userID, startUTC, endUTC).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count nutrition entries: %w", err)
	}

	query := `
		SELECT id, dish_id, grams, meal_type, title, protein_g, fat_g, carbs_g, calories, consumed_at, created_at
		FROM sportapp.nutrition_entries
		WHERE user_id = $1 AND consumed_at >= $2 AND consumed_at < $3
		ORDER BY consumed_at DESC, id DESC`

	var args []any
	if limit > 0 {
		query += " LIMIT $4 OFFSET $5"
		args = []any{userID, startUTC, endUTC, limit, offset}
	} else {
		args = []any{userID, startUTC, endUTC}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query nutrition entries: %w", err)
	}
	defer rows.Close()

	items := make([]NutritionEntry, 0, 128)
	for rows.Next() {
		var it NutritionEntry
		if err := rows.Scan(
			&it.ID, &it.DishID, &it.Grams, &it.MealType, &it.Title,
			&it.ProteinG, &it.FatG, &it.CarbsG, &it.Calories,
			&it.ConsumedAt, &it.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan nutrition entry: %w", err)
		}
		items = append(items, it)
	}

	return items, total, nil
}

// CreateEntry inserts a new nutrition entry into the database.
func (r *Repository) CreateEntry(
	ctx context.Context,
	userID int,
	entry NutritionEntry,
) (NutritionEntry, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO sportapp.nutrition_entries
			(user_id, dish_id, grams, meal_type, title, protein_g, fat_g, carbs_g, calories, consumed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at`

	var id int
	var createdAt time.Time
	err := r.pool.QueryRow(
		ctx, query,
		userID, entry.DishID, entry.Grams, entry.MealType, entry.Title,
		entry.ProteinG, entry.FatG, entry.CarbsG, entry.Calories, entry.ConsumedAt,
	).Scan(&id, &createdAt)
	if err != nil {
		return NutritionEntry{}, fmt.Errorf("insert nutrition entry: %w", err)
	}

	entry.ID = id
	entry.CreatedAt = createdAt
	return entry, nil
}

// UpdateEntry updates an existing nutrition entry owned by the user.
func (r *Repository) UpdateEntry(
	ctx context.Context,
	userID int,
	entryID int,
	entry NutritionEntry,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE sportapp.nutrition_entries
		SET grams = $1, meal_type = $2, title = $3, protein_g = $4, fat_g = $5,
		    carbs_g = $6, calories = $7, consumed_at = $8, updated_at = NOW()
		WHERE id = $9 AND user_id = $10`

	tag, err := r.pool.Exec(
		ctx, query,
		entry.Grams, entry.MealType, entry.Title, entry.ProteinG, entry.FatG,
		entry.CarbsG, entry.Calories, entry.ConsumedAt, entryID, userID,
	)
	if err != nil {
		return fmt.Errorf("update nutrition entry: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("nutrition entry id='%d' not found: %w", entryID, core_errors.ErrNotFound)
	}

	return nil
}

// DeleteEntry deletes a nutrition entry owned by the user.
func (r *Repository) DeleteEntry(
	ctx context.Context,
	userID int,
	entryID int,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `DELETE FROM sportapp.nutrition_entries WHERE id = $1 AND user_id = $2`
	tag, err := r.pool.Exec(ctx, query, entryID, userID)
	if err != nil {
		return fmt.Errorf("delete nutrition entry: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("nutrition entry id='%d' not found: %w", entryID, core_errors.ErrNotFound)
	}

	return nil
}
