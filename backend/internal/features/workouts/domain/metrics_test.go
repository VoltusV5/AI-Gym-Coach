package workouts_domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	plans_postgres_repository "sport_app/internal/features/plans/repository/postgres"
)

func at(raw string) time.Time {
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		panic(err)
	}
	return t
}
func workout(weight float64, reps int, when string) Workout {
	return Workout{ID: uuid.NewString(), Input: Input{ClientSessionID: uuid.NewString(), PlanID: uuid.NewString(), PlanVersion: 1, PlanDayID: uuid.NewString(), StartedAt: at(when).Add(-time.Hour), FinishedAt: at(when), Slots: []Slot{{ExerciseID: 1, Position: 0, Sets: []Set{{Reps: 20, WeightKg: 100, IsWarmup: true}, {Reps: reps, WeightKg: weight}, {Reps: reps, WeightKg: weight}, {Reps: reps, WeightKg: weight}}}}}}
}
func TestMetricsControlDataset(t *testing.T) {
	first := workout(60, 10, "2026-10-26T12:00:00Z")
	second := workout(60, 11, "2026-10-27T12:00:00Z")
	require.Equal(t, 1800.0, Measure(first.Input).VolumeKg)
	require.Equal(t, 1980.0, Measure(second.Input).VolumeKg)
	require.Equal(t, 3, Measure(second.Input).WorkingSets)
	require.Empty(t, Records(first, nil))
	records := Records(second, []Workout{first})
	require.Len(t, records, 1)
	require.Equal(t, "volume_kg", records[0].Metric)
	require.Equal(t, 10.0, *records[0].DeltaPercent)
	points := ExerciseHistory([]Workout{second, first}, 1)
	require.Len(t, points.Points, 2)
	require.Equal(t, 60.0, points.Points[1].WorkingWeightKg)
	require.Equal(t, 11, points.Points[1].Reps)
	empty := ExerciseHistory([]Workout{first}, 2)
	require.Empty(t, empty.Points)
}
func TestProgressTimeZoneAndEmpty(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	start, end, err := Bounds("week", at("2026-11-01T12:00:00Z"), loc)
	require.NoError(t, err)
	require.Equal(t, 169*time.Hour, end.Sub(start))
	require.Equal(t, "2026-10-26T04:00:00Z", start.UTC().Format(time.RFC3339))
	p := Aggregate(Data{}, "week", start, end, end, loc)
	require.Zero(t, p.WorkoutCount)
	require.Nil(t, p.Adherence)
	require.Nil(t, p.WeightChangeKg)
	require.Empty(t, p.Highlights)
	require.Zero(t, p.StreakWeeks)
	outside := workout(60, 10, "2026-10-26T03:59:59Z")
	inside := workout(60, 11, "2026-10-26T04:00:00Z")
	data := Data{Workouts: []Workout{outside, inside}, Weights: []Weight{{Day: "2026-10-26", WeightKg: 80.5}, {Day: "2026-11-01", WeightKg: 79.25}}}
	p = Aggregate(data, "week", start, end, end, loc)
	require.Equal(t, 1, p.WorkoutCount)
	require.Equal(t, 1980.0, p.VolumeKg)
	require.Equal(t, -1.25, *p.WeightChangeKg)
}
func TestScheduleSnapshotAndClosedWeek(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Moscow")
	require.NoError(t, err)
	p := plans_postgres_repository.Plan{ID: uuid.NewString(), IsActive: true, Schedule: plans_postgres_repository.Schedule{TimeZone: loc.String(), StartDate: "2026-10-19", Entries: []plans_postgres_repository.ScheduleEntry{{Weekday: 1, PlanDayID: uuid.NewString()}}}}
	first := workout(60, 10, "2026-10-19T09:00:00Z")
	first.PlanID = p.ID
	first.PlanDayID = p.Schedule.Entries[0].PlanDayID
	data := Data{Workouts: []Workout{first}, Revisions: []Revision{{Plan: p, At: at("2026-10-18T00:00:00Z")}}}
	now := at("2026-10-26T12:00:00Z")
	start, end, _ := Bounds("week", now, loc)
	progress := Aggregate(data, "week", start, end, now, loc)
	require.Equal(t, 1, progress.StreakWeeks)
	require.Equal(t, 0.0, *progress.Adherence)
	p.Schedule.Entries = []plans_postgres_repository.ScheduleEntry{{Weekday: 3, PlanDayID: uuid.NewString()}}
	data.Revisions = append(data.Revisions, Revision{Plan: p, At: at("2026-10-26T12:01:00Z")})
	progress = Aggregate(data, "week", start, end, now, loc)
	require.Equal(t, 1, progress.StreakWeeks)
}
func TestWorkoutValidation(t *testing.T) {
	w := workout(60, 10, "2026-10-26T12:00:00Z")
	_, err := Validate(w.Input, at("2026-10-27T00:00:00Z"))
	require.NoError(t, err)
	w.Slots = append(w.Slots, w.Slots[0])
	_, err = Validate(w.Input, at("2026-10-27T00:00:00Z"))
	require.Error(t, err)
}
func TestEveningEditKeepsCompletedSchedule(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Moscow")
	require.NoError(t, err)
	id := uuid.NewString()
	dayA := uuid.NewString()
	dayB := uuid.NewString()
	plan := plans_postgres_repository.Plan{ID: id, IsActive: true, Schedule: plans_postgres_repository.Schedule{TimeZone: loc.String(), StartDate: "2026-10-26", Entries: []plans_postgres_repository.ScheduleEntry{{Weekday: 1, PlanDayID: dayA}}}}
	first := Revision{Plan: plan, At: at("2026-10-25T12:00:00Z")}
	plan.Schedule.Entries = []plans_postgres_repository.ScheduleEntry{{Weekday: 1, PlanDayID: dayB}}
	second := Revision{Plan: plan, At: at("2026-10-26T15:00:00Z")}
	w := workout(60, 10, "2026-10-26T09:00:00Z")
	w.PlanID = id
	w.PlanDayID = dayA
	now := at("2026-10-26T18:00:00Z")
	start, end, _ := Bounds("week", now, loc)
	p := Aggregate(Data{Workouts: []Workout{w}, Revisions: []Revision{first, second}}, "week", start, end, now, loc)
	require.NotNil(t, p.Adherence)
	require.Equal(t, 1.0, *p.Adherence)
	second.At = at("2026-10-26T21:00:00Z")
	p = Aggregate(Data{Workouts: []Workout{w}, Revisions: []Revision{first, second}}, "week", start, end, at("2026-10-27T12:00:00Z"), loc)
	require.Equal(t, 1.0, *p.Adherence)
}
func TestPlanTimeZoneUsesOccurrenceBoundary(t *testing.T) {
	plan := plans_postgres_repository.Plan{ID: uuid.NewString(), IsActive: true, Schedule: plans_postgres_repository.Schedule{TimeZone: "America/New_York", StartDate: "2026-10-01", Entries: []plans_postgres_repository.ScheduleEntry{{Weekday: 7, PlanDayID: uuid.NewString()}}}}
	data := Data{Revisions: []Revision{{Plan: plan, At: at("2026-10-01T00:00:00Z")}}}
	p, c := scheduleCounts(data, at("2026-10-26T00:00:00Z"), at("2026-10-27T00:00:00Z"), at("2026-10-28T00:00:00Z"), time.UTC)
	require.Zero(t, p)
	require.Zero(t, c)
}
func TestAllTimeIncludesMissedWeeksAndSummaryStreakIsHistorical(t *testing.T) {
	plan := plans_postgres_repository.Plan{ID: uuid.NewString(), IsActive: true, Schedule: plans_postgres_repository.Schedule{TimeZone: "UTC", StartDate: "2026-10-12", Entries: []plans_postgres_repository.ScheduleEntry{{Weekday: 1, PlanDayID: uuid.NewString()}}}}
	w := workout(60, 10, "2026-10-26T12:00:00Z")
	w.PlanID = plan.ID
	w.PlanDayID = plan.Schedule.Entries[0].PlanDayID
	data := Data{Revisions: []Revision{{Plan: plan, At: at("2026-10-11T00:00:00Z")}}, Workouts: []Workout{w}}
	p := Aggregate(data, "all", at("0001-01-01T00:00:00Z"), at("2026-10-27T00:00:00Z"), at("2026-10-27T00:00:00Z"), time.UTC)
	require.Equal(t, 1.0/3, *p.Adherence)
	start, end, _ := Bounds("week", at("2026-10-19T12:00:00Z"), time.UTC)
	p = Aggregate(data, "week", start, end, at("2026-11-10T00:00:00Z"), time.UTC)
	require.Zero(t, p.StreakWeeks)
}
