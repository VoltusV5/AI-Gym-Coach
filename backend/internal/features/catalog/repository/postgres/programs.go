package catalog_postgres_repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	core_errors "sport_app/internal/core/errors"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"

	"github.com/google/uuid"
)

const cardJSON = `jsonb_build_object(
	'id', p.id, 'title', v.title,
	'trainer', jsonb_build_object('id', t.user_id, 'display_name', t.display_name, 'verified', t.verified_at IS NOT NULL),
	'goal', v.goal, 'level', v.level, 'duration_weeks', v.duration_weeks,
	'sessions_per_week', v.sessions_per_week, 'price_minor', v.price_minor, 'currency', v.currency,
	'rating', NULL, 'review_count', 0, 'published_version_id', v.id,
	'equipment', COALESCE((SELECT jsonb_agg(jsonb_build_object('id', e.id, 'code', e.code, 'name', e.name) ORDER BY e.id)
		FROM sportapp.program_version_equipment ve JOIN sportapp.equipment e ON e.id = ve.equipment_id
		WHERE ve.version_id = v.id), '[]'::jsonb))`

const publishedFrom = ` FROM sportapp.programs p
	JOIN sportapp.program_versions v ON v.id = p.published_version_id AND v.program_id = p.id
	JOIN sportapp.trainer_profiles t ON t.user_id = p.trainer_id
	WHERE p.status = 'active' AND v.status = 'published'`

func (r *Repository) ListPrograms(ctx context.Context, filter Filter) ([]ProgramCard, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	order := "p.created_at DESC, p.id DESC"
	after := `(p.created_at, p.id) < ($10::timestamptz, $1::uuid)`
	var sortValue any = filter.AfterCreatedAt
	switch filter.Sort {
	case "price_asc":
		order = "v.price_minor ASC, p.id ASC"
		after = `(v.price_minor, p.id) > ($10::bigint, $1::uuid)`
		sortValue = filter.AfterPrice
	case "price_desc":
		order = "v.price_minor DESC, p.id DESC"
		after = `(v.price_minor, p.id) < ($10::bigint, $1::uuid)`
		sortValue = filter.AfterPrice
	}
	query := `SELECT ` + cardJSON + `, p.created_at` + publishedFrom + `
		AND ($1::uuid IS NULL OR (p.id <> $1::uuid AND ` + after + `))
		AND ($2 = '' OR v.goal = $2) AND ($3 = '' OR v.level = $3)
		AND ($4 = 0 OR v.duration_weeks = $4) AND ($5 = 0 OR p.trainer_id = $5)
		AND ($6 = '' OR ($6 = 'free' AND v.price_minor = 0) OR ($6 = 'paid' AND v.price_minor > 0))
		AND (NOT $7 OR NOT EXISTS (SELECT 1 FROM sportapp.program_version_equipment ve
			WHERE ve.version_id = v.id AND NOT (ve.equipment_id = ANY($8::text[]))))
		ORDER BY ` + order + ` LIMIT $9`
	var cursor any
	if filter.After != "" {
		cursor = filter.After
	}
	rows, err := r.pool.Query(ctx, query, cursor, filter.Goal, filter.Level,
		filter.DurationWeeks, filter.TrainerID, filter.PriceType, filter.EquipmentSet, filter.Equipment, filter.Limit+1,
		sortValue)
	if err != nil {
		return nil, fmt.Errorf("list programs: %w", err)
	}
	defer rows.Close()
	items := make([]ProgramCard, 0, filter.Limit+1)
	for rows.Next() {
		var raw []byte
		var item ProgramCard
		if err := rows.Scan(&raw, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan program: %w", err)
		}
		if err := json.Unmarshal(raw, &item); err != nil {
			return nil, fmt.Errorf("decode program: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read programs: %w", err)
	}
	return items, nil
}

func (r *Repository) GetProgram(ctx context.Context, programID string, userID int) (ProgramDetail, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	query := `SELECT ` + cardJSON + ` || jsonb_build_object(
		'description', v.description, 'allow_public_remix', v.allow_public_remix,
		'access', CASE WHEN a.allowed THEN CASE WHEN v.price_minor = 0 THEN 'free' ELSE 'paid' END ELSE 'preview' END,
		'allowed_actions', CASE WHEN a.allowed THEN '[]'::jsonb
			WHEN v.price_minor = 0 THEN '["access"]'::jsonb ELSE '[]'::jsonb END,
		'days', CASE WHEN a.allowed THEN COALESCE((SELECT jsonb_agg(jsonb_build_object(
			'id', d.id, 'title', d.title, 'position', d.position,
			'exercises', COALESCE((SELECT jsonb_agg(jsonb_build_object(
				'exercise_id', ex.exercise_id, 'position', ex.position, 'sets', ex.sets,
				'reps_min', ex.reps_min, 'reps_max', ex.reps_max, 'weight_kg', ex.weight_kg, 'rest_sec', ex.rest_sec)
				ORDER BY ex.position) FROM sportapp.program_exercises ex WHERE ex.program_day_id = d.id), '[]'::jsonb))
			ORDER BY d.position) FROM sportapp.program_days d WHERE d.program_version_id = v.id), '[]'::jsonb)
			ELSE '[]'::jsonb END)
		FROM sportapp.programs p
		JOIN sportapp.program_versions v ON v.id = p.published_version_id AND v.program_id = p.id
		JOIN sportapp.trainer_profiles t ON t.user_id = p.trainer_id
		CROSS JOIN LATERAL (SELECT EXISTS (SELECT 1 FROM sportapp.entitlements en
			WHERE en.program_id = p.id AND en.user_id = $2 AND en.status = 'active') AS allowed) a
		WHERE p.id = $1 AND p.status = 'active' AND v.status = 'published'`
	var raw []byte
	if err := r.pool.QueryRow(ctx, query, programID, userID).Scan(&raw); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return ProgramDetail{}, fmt.Errorf("program not found: %w", core_errors.ErrNotFound)
		}
		return ProgramDetail{}, fmt.Errorf("get program: %w", err)
	}
	var item ProgramDetail
	if err := json.Unmarshal(raw, &item); err != nil {
		return ProgramDetail{}, fmt.Errorf("decode program detail: %w", err)
	}
	return item, nil
}

