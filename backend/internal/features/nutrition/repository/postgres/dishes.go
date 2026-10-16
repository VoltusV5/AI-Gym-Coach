package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"

	"github.com/jackc/pgx/v5/pgconn"
)

// SearchDishes searches dishes in the catalog by substring match with stable ordering.
func (r *Repository) SearchDishes(
	ctx context.Context,
	searchQuery string,
	limit int,
	offset int,
) ([]CatalogDish, int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	pattern := "%" + strings.ToLower(searchQuery) + "%"

	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM sportapp.nutrition_dishes
		WHERE ($1 = '' OR LOWER(title) LIKE $2)`
	if err := r.pool.QueryRow(ctx, countQuery, searchQuery, pattern).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count dishes: %w", err)
	}

	query := `
		SELECT id, title, protein_g, fat_g, carbs_g, calories, base_grams
		FROM sportapp.nutrition_dishes
		WHERE ($1 = '' OR LOWER(title) LIKE $2)
		ORDER BY title ASC, id ASC`

	var args []any
	if limit > 0 {
		query += " LIMIT $3 OFFSET $4"
		args = []any{searchQuery, pattern, limit, offset}
	} else {
		args = []any{searchQuery, pattern}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("search dishes: %w", err)
	}
	defer rows.Close()

	items := make([]CatalogDish, 0, limit)
	for rows.Next() {
		var item CatalogDish
		if err := rows.Scan(
			&item.ID, &item.Title, &item.ProteinG, &item.FatG,
			&item.CarbsG, &item.Calories, &item.BaseGrams,
		); err != nil {
			return nil, 0, fmt.Errorf("scan dish: %w", err)
		}
		items = append(items, item)
	}

	return items, total, nil
}

// GetDishByID retrieves a dish by its ID.
func (r *Repository) GetDishByID(
	ctx context.Context,
	dishID int,
) (CatalogDish, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, title, protein_g, fat_g, carbs_g, calories, base_grams
		FROM sportapp.nutrition_dishes
		WHERE id = $1`

	var item CatalogDish
	err := r.pool.QueryRow(ctx, query, dishID).Scan(
		&item.ID, &item.Title, &item.ProteinG, &item.FatG,
		&item.CarbsG, &item.Calories, &item.BaseGrams,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return CatalogDish{}, fmt.Errorf("dish id='%d' not found: %w", dishID, core_errors.ErrNotFound)
		}
		return CatalogDish{}, fmt.Errorf("query dish by id: %w", err)
	}

	return item, nil
}

