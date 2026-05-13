package users_postgres_repository

import (
	"context"
	"fmt"
	"strings"
)

func (r *UsersRepository) GetExerciseNamesByIDs(
	ctx context.Context,
	ids []int,
) (map[int]string, error) {
	out := make(map[int]string, len(ids))
	if len(ids) == 0 {
		return out, nil
	}

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	args := make([]any, len(ids))
	ph := make([]string, len(ids))
	for i, id := range ids {
		args[i] = id
		ph[i] = fmt.Sprintf("$%d", i+1)
	}
	q := `SELECT id, COALESCE(exercises_name, '') FROM sportapp.exercises WHERE id IN (` +
		strings.Join(ph, ",") + `)`

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query exercise names: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf("scan exercise name: %w", err)
		}
		out[id] = name
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate exercise names: %w", err)
	}

	return out, nil
}