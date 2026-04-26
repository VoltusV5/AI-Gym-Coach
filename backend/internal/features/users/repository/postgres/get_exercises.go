package users_postgres_repository

import (
	"context"
	"fmt"
	"math"

	"sport_app/internal/features/mlclient"

	"github.com/jackc/pgx/v5"
)

func (r *UsersRepository) GetExercises(
	ctx context.Context,
	plan mlclient.Plan,
	userID string,
) (EPlanWithWeight, EPlanNoWeight, error) {
	profile, err := r.GetProfile(ctx, userID)
	if err != nil {
		return EPlanWithWeight{}, EPlanNoWeight{}, fmt.Errorf("get profile for weights: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	planWithWeight := EPlanWithWeight{
		Split: plan.Split_type,
		Plan:  make([]EDaysWithWeight, len(plan.Plan_week)),
	}
	planNoWeight := EPlanNoWeight{
		Split: plan.Split_type,
		Plan:  make([]EDaysNoWeight, len(plan.Plan_week)),
	}

	for i, day := range plan.Plan_week {
		dayWithWeight := EDaysWithWeight{
			Day:       day.Day,
			DName:     day.Type_day,
			Exercises: make([][]ExWithWeight, len(day.Exercises)),
		}
		dayNoWeight := EDaysNoWeight{
			Day:       day.Day,
			DName:     day.Type_day,
			Exercises: make([][]ExNoWeight, len(day.Exercises)),
		}

		for j, ex := range day.Exercises {
			withWeight, noWeight, err := r.queryExerciseGroup(ctx, ex, profile)
			if err != nil {
				return EPlanWithWeight{}, EPlanNoWeight{}, fmt.Errorf(
					"query group=%q subgroup=%v: %w", ex.Group, ex.Sub_group, err,
				)
			}
			dayWithWeight.Exercises[j] = withWeight
			dayNoWeight.Exercises[j] = noWeight
		}

		planWithWeight.Plan[i] = dayWithWeight
		planNoWeight.Plan[i] = dayNoWeight
	}

	return planWithWeight, planNoWeight, nil
}

func (r *UsersRepository) queryExerciseGroup(
	ctx context.Context,
	ex mlclient.Muscules,
	profile Profile,
) ([]ExWithWeight, []ExNoWeight, error) {
	var (
		rows pgx.Rows
		err  error
	)

	if ex.Sub_group == nil {
		rows, err = r.pool.Query(ctx, `
			SELECT id, exercises_name, working_weights
			FROM sportapp.exercises
			WHERE muscular_group = $1 AND muscular_subgroup IS NULL
			ORDER BY id
			LIMIT 5
		`, ex.Group)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, exercises_name, working_weights
			FROM sportapp.exercises
			WHERE muscular_group = $1 AND muscular_subgroup = $2
			ORDER BY id
			LIMIT 5
		`, ex.Group, *ex.Sub_group)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("query exercises: %w", err)
	}
	defer rows.Close()

	var (
		withWeight []ExWithWeight
		noWeight   []ExNoWeight
	)

	for rows.Next() {
		var (
			id         int
			name       string
			baseWeight *int
		)
		if err := rows.Scan(&id, &name, &baseWeight); err != nil {
			return nil, nil, fmt.Errorf("scan exercise: %w", err)
		}

		var weight *float64
		if baseWeight != nil {
			w := computeUserWeight(float64(*baseWeight), profile)
			weight = &w
		}

		withWeight = append(withWeight, ExWithWeight{ID: id, EXName: name, Weight: weight})
		noWeight = append(noWeight, ExNoWeight{ID: id, EXName: name})
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("rows iteration: %w", err)
	}

	withWeight = normalizeWeightedBodyweightVariants(withWeight)

	return withWeight, noWeight, nil
}

func computeUserWeight(baseWeight float64, profile Profile) float64 {
	k := 1.0

	switch derefStr(profile.FitnessLevel) {
	case "Новичок":
		k *= 0.4
	case "Любитель":
		k *= 0.6
	default:
		k *= 0.8
	}

	switch derefStr(profile.ActivityLevel) {
	case "Высокая", "Очень высокая":
		k *= 1.1
	}

	if derefStr(profile.Gender) == "Женщина" {
		k *= 0.7
	}

	if profile.Age != nil && (*profile.Age < 20 || *profile.Age > 50) {
		k *= 0.8
	}

	if profile.InjuriesNotes != nil && *profile.InjuriesNotes {
		k *= 0.7
	}

	switch derefStr(profile.Goal) {
	case "Сжечь жир", "Сбросить вес":
		k *= 0.9
	}

	return roundWeightDownToGymStep(baseWeight * k)
}

const (
	gymPlateStepSmallKG     = 2.5
	gymPlateStepLargeKG     = 5.0
	gymPlateStepLargeFromKG = 50.0
)

func roundWeightDownToGymStep(kg float64) float64 {
	if kg <= 0 {
		return 0
	}
	step := gymPlateStepSmallKG
	if kg >= gymPlateStepLargeFromKG {
		step = gymPlateStepLargeKG
	}
	rounded := math.Floor(kg/step) * step
	return math.Round(rounded*10) / 10
}

func normalizeWeightedBodyweightVariants(list []ExWithWeight) []ExWithWeight {
	filtered := make([]ExWithWeight, 0, len(list))
	for _, ex := range list {
		// Если для версии "с отягощением" вес нулевой/пустой —
		// оставляем только базовую версию без отягощения.
		if isWeightedBodyweightExercise(ex.EXName) && !hasPositiveWeight(ex.Weight) {
			continue
		}
		filtered = append(filtered, ex)
	}
	return filtered
}

func hasPositiveWeight(w *float64) bool {
	return w != nil && *w > 0
}

func isWeightedBodyweightExercise(name string) bool {
	switch name {
	case "Отжимания на брусьях с отягощением", "Подтягивания с отягощением":
		return true
	default:
		return false
	}
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
