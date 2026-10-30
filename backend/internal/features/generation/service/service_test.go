package generation_service

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	core_errors "sport_app/internal/core/errors"
	generation_postgres_repository "sport_app/internal/features/generation/repository/postgres"
	"sport_app/internal/features/mlclient"
	plans_postgres_repository "sport_app/internal/features/plans/repository/postgres"
	profile_postgres_repository "sport_app/internal/features/profile/repository/postgres"
	workouts_domain "sport_app/internal/features/workouts/domain"
)

type repoStub struct {
	profile    profile_postgres_repository.Profile
	history    workouts_domain.Data
	saved      int
	cached     bool
	input      plans_postgres_repository.Input
	exercises  []mlclient.ExerciseCandidate
	catalogErr error
}

func (r *repoStub) Cached(context.Context, int, string, string) (generation_postgres_repository.Result, bool, error) {
	return generation_postgres_repository.Result{PlanID: "cached"}, r.cached, nil
}
func (r *repoStub) Profile(context.Context, int) (profile_postgres_repository.Profile, error) {
	return r.profile, nil
}
func (r *repoStub) History(context.Context, int) (workouts_domain.Data, error) { return r.history, nil }
func (r *repoStub) Exercises(context.Context) ([]mlclient.ExerciseCandidate, error) {
	if r.exercises != nil || r.catalogErr != nil {
		return r.exercises, r.catalogErr
	}
	return []mlclient.ExerciseCandidate{{ExerciseID: 1, MuscleGroup: "Ноги", EquipmentIDs: []string{}}}, nil
}
func (r *repoStub) Save(_ context.Context, _ int, _, _ string, _ int64, input plans_postgres_repository.Input, _ mlclient.GenerateResult) (generation_postgres_repository.Result, error) {
	r.saved++
	r.input = input
	return generation_postgres_repository.Result{PlanID: "saved"}, nil
}

type clientStub struct {
	calls  int
	input  mlclient.GenerateInput
	result mlclient.GenerateResult
	err    error
}

func (c *clientStub) Generate(_ context.Context, in mlclient.GenerateInput) (mlclient.GenerateResult, error) {
	c.calls++
	c.input = in
	return c.result, c.err
}
func completeProfile() profile_postgres_repository.Profile {
	age, height, duration := 26, 175, 60
	weight := 75.5
	level, goal := "beginner", "muscle_gain"
	return profile_postgres_repository.Profile{Version: 3, Age: &age, HeightCm: &height, WeightKg: &weight, Level: &level, Goal: &goal, SessionDurationMin: &duration, TrainingDays: []int{1}, EquipmentIDs: []string{}, ExcludedExerciseIDs: []int{}}
}
func TestGenerationPreservesSourceAndRejectsInvalidResults(t *testing.T) {
	repo := &repoStub{profile: completeProfile()}
	client := &clientStub{result: mlclient.GenerateResult{RequestID: "req", ModelVersion: "model-test", Split: "Full body", Days: []mlclient.GeneratedDay{{Weekday: 1, Title: "Day A", Exercises: []plans_postgres_repository.Prescription{{ExerciseID: 1, Position: 0, Sets: 3, RepsMin: 8, RepsMax: 12, RestSec: 90}}}}}}
	service := NewService(repo, client)
	_, err := service.Generate(context.Background(), "1", "key", "req", 3)
	require.NoError(t, err)
	require.Equal(t, 1, repo.saved)
	require.Len(t, repo.input.Days, 1)
	require.Equal(t, "profile-v2", client.input.FeatureSchemaVersion)
	require.Equal(t, 1, client.input.Exercises[0].ExerciseID)
	client.result.Days[0].Weekday = 2
	_, err = service.Generate(context.Background(), "1", "another", "req", 3)
	var domain *core_errors.DomainError
	require.ErrorAs(t, err, &domain)
	require.Equal(t, "NO_FEASIBLE_PLAN", domain.Code)
	require.Equal(t, 1, repo.saved)
	client.err = context.DeadlineExceeded
	_, err = service.Generate(context.Background(), "1", "deadline", "req", 3)
	require.True(t, errors.Is(err, context.DeadlineExceeded))
	require.Equal(t, 1, repo.saved)
}

