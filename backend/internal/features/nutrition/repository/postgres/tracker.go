package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
)

// UpsertWater records water consumption for a user on a given date.
func (r *Repository) UpsertWater(
	ctx context.Context,
	userID int,
	loggedDay string,
	amountML int,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO sportapp.nutrition_water_logs (user_id, logged_on, amount_ml)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, logged_on) DO UPDATE
		SET amount_ml = EXCLUDED.amount_ml, updated_at = NOW()`

	_, err := r.pool.Exec(ctx, query, userID, loggedDay, amountML)
	if err != nil {
		return fmt.Errorf("upsert water: %w", err)
	}

	return nil
}

// UpsertWeight records body weight for a user on a given date.
func (r *Repository) UpsertWeight(
	ctx context.Context,
	userID int,
	loggedDay string,
	weightKg float64,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO sportapp.nutrition_weight_logs (user_id, logged_on, weight_kg)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, logged_on) DO UPDATE
		SET weight_kg = EXCLUDED.weight_kg, updated_at = NOW()`

	_, err := r.pool.Exec(ctx, query, userID, loggedDay, weightKg)
	if err != nil {
		return fmt.Errorf("upsert weight: %w", err)
	}

	return nil
}

// GetDayTotals calculates total macronutrients for a given date range.
func (r *Repository) GetDayTotals(
	ctx context.Context,
	userID int,
	startUTC time.Time,
	endUTC time.Time,
) (DailyMacroTotals, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT COALESCE(SUM(protein_g), 0),
		       COALESCE(SUM(fat_g), 0),
		       COALESCE(SUM(carbs_g), 0),
		       COALESCE(SUM(calories), 0)
		FROM sportapp.nutrition_entries
		WHERE user_id = $1 AND consumed_at >= $2 AND consumed_at < $3`

	var t DailyMacroTotals
	err := r.pool.QueryRow(ctx, query, userID, startUTC, endUTC).Scan(
		&t.Protein, &t.Fat, &t.Carbs, &t.Calories,
	)
	if err != nil {
		return DailyMacroTotals{}, fmt.Errorf("get day totals: %w", err)
	}

	return t, nil
}

// GetStreakDays computes the number of consecutive days with logged activity.
func (r *Repository) GetStreakDays(
	ctx context.Context,
	userID int,
) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		WITH days AS (
			SELECT d::date AS day
			FROM generate_series(CURRENT_DATE - INTERVAL '180 days', CURRENT_DATE, '1 day') d
		),
		filled AS (
			SELECT day,
			       (
			         EXISTS (SELECT 1 FROM sportapp.nutrition_entries e WHERE e.user_id = $1 AND DATE(e.consumed_at) = day)
			         OR EXISTS (SELECT 1 FROM sportapp.nutrition_water_logs w WHERE w.user_id = $1 AND w.logged_on = day)
			         OR EXISTS (SELECT 1 FROM sportapp.nutrition_weight_logs wl WHERE wl.user_id = $1 AND wl.logged_on = day)
			       ) AS is_filled
			FROM days
		),
		latest_break AS (
			SELECT COALESCE(MAX(day), CURRENT_DATE - INTERVAL '365 days') AS last_break
			FROM filled
			WHERE is_filled = false AND day <= CURRENT_DATE
		)
		SELECT COUNT(*)
		FROM filled f, latest_break b
		WHERE f.is_filled = true AND f.day > b.last_break`

	var streak int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&streak)
	if err != nil {
		return 0, nil
	}

	return streak, nil
}

// GetWaterLogML returns the amount of water logged for a specific date.
func (r *Repository) GetWaterLogML(
	ctx context.Context,
	userID int,
	loggedDay string,
) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `SELECT amount_ml FROM sportapp.nutrition_water_logs WHERE user_id = $1 AND logged_on = $2`
	var waterML int
	err := r.pool.QueryRow(ctx, query, userID, loggedDay).Scan(&waterML)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("get water log: %w", err)
	}

	return waterML, nil
}

// GetLatestWeightInfo returns the latest weight and date it was logged.
func (r *Repository) GetLatestWeightInfo(
	ctx context.Context,
	userID int,
) (*float64, *time.Time, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT weight_kg, logged_on::timestamp
		FROM sportapp.nutrition_weight_logs
		WHERE user_id = $1
		ORDER BY logged_on DESC
		LIMIT 1`

	var weight *float64
	var loggedDay *time.Time
	err := r.pool.QueryRow(ctx, query, userID).Scan(&weight, &loggedDay)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("get latest weight info: %w", err)
	}

	return weight, loggedDay, nil
}

// GetRangeSums aggregates macronutrients starting from a given timestamp.
func (r *Repository) GetRangeSums(
	ctx context.Context,
	userID int,
	from time.Time,
) (DailyMacroTotals, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT COALESCE(SUM(protein_g), 0),
		       COALESCE(SUM(fat_g), 0),
		       COALESCE(SUM(carbs_g), 0),
		       COALESCE(SUM(calories), 0)
		FROM sportapp.nutrition_entries
		WHERE user_id = $1 AND consumed_at >= $2`

	var acc DailyMacroTotals
	err := r.pool.QueryRow(ctx, query, userID, from).Scan(
		&acc.Protein, &acc.Fat, &acc.Carbs, &acc.Calories,
	)
	if err != nil {
		return DailyMacroTotals{}, fmt.Errorf("get range sums: %w", err)
	}

	return acc, nil
}

// GetAnalyticsFood aggregates food entries grouped by day.
func (r *Repository) GetAnalyticsFood(
	ctx context.Context,
	userID int,
	daysLimit int,
) ([]FoodDaySummary, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT DATE(consumed_at) as d,
		       COALESCE(SUM(protein_g), 0),
		       COALESCE(SUM(fat_g), 0),
		       COALESCE(SUM(carbs_g), 0),
		       COALESCE(SUM(calories), 0)
		FROM sportapp.nutrition_entries
		WHERE user_id = $1 AND consumed_at >= NOW() - ($2::int * interval '1 day')
		GROUP BY d
		ORDER BY d DESC`

	rows, err := r.pool.Query(ctx, query, userID, daysLimit)
	if err != nil {
		return nil, fmt.Errorf("query analytics food: %w", err)
	}
	defer rows.Close()

	items := make([]FoodDaySummary, 0, daysLimit)
	for rows.Next() {
		var d time.Time
		var p, f, c, kcal float64
		if err := rows.Scan(&d, &p, &f, &c, &kcal); err != nil {
			return nil, fmt.Errorf("scan analytics food: %w", err)
		}
		items = append(items, FoodDaySummary{
			Day:      d.Format("2006-01-02"),
			ProteinG: p,
			FatG:     f,
			CarbsG:   c,
			Calories: kcal,
		})
	}

	return items, nil
}

// GetAnalyticsWeight returns weight logs for analytics report.
func (r *Repository) GetAnalyticsWeight(
	ctx context.Context,
	userID int,
	daysLimit int,
) ([]WeightLogSummary, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT logged_on, weight_kg
		FROM sportapp.nutrition_weight_logs
		WHERE user_id = $1 AND logged_on >= CURRENT_DATE - ($2::int * interval '1 day')
		ORDER BY logged_on DESC`

	rows, err := r.pool.Query(ctx, query, userID, daysLimit)
	if err != nil {
		return nil, fmt.Errorf("query analytics weight: %w", err)
	}
	defer rows.Close()

	items := make([]WeightLogSummary, 0, daysLimit)
	for rows.Next() {
		var day time.Time
		var weight float64
		if err := rows.Scan(&day, &weight); err != nil {
			return nil, fmt.Errorf("scan analytics weight: %w", err)
		}
		items = append(items, WeightLogSummary{
			Day:      day.Format("2006-01-02"),
			WeightKg: weight,
		})
	}

	return items, nil
}

// GetAnalyticsWater returns water intake logs for analytics report.
func (r *Repository) GetAnalyticsWater(
	ctx context.Context,
	userID int,
	daysLimit int,
) ([]WaterLogSummary, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT logged_on, amount_ml
		FROM sportapp.nutrition_water_logs
		WHERE user_id = $1 AND logged_on >= CURRENT_DATE - ($2::int * interval '1 day')
		ORDER BY logged_on DESC`

	rows, err := r.pool.Query(ctx, query, userID, daysLimit)
	if err != nil {
		return nil, fmt.Errorf("query analytics water: %w", err)
	}
	defer rows.Close()

	items := make([]WaterLogSummary, 0, daysLimit)
	for rows.Next() {
		var day time.Time
		var ml int
		if err := rows.Scan(&day, &ml); err != nil {
			return nil, fmt.Errorf("scan analytics water: %w", err)
		}
		items = append(items, WaterLogSummary{
			Day:      day.Format("2006-01-02"),
			AmountML: ml,
		})
	}

	return items, nil
}
