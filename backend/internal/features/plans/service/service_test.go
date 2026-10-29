package plans_service

import (
	"context"
	"math"
	"testing"
	"time"

	core_errors "sport_app/internal/core/errors"
	plans_postgres_repository "sport_app/internal/features/plans/repository/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validInput() plans_postgres_repository.Input {
	return plans_postgres_repository.Input{
		Title: "План", Version: 1,
		Days:     []plans_postgres_repository.Day{{ID: "28000000-0000-4000-8000-000000000001", Title: "День", Exercises: []plans_postgres_repository.Prescription{{ExerciseID: 1, Sets: 3, RepsMin: 8, RepsMax: 12, RestSec: 90}}}},
		Schedule: plans_postgres_repository.Schedule{TimeZone: "Europe/Moscow", StartDate: "2026-10-29", Entries: []plans_postgres_repository.ScheduleEntry{{Weekday: 1, PlanDayID: "28000000-0000-4000-8000-000000000001"}}},
	}
}

func TestPlanValidation(t *testing.T) {
	for _, tt := range []struct {
		name   string
		change func(*plans_postgres_repository.Input)
	}{
		{"empty title", func(p *plans_postgres_repository.Input) { p.Title = " " }},
		{"null days", func(p *plans_postgres_repository.Input) { p.Days = nil }},
		{"bad version", func(p *plans_postgres_repository.Input) { p.Version = 0 }},
		{"overflow version", func(p *plans_postgres_repository.Input) { p.Version = math.MaxInt64 }},
		{"bad timezone", func(p *plans_postgres_repository.Input) { p.Schedule.TimeZone = "invalid" }},
		{"local timezone", func(p *plans_postgres_repository.Input) { p.Schedule.TimeZone = "Local" }},
		{"bad date", func(p *plans_postgres_repository.Input) { p.Schedule.StartDate = "2026-02-30" }},
		{"foreign day", func(p *plans_postgres_repository.Input) {
			p.Schedule.Entries[0].PlanDayID = "28000000-0000-4000-8000-000000000002"
		}},
		{"repeated weekday", func(p *plans_postgres_repository.Input) {
			p.Schedule.Entries = append(p.Schedule.Entries, p.Schedule.Entries[0])
		}},
		{"bad weekday", func(p *plans_postgres_repository.Input) { p.Schedule.Entries[0].Weekday = 8 }},
		{"duplicate day", func(p *plans_postgres_repository.Input) { p.Days = append(p.Days, p.Days[0]) }},
		{"bad day id", func(p *plans_postgres_repository.Input) { p.Days[0].ID = "bad" }},
		{"negative position", func(p *plans_postgres_repository.Input) { p.Days[0].Position = -1 }},
		{"unknown integer range", func(p *plans_postgres_repository.Input) { p.Days[0].Exercises[0].ExerciseID = 1 << 32 }},
		{"zero sets", func(p *plans_postgres_repository.Input) { p.Days[0].Exercises[0].Sets = 0 }},
		{"reversed reps", func(p *plans_postgres_repository.Input) { p.Days[0].Exercises[0].RepsMax = 2 }},
		{"negative rest", func(p *plans_postgres_repository.Input) { p.Days[0].Exercises[0].RestSec = -1 }},
		{"nonfinite weight", func(p *plans_postgres_repository.Input) { w := math.Inf(1); p.Days[0].Exercises[0].WeightKg = &w }},
		{"duplicate exercise position", func(p *plans_postgres_repository.Input) {
			p.Days[0].Exercises = append(p.Days[0].Exercises, p.Days[0].Exercises[0])
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			input := validInput()
			tt.change(&input)
			_, err := prepareInput(input, true)
			assert.ErrorIs(t, err, core_errors.ErrInvalidArgument)
		})
	}
	input := validInput()
	input.Days[0].ID = ""
	input.Schedule.Entries = []plans_postgres_repository.ScheduleEntry{}
	ready, err := prepareInput(input, false)
	require.NoError(t, err)
	assert.NotEmpty(t, ready.Days[0].ID)
	assert.Empty(t, input.Days[0].ID)
	_, err = prepareInput(plans_postgres_repository.Input{Title: "Новый", Days: []plans_postgres_repository.Day{}, Schedule: plans_postgres_repository.Schedule{TimeZone: "UTC", StartDate: "2026-10-29", Entries: []plans_postgres_repository.ScheduleEntry{}}}, false)
	assert.NoError(t, err)
}

type testRepository struct {
	Repository
	filter plans_postgres_repository.Filter
	items  []plans_postgres_repository.Plan
}

func (r *testRepository) ListPlans(_ context.Context, _ int, filter plans_postgres_repository.Filter) ([]plans_postgres_repository.Plan, error) {
	r.filter = filter
	return r.items, nil
}

func TestPlanCursorAndIdentity(t *testing.T) {
	repo := &testRepository{items: []plans_postgres_repository.Plan{{ID: "18000000-0000-4000-8000-000000000001", CreatedAt: time.Now()}, {ID: "18000000-0000-4000-8000-000000000002", CreatedAt: time.Now()}}}
	svc := NewService(repo)
	first, err := svc.ListPlans(context.Background(), "101", 1, "")
	require.NoError(t, err)
	require.NotNil(t, first.NextCursor)
	repo.items = repo.items[1:]
	last, err := svc.ListPlans(context.Background(), "101", 1, *first.NextCursor)
	require.NoError(t, err)
	assert.Nil(t, last.NextCursor)
	assert.Equal(t, first.Items[0].ID, repo.filter.AfterID)
	_, err = svc.ListPlans(context.Background(), "102", 1, *first.NextCursor)
	assert.ErrorIs(t, err, core_errors.ErrInvalidArgument)
	for _, raw := range []string{"", "abc", "0", "4294967296"} {
		_, err := svc.ListPlans(context.Background(), raw, 20, "")
		assert.ErrorIs(t, err, core_errors.ErrUnauthorized)
	}
	_, err = svc.ListPlans(context.Background(), "101", 101, "")
	assert.ErrorIs(t, err, core_errors.ErrInvalidArgument)
	_, err = svc.ListPlans(context.Background(), "101", 20, "invalid")
	assert.ErrorIs(t, err, core_errors.ErrInvalidArgument)
}