func TestGenerationCatalogFailureDoesNotCallMLOrSave(t *testing.T) {
	for _, catalogErr := range []error{nil, errors.New("catalog unavailable")} {
		repo := &repoStub{profile: completeProfile(), exercises: []mlclient.ExerciseCandidate{}, catalogErr: catalogErr}
		client := &clientStub{}
		_, err := NewService(repo, client).Generate(context.Background(), "1", "key", "req", 3)
		require.Error(t, err)
		if catalogErr != nil {
			require.ErrorIs(t, err, catalogErr)
		}
		require.Zero(t, client.calls)
		require.Zero(t, repo.saved)
	}
}

// TestGenerationPythonContract runs against an explicitly started disposable ML service.
// It exercises real CatBoost/Pydantic and Go validation without a database or user data.
func TestGenerationPythonContract(t *testing.T) {
	url := os.Getenv("SPORTIK_ML_TEST_URL")
	if url == "" {
		t.Skip("set SPORTIK_ML_TEST_URL to a disposable local ML service")
	}
	profile := completeProfile()
	profile.TrainingDays = []int{1, 3, 5}
	gender, activity, injuries := "м", "Средняя активность", false
	profile.Gender, profile.ActivityLevel, profile.InjuriesNotes = &gender, &activity, &injuries
	candidates := []mlclient.ExerciseCandidate{}
	for i, group := range []string{"Ноги", "Грудь", "Спина", "Плечи", "Руки", "Пресс"} {
		candidates = append(candidates, mlclient.ExerciseCandidate{ExerciseID: 1001 + i, MuscleGroup: group, EquipmentIDs: []string{}})
	}
	repo := &repoStub{profile: profile, exercises: candidates}
	client := mlclient.NewV2Client(mlclient.Config{BaseURL: url, ServiceToken: "synthetic-integration-token"})
	service := NewService(repo, client)
	_, err := service.Generate(context.Background(), "1", "integration", "req-python-contract", 3)
	require.NoError(t, err)
	require.Equal(t, 1, repo.saved)
	require.Len(t, repo.input.Days, 3)
	for _, day := range repo.input.Days {
		require.NotEmpty(t, day.Exercises)
		for _, ex := range day.Exercises {
			require.GreaterOrEqual(t, ex.ExerciseID, 1001)
			require.LessOrEqual(t, ex.ExerciseID, 1006)
			require.Nil(t, ex.WeightKg)
		}
	}
	// Exclude the entire catalog: failure must leave the previously saved plan alone.
	for _, ex := range candidates {
		repo.profile.ExcludedExerciseIDs = append(repo.profile.ExcludedExerciseIDs, ex.ExerciseID)
	}
	_, err = service.Generate(context.Background(), "1", "infeasible", "req-python-infeasible", 3)
	var domain *core_errors.DomainError
	require.ErrorAs(t, err, &domain)
	require.Equal(t, "NO_FEASIBLE_PLAN", domain.Code)
	require.Equal(t, 1, repo.saved)
}
func TestGenerationIncompleteStaleAndCached(t *testing.T) {
	repo := &repoStub{profile: completeProfile()}
	client := &clientStub{}
	service := NewService(repo, client)
	_, err := service.Generate(context.Background(), "1", "key", "req", 2)
	require.ErrorIs(t, err, core_errors.ErrConflict)
	require.Zero(t, client.calls)
	repo.profile.WeightKg = nil
	_, err = service.Generate(context.Background(), "1", "key", "req", 3)
	var domain *core_errors.DomainError
	require.ErrorAs(t, err, &domain)
	require.Equal(t, "PROFILE_INCOMPLETE", domain.Code)
	require.Zero(t, client.calls)
	repo.cached = true
	result, err := service.Generate(context.Background(), "1", "key", "req", 3)
	require.NoError(t, err)
	require.Equal(t, "cached", result.PlanID)
	require.Zero(t, client.calls)
}
func TestGenerationHistoryUsesLast28Days(t *testing.T) {
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	repo := &repoStub{profile: completeProfile(), history: workouts_domain.Data{Workouts: []workouts_domain.Workout{{Input: workouts_domain.Input{FinishedAt: now.AddDate(0, 0, -29)}}, {Input: workouts_domain.Input{FinishedAt: now.AddDate(0, 0, -1)}}}}}
	client := &clientStub{err: context.Canceled}
	service := NewService(repo, client)
	service.now = func() time.Time { return now }
	_, err := service.Generate(context.Background(), "1", "key", "req", 3)
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 1, client.input.History.Completed28d)
}
