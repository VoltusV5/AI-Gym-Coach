package postgres

import (
	"context"
	"fmt"

	core_errors "sport_app/internal/core/errors"
)

// ListFavorites возвращает избранные продукты пользователя.
func (r *Repository) ListFavorites(
	ctx context.Context,
	userID int,
	limit int,
	offset int,
) ([]NutritionFavorite, int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var total int
	countQuery := `SELECT COUNT(*) FROM sportapp.nutrition_favorites WHERE user_id = $1`
	if err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count nutrition favorites: %w", err)
	}

	query := `
		SELECT id, title, protein_g, fat_g, carbs_g, unit_type
		FROM sportapp.nutrition_favorites
		WHERE user_id = $1
		ORDER BY id DESC`

	var args []any
	if limit > 0 {
		query += " LIMIT $2 OFFSET $3"
		args = []any{userID, limit, offset}
	} else {
		args = []any{userID}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query nutrition favorites: %w", err)
	}
	defer rows.Close()

	items := make([]NutritionFavorite, 0, 64)
	for rows.Next() {
		var it NutritionFavorite
		if err := rows.Scan(&it.ID, &it.Title, &it.ProteinG, &it.FatG, &it.CarbsG, &it.UnitType); err != nil {
			return nil, 0, fmt.Errorf("scan nutrition favorite: %w", err)
		}
		items = append(items, it)
	}

	return items, total, nil
}

// CreateFavorite сохраняет продукт в избранном пользователя.
func (r *Repository) CreateFavorite(
	ctx context.Context,
	userID int,
	fav NutritionFavorite,
) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO sportapp.nutrition_favorites (user_id, title, protein_g, fat_g, carbs_g, unit_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	var id int
	err := r.pool.QueryRow(
		ctx, query,
		userID, fav.Title, fav.ProteinG, fav.FatG, fav.CarbsG, fav.UnitType,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert nutrition favorite: %w", err)
	}

	return id, nil
}

// DeleteFavorite удаляет продукт из избранного только его владельца.
func (r *Repository) DeleteFavorite(
	ctx context.Context,
	userID int,
	favID int,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `DELETE FROM sportapp.nutrition_favorites WHERE id = $1 AND user_id = $2`
	tag, err := r.pool.Exec(ctx, query, favID, userID)
	if err != nil {
		return fmt.Errorf("delete nutrition favorite: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("favorite id='%d' not found: %w", favID, core_errors.ErrNotFound)
	}

	return nil
}