func (r *Repository) GrantFreeAccess(ctx context.Context, programID string, userID int) (Entitlement, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	query := `WITH available AS MATERIALIZED (
		SELECT p.id, v.price_minor FROM sportapp.programs p
		JOIN sportapp.program_versions v ON v.id = p.published_version_id AND v.program_id = p.id
		WHERE p.id = $1 AND p.status = 'active' AND v.status = 'published' FOR SHARE OF p, v
	), member AS MATERIALIZED (
		SELECT id, is_anonymous FROM sportapp.users WHERE id = $2 FOR SHARE
	), granted AS (
		INSERT INTO sportapp.entitlements (id, user_id, program_id, status)
		SELECT $3, m.id, a.id, 'active' FROM available a CROSS JOIN member m
		WHERE a.price_minor = 0 AND NOT m.is_anonymous
		ON CONFLICT (user_id, program_id) DO UPDATE SET status = 'active',
			granted_at = CASE WHEN sportapp.entitlements.status = 'active' THEN sportapp.entitlements.granted_at ELSE NOW() END
		RETURNING id, program_id, status, granted_at, order_id
	)
	SELECT EXISTS(SELECT 1 FROM available), EXISTS(SELECT 1 FROM member),
		COALESCE((SELECT is_anonymous FROM member), TRUE),
		COALESCE((SELECT price_minor FROM available), 0),
		(SELECT row_to_json(g) FROM granted g)`
	var available, member, anonymous bool
	var price int64
	var raw []byte
	err := r.pool.QueryRow(ctx, query, programID, userID, uuid.NewString()).Scan(&available, &member, &anonymous, &price, &raw)
	if err != nil {
		return Entitlement{}, fmt.Errorf("grant access: %w", err)
	}
	if !member {
		return Entitlement{}, fmt.Errorf("user not found: %w", core_errors.ErrUnauthorized)
	}
	if anonymous {
		return Entitlement{}, fmt.Errorf("registration required: %w", core_errors.ErrForbidden)
	}
	if !available {
		return Entitlement{}, fmt.Errorf("program not found: %w", core_errors.ErrNotFound)
	}
	if price != 0 {
		return Entitlement{}, fmt.Errorf("program is not free: %w", core_errors.ErrUnprocessable)
	}
	var item Entitlement
	if err := json.Unmarshal(raw, &item); err != nil {
		return Entitlement{}, fmt.Errorf("decode entitlement: %w", err)
	}
	return item, nil
}
