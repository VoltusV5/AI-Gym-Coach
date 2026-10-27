package catalog_postgres_repository

import (
	"context"
	"fmt"
	"strconv"
)

func (r *Repository) ListEquipment(ctx context.Context, filter Filter) ([]Equipment, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	rows, err := r.pool.Query(ctx, `SELECT id, code, name FROM sportapp.equipment
		WHERE id > $1 ORDER BY id LIMIT $2`, filter.After, filter.Limit+1)
	if err != nil {
		return nil, fmt.Errorf("list equipment: %w", err)
	}
	defer rows.Close()
	items := make([]Equipment, 0, filter.Limit+1)
	for rows.Next() {
		var item Equipment
		if err := rows.Scan(&item.ID, &item.Code, &item.Name); err != nil {
			return nil, fmt.Errorf("scan equipment: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read equipment: %w", err)
	}
	return items, nil
}

func (r *Repository) ListExercises(ctx context.Context, filter Filter) ([]Exercise, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	after, _ := strconv.Atoi(filter.After)
	rows, err := r.pool.Query(ctx, `SELECT e.id, COALESCE(e.exercises_name, ''),
		COALESCE(e.muscular_group, ''), e.muscular_subgroup,
		ARRAY(SELECT equipment_id FROM sportapp.exercise_equipment WHERE exercise_id = e.id ORDER BY equipment_id),
		COALESCE(e.safe_for_injuries, FALSE), e.image_url, e.video_url, e.is_active
		FROM sportapp.exercises e WHERE e.is_active AND e.id > $1
		AND ($2 = '' OR strpos(lower(e.exercises_name), lower($2)) > 0)
		AND ($3 = '' OR e.muscular_group = $3)
		AND (NOT $4 OR NOT EXISTS (SELECT 1 FROM sportapp.exercise_equipment ee
			WHERE ee.exercise_id = e.id AND NOT (ee.equipment_id = ANY($5::text[]))))
		ORDER BY e.id LIMIT $6`, after, filter.Query, filter.MuscleGroup,
		filter.EquipmentSet, filter.Equipment, filter.Limit+1)
	if err != nil {
		return nil, fmt.Errorf("list exercises: %w", err)
	}
	defer rows.Close()
	items := make([]Exercise, 0, filter.Limit+1)
	for rows.Next() {
		var item Exercise
		if err := rows.Scan(&item.ID, &item.Name, &item.MuscleGroup, &item.MuscleSubgroup,
			&item.EquipmentIDs, &item.SafeForInjuries, &item.ImageURL, &item.VideoURL, &item.IsActive); err != nil {
			return nil, fmt.Errorf("scan exercise: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read exercises: %w", err)
	}
	return items, nil
}

func (r *Repository) EquipmentExists(ctx context.Context, ids []string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT NOT EXISTS (SELECT 1 FROM unnest($1::text[]) AS candidate(id)
		WHERE NOT EXISTS (SELECT 1 FROM sportapp.equipment e WHERE e.id = candidate.id))`, ids).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check equipment: %w", err)
	}
	return exists, nil
}
