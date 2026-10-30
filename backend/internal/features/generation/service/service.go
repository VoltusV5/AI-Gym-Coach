package generation_service

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	core_errors "sport_app/internal/core/errors"
	"sport_app/internal/core/transport/http/v2request"
	generation_postgres_repository "sport_app/internal/features/generation/repository/postgres"
	"sport_app/internal/features/mlclient"
	plans_postgres_repository "sport_app/internal/features/plans/repository/postgres"
	plans_service "sport_app/internal/features/plans/service"
	profile_postgres_repository "sport_app/internal/features/profile/repository/postgres"
	workouts_domain "sport_app/internal/features/workouts/domain"
	workouts_service "sport_app/internal/features/workouts/service"
)

type Repository interface {
	Cached(context.Context, int, string, string) (generation_postgres_repository.Result, bool, error)
	Profile(context.Context, int) (profile_postgres_repository.Profile, error)
	History(context.Context, int) (workouts_domain.Data, error)
	Save(context.Context, int, string, string, int64, plans_postgres_repository.Input, mlclient.GenerateResult) (generation_postgres_repository.Result, error)
}
type Client interface {
	Generate(context.Context, mlclient.GenerateInput) (mlclient.GenerateResult, error)
}
type Service struct {
	repo   Repository
	client Client
	now    func() time.Time
}

func NewService(repo Repository, client Client) *Service {
	return &Service{repo: repo, client: client, now: time.Now}
}
func (s *Service) Generate(ctx context.Context, raw, key, requestID string, version int64) (generation_postgres_repository.Result, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return generation_postgres_repository.Result{}, err
	}
	if version < 1 {
		return generation_postgres_repository.Result{}, v2request.Invalid()
	}
	if err := workouts_service.Key(key); err != nil {
		return generation_postgres_repository.Result{}, err
	}
	hash := workouts_service.Hash(struct {
		ProfileVersion int64 `json:"profile_version"`
	}{version})
	result, hit, err := s.repo.Cached(ctx, id, key, hash)
	if err != nil || hit {
		return result, err
	}
	profile, err := s.repo.Profile(ctx, id)
	if err != nil {
		return result, err
	}
	if profile.Version != version {
		return result, fmt.Errorf("profile version changed: %w", core_errors.ErrConflict)
	}
	if profile.Age == nil || profile.HeightCm == nil || profile.WeightKg == nil || profile.Level == nil || profile.Goal == nil || len(profile.TrainingDays) == 0 || profile.EquipmentIDs == nil || profile.SessionDurationMin == nil {
		return result, &core_errors.DomainError{Code: "PROFILE_INCOMPLETE", Message: "Complete your profile before generating a plan", Status: 422}
	}
	history, err := s.repo.History(ctx, id)
	if err != nil {
		return result, err
	}
	now := s.now()
	start := now.AddDate(0, 0, -28)
	aggregate := workouts_domain.Aggregate(history, "window", start, now, now, time.UTC)
	mlHistory := mlclient.History{Adherence28d: aggregate.Adherence, ExerciseMetrics: []mlclient.HistoryMetric{}}
	for _, w := range history.Workouts {
		if !w.FinishedAt.Before(start) && w.FinishedAt.Before(now) {
			mlHistory.Completed28d++
		}
	}
	ids := map[int]bool{}
	for _, w := range history.Workouts {
		for _, slot := range w.Slots {
			ids[slot.ExerciseID] = true
		}
	}
	ordered := []int{}
	for exercise := range ids {
		ordered = append(ordered, exercise)
	}
	slices.Sort(ordered)
	for _, exercise := range ordered {
		series := workouts_domain.ExerciseHistory(append([]workouts_domain.Workout(nil), history.Workouts...), exercise)
		if len(series.Points) == 0 {
			continue
		}
		p := series.Points[len(series.Points)-1]
		mlHistory.ExerciseMetrics = append(mlHistory.ExerciseMetrics, mlclient.HistoryMetric{ExerciseID: exercise, WeightKg: p.WorkingWeightKg, Reps: p.Reps, VolumeKg: p.VolumeKg, MeasuredAt: p.FinishedAt})
	}
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	ml, err := s.client.Generate(ctx, mlclient.GenerateInput{RequestID: requestID, FeatureSchemaVersion: "profile-v2", Profile: profile, History: &mlHistory})
	if err != nil {
		return result, err
	}
	if len(ml.Days) != len(profile.TrainingDays) || ml.RequestID != requestID || !v2request.Text(ml.ModelVersion, 200, true) {
		return result, mlclient.NoFeasiblePlan()
	}
	input := plans_postgres_repository.Input{Title: ml.Split, Days: []plans_postgres_repository.Day{}, Schedule: plans_postgres_repository.Schedule{TimeZone: "Europe/Moscow", StartDate: now.In(time.FixedZone("MSK", 3*3600)).Format(time.DateOnly), Entries: []plans_postgres_repository.ScheduleEntry{}}}
	weekdays := map[int]bool{}
	for i, day := range ml.Days {
		if !slices.Contains(profile.TrainingDays, day.Weekday) || weekdays[day.Weekday] || len(day.Exercises) == 0 {
			return result, mlclient.NoFeasiblePlan()
		}
		weekdays[day.Weekday] = true
		did := uuid.NewString()
		input.Days = append(input.Days, plans_postgres_repository.Day{ID: did, Title: day.Title, Position: i, Exercises: day.Exercises})
		input.Schedule.Entries = append(input.Schedule.Entries, plans_postgres_repository.ScheduleEntry{Weekday: day.Weekday, PlanDayID: did})
	}
	input, err = plans_service.PrepareInput(input, false)
	if err != nil {
		return result, mlclient.NoFeasiblePlan()
	}
	if ml.Warnings == nil {
		ml.Warnings = []string{}
	}
	return s.repo.Save(ctx, id, key, hash, version, input, ml)
}
