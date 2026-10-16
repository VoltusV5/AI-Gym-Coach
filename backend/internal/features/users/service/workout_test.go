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
		// recommended=50, actual=50, reps=12
		next := users_postgres_repository.NextRecommendedWeight(currentWeight, currentWeight, 12)
		assert.Greater(t, next, currentWeight, "expected weight to increase for 12 reps")
	})

	t.Run("Failing minimum reps (<6) triggers weight reduction", func(t *testing.T) {
		currentWeight := 80.0
		// recommended=80, actual=80, reps=4
		next := users_postgres_repository.NextRecommendedWeight(currentWeight, currentWeight, 4)
		assert.Less(t, next, currentWeight, "expected weight to decrease for 4 reps")
	})

	t.Run("Standard reps (8 reps) maintains weight", func(t *testing.T) {
		currentWeight := 60.0
		// recommended=60, actual=60, reps=8
		next := users_postgres_repository.NextRecommendedWeight(currentWeight, currentWeight, 8)
		assert.Equal(t, currentWeight, next, "expected weight to remain steady for 8 reps")
	})
}

func TestUsersService_CompleteWorkout_Flow(t *testing.T) {
	var savedWorkout users_postgres_repository.WorkoutCompleteRequest
	var completeCalled bool

	repo := &mockUsersRepository{
		LoadCompletedWorkoutsFunc: func(ctx context.Context, userID string) ([]users_postgres_repository.WorkoutCompleteRequest, error) {
			// Previous session with 50kg on exercise 1
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
					{WeightKg: 60.0, Reps: 12}, // Personal Record + 12 reps
				},
			},
		},
	}

	result, err := svc.CompleteWorkout(context.Background(), "user-100", req)
	require.NoError(t, err)
	assert.True(t, completeCalled)
	assert.Equal(t, "day_1", savedWorkout.DayCode)

	// Verify achievements unlocked
	require.Len(t, result.NewAchievements, 1)
	assert.Equal(t, "First 60kg Bench", result.NewAchievements[0].Title)

	// Verify session highlights contains personal record
	require.NotEmpty(t, result.SessionHighlights)
	assert.Equal(t, 1, result.SessionHighlights[0].ExerciseID)
	assert.Equal(t, "Bench Press", result.SessionHighlights[0].ExerciseName)

	// Verify plan updates reflect weight progression
	require.NotEmpty(t, result.PlanUpdates)
	assert.Equal(t, 1, result.PlanUpdates[0].ExerciseID)
	assert.Equal(t, 50.0, result.PlanUpdates[0].OldWeight)
	assert.Greater(t, result.PlanUpdates[0].NewWeight, 50.0)
}
