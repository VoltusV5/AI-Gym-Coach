package users_postgres_repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	core_errors "sport_app/internal/core/errors"
)

func (r *UsersRepository) GetWorkingWeights(
	ctx context.Context,
	userID string,
) (map[string]float64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var raw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT working_weights FROM sportapp.user_data WHERE user_id = $1
	`, userID).Scan(&raw)
	if err != nil {
		return nil, fmt.Errorf("select working_weights: %w", err)
	}
	if len(raw) == 0 || string(raw) == "null" {
		return map[string]float64{}, nil
	}

	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return map[string]float64{}, nil
	}

	out := make(map[string]float64, len(m))
	for k, v := range m {
		switch val := v.(type) {
		case float64:
			out[k] = val
		case int:
			out[k] = float64(val)
		case int64:
			out[k] = float64(val)
		}
	}
	return out, nil
}

func (r *UsersRepository) SaveWorkingWeights(
	ctx context.Context,
	userID string,
	expectedVersion int64,
	workingWeights []byte,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	UPDATE sportapp.user_data
	SET working_weights = $1,
	    updated_at = $2,
	    version = version + 1
	WHERE user_id = $3 AND version = $4
	`

	tag, err := r.pool.Exec(ctx, query, workingWeights, time.Now(), userID, expectedVersion)
	if err != nil {
		return fmt.Errorf("update working_weights: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf(
			"user_data for user_id='%s' concurrently accessed or not found: %w",
			userID,
			core_errors.ErrConflict,
		)
	}

	return nil
}

func ApplyExistingWeightsToPlan(plan *EPlanWithWeight, existing map[string]float64) {
	if len(existing) == 0 {
		return
	}
	for i := range plan.Plan {
		for j := range plan.Plan[i].Exercises {
			for k := range plan.Plan[i].Exercises[j] {
				ex := &plan.Plan[i].Exercises[j][k]
				key := strconv.Itoa(ex.ID)
				if w, ok := existing[key]; ok {
					wCopy := w
					ex.Weight = &wCopy
				}
			}
		}
	}
}

func MergeWorkingWeightsJSON(existing map[string]float64, plan EPlanWithWeight) ([]byte, error) {
	merged := make(map[string]float64, len(existing)+64)
	for k, v := range existing {
		merged[k] = v
	}
	for _, day := range plan.Plan {
		for _, list := range day.Exercises {
			for _, ex := range list {
				if ex.Weight == nil {
					continue
				}
				key := strconv.Itoa(ex.ID)
				if _, ok := merged[key]; !ok {
					merged[key] = *ex.Weight
				}
			}
		}
	}
	data, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("marshal working_weights: %w", err)
	}
	return data, nil
}
