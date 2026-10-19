package users_service_test

import (
	"context"
	"testing"
	"time"

	users_postgres_repository "sport_app/internal/features/users/repository/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProgressionCalculation(t *testing.T) {
	t.Run("Exceeding max reps (>=12) triggers weight increase", func(t *testing.T) {
		currentWeight := 50.0
		// Рекомендация 50 кг, фактический вес 50 кг, 12 повторений.
		next := users_postgres_repository.NextRecommendedWeight(currentWeight, currentWeight, 12)
		assert.Greater(t, next, currentWeight, "expected weight to increase for 12 reps")
	})

	t.Run("Failing minimum reps (<6) triggers weight reduction", func(t *testing.T) {
		currentWeight := 80.0
		// Рекомендация 80 кг, фактический вес 80 кг, 4 повторения.
		next := users_postgres_repository.NextRecommendedWeight(currentWeight, currentWeight, 4)
		assert.Less(t, next, currentWeight, "expected weight to decrease for 4 reps")
	})

	t.Run("Standard reps (8 reps) maintains weight", func(t *testing.T) {
		currentWeight := 60.0
		// Рекомендация 60 кг, фактический вес 60 кг, 8 повторений.
		next := users_postgres_repository.NextRecommendedWeight(currentWeight, currentWeight, 8)
		assert.Equal(t, currentWeight, next, "expected weight to remain steady for 8 reps")
	})
}

func TestUsersService_CompleteWorkout_Flow(t *testing.T) {
	var savedWorkout users_postgres_repository.WorkoutCompleteRequest
	var completeCalled bool

	repo := &mockUsersRepository{
		LoadCompletedWorkoutsFunc: func(ctx context.Context, userID string) ([]users_postgres_repository.WorkoutCompleteRequest, error) {
			// Предыдущая сессия: 50 кг в упражнении 1.
			return []users_postgres_repository.WorkoutCompleteRequest{
				{
					DayCode:    "day_1",
					FinishedAt: time.Now().Add(-48 * time.Hour),
					Slots: []users_postgres_repository.SlotData{
						{
							SlotIndex:  0,
							ExerciseID: 1,
							Sets: []users_postgres_repository.SetData{
								{WeightKg: 50.0, Reps: 10},
							},
						},
					},
				},
			}, nil
		},
		GetUserDataVersionFunc: func(ctx context.Context, userID string) (int64, error) {
			return 1, nil
		},
		GetWorkingWeightsFunc: func(ctx context.Context, userID string) (map[string]float64, error) {
			return map[string]float64{"1": 50.0}, nil
		},
		GetExerciseNamesByIDsFunc: func(ctx context.Context, ids []int) (map[int]string, error) {
			return map[int]string{1: "Bench Press"}, nil
		},
		CompleteWorkoutFunc: func(ctx context.Context, userID string, req users_postgres_repository.WorkoutCompleteRequest, newWorkingWeights []byte, expectedVersion int64) error {
			completeCalled = true
			savedWorkout = req
			return nil
		},
		UnlockNewAchievementsFunc: func(ctx context.Context, userID string) ([]users_postgres_repository.AchievementUnlocked, error) {
			return []users_postgres_repository.AchievementUnlocked{
				{ID: 1, Title: "First 60kg Bench"},
			}, nil
		},
	}

	svc, _ := newTestUsersService(repo)

	req := users_postgres_repository.WorkoutCompleteRequest{
		DayCode:    "day_1",
		FinishedAt: time.Now(),
		Slots: []users_postgres_repository.SlotData{
			{
				SlotIndex:  0,
				ExerciseID: 1,
				Sets: []users_postgres_repository.SetData{
					{WeightKg: 60.0, Reps: 12}, // Личный рекорд, 12 повторений.
				},
			},
		},
	}

	result, err := svc.CompleteWorkout(context.Background(), "user-100", req)
	require.NoError(t, err)
	assert.True(t, completeCalled)
	assert.Equal(t, "day_1", savedWorkout.DayCode)

	// Проверяем разблокировку достижений.
	require.Len(t, result.NewAchievements, 1)
	assert.Equal(t, "First 60kg Bench", result.NewAchievements[0].Title)

	// Проверяем личный рекорд в итогах сессии.
	require.NotEmpty(t, result.SessionHighlights)
	assert.Equal(t, 1, result.SessionHighlights[0].ExerciseID)
	assert.Equal(t, "Bench Press", result.SessionHighlights[0].ExerciseName)

	// Проверяем обновление плана с учётом прогрессии весов.
	require.NotEmpty(t, result.PlanUpdates)
	assert.Equal(t, 1, result.PlanUpdates[0].ExerciseID)
	assert.Equal(t, 50.0, result.PlanUpdates[0].OldWeight)
	assert.Greater(t, result.PlanUpdates[0].NewWeight, 50.0)
}