// CreateOrUpsertDish inserts a new dish or updates macros on conflict title.
func (r *Repository) CreateOrUpsertDish(
	ctx context.Context,
	userID int,
	dish CatalogDish,
) (CatalogDish, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO sportapp.nutrition_dishes
			(title, protein_g, fat_g, carbs_g, calories, base_grams, created_by_user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (title) DO UPDATE
		SET protein_g = EXCLUDED.protein_g,
		    fat_g = EXCLUDED.fat_g,
		    carbs_g = EXCLUDED.carbs_g,
		    calories = EXCLUDED.calories,
		    base_grams = EXCLUDED.base_grams,
		    created_by_user_id = COALESCE(sportapp.nutrition_dishes.created_by_user_id, EXCLUDED.created_by_user_id),
		    updated_at = NOW()
		RETURNING id, title, protein_g, fat_g, carbs_g, calories, base_grams`

	var item CatalogDish
	err := r.pool.QueryRow(
		ctx, query,
		dish.Title, dish.ProteinG, dish.FatG, dish.CarbsG,
		dish.Calories, dish.BaseGrams, userID,
	).Scan(
		&item.ID, &item.Title, &item.ProteinG, &item.FatG,
		&item.CarbsG, &item.Calories, &item.BaseGrams,
	)
	if err != nil {
		var pe *pgconn.PgError
		if errors.As(err, &pe) && pe.Code == "23505" {
			return CatalogDish{}, fmt.Errorf("dish title already exists: %w", core_errors.ErrConflict)
		}
		return CatalogDish{}, fmt.Errorf("upsert dish: %w", err)
	}

	return item, nil
}

// ListMyDishes retrieves custom dishes created by a specific user.
func (r *Repository) ListMyDishes(
	ctx context.Context,
	userID int,
	limit int,
	offset int,
) ([]CatalogDish, int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var total int
	countQuery := `SELECT COUNT(*) FROM sportapp.nutrition_dishes WHERE created_by_user_id = $1`
	if err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count my dishes: %w", err)
	}

	query := `
		SELECT id, title, protein_g, fat_g, carbs_g, calories, base_grams
		FROM sportapp.nutrition_dishes
		WHERE created_by_user_id = $1
		ORDER BY COALESCE(updated_at, created_at) DESC, id DESC`

	var args []any
	if limit > 0 {
		query += " LIMIT $2 OFFSET $3"
		args = []any{userID, limit, offset}
	} else {
		args = []any{userID}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list my dishes: %w", err)
	}
	defer rows.Close()

	items := make([]CatalogDish, 0, 32)
	for rows.Next() {
		var item CatalogDish
		if err := rows.Scan(
			&item.ID, &item.Title, &item.ProteinG, &item.FatG,
			&item.CarbsG, &item.Calories, &item.BaseGrams,
		); err != nil {
			return nil, 0, fmt.Errorf("scan my dish: %w", err)
		}
		items = append(items, item)
	}

	return items, total, nil
}

// UpdateMyDish updates a custom dish created by the user.
func (r *Repository) UpdateMyDish(
	ctx context.Context,
	userID int,
	dishID int,
	dish CatalogDish,
) (CatalogDish, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE sportapp.nutrition_dishes
		SET title = $1, protein_g = $2, fat_g = $3, carbs_g = $4,
		    calories = $5, base_grams = $6, updated_at = NOW()
		WHERE id = $7 AND created_by_user_id = $8
		RETURNING id, title, protein_g, fat_g, carbs_g, calories, base_grams`

	var item CatalogDish
	err := r.pool.QueryRow(
		ctx, query,
		dish.Title, dish.ProteinG, dish.FatG, dish.CarbsG,
		dish.Calories, dish.BaseGrams, dishID, userID,
	).Scan(
		&item.ID, &item.Title, &item.ProteinG, &item.FatG,
		&item.CarbsG, &item.Calories, &item.BaseGrams,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return CatalogDish{}, fmt.Errorf("dish id='%d' not found or not owned: %w", dishID, core_errors.ErrNotFound)
		}
		var pe *pgconn.PgError
		if errors.As(err, &pe) && pe.Code == "23505" {
			return CatalogDish{}, fmt.Errorf("dish title already exists: %w", core_errors.ErrConflict)
		}
		return CatalogDish{}, fmt.Errorf("update my dish: %w", err)
	}

	return item, nil
}

// DeleteMyDish detaches user entries from the dish and deletes the dish.
func (r *Repository) DeleteMyDish(
	ctx context.Context,
	userID int,
	dishID int,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	detachQuery := `
		UPDATE sportapp.nutrition_entries
		SET dish_id = NULL
		WHERE user_id = $1 AND dish_id = $2`
	if _, err := r.pool.Exec(ctx, detachQuery, userID, dishID); err != nil {
		return fmt.Errorf("detach entries from dish: %w", err)
	}

	deleteQuery := `
		DELETE FROM sportapp.nutrition_dishes
		WHERE id = $1 AND created_by_user_id = $2`
	res, err := r.pool.Exec(ctx, deleteQuery, dishID, userID)
	if err != nil {
		return fmt.Errorf("delete dish: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("dish id='%d' not found or not owned: %w", dishID, core_errors.ErrNotFound)
	}

	return nil
}
