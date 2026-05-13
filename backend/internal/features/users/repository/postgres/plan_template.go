package users_postgres_repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
)

func (r *UsersRepository) GetPlanTemplateJSON(
	ctx context.Context,
	userID string,
) (json.RawMessage, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var raw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(plan_template, '{}'::jsonb)
		FROM sportapp.user_programs
		WHERE user_id = $1
	`, userID).Scan(&raw)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return json.RawMessage(`{}`), nil
		}
		return nil, fmt.Errorf("select plan_template: %w", err)
	}
	if len(raw) == 0 {
		return json.RawMessage(`{}`), nil
	}
	return json.RawMessage(raw), nil
}