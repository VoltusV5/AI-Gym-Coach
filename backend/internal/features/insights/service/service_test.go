package insights_service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	catalog "sport_app/internal/features/catalog/repository/postgres"
	repository "sport_app/internal/features/insights/repository/postgres"
	"sport_app/internal/features/mlclient"
	plans "sport_app/internal/features/plans/repository/postgres"
	domain "sport_app/internal/features/workouts/domain"
	"testing"
	"time"
)

type testRepository struct {
	snapshot repository.Snapshot
	visible  map[string]string
}

func (r testRepository) Snapshot(context.Context, int) (repository.Snapshot, error) {
	return r.snapshot, nil
}
func (r testRepository) Visible(context.Context, []string) (map[string]string, error) {
	return r.visible, nil
}
func (r testRepository) Dashboard(context.Context, int) (repository.DashboardSnapshot, error) {
	return repository.DashboardSnapshot{Snapshot: r.snapshot}, nil
}
func (r testRepository) Analytics(context.Context, int, bool, Period) (map[string]any, error) {
	return nil, nil
}

type testRanker struct{ foreign, unavailable bool }

func (r testRanker) Rank(_ context.Context, in mlclient.RankInput) (mlclient.RankResult, error) {
	if r.unavailable {
		return mlclient.RankResult{}, errors.New("dependency failed")
	}
	id := "p"
	if r.foreign {
		id = "foreign"
	}
	return mlclient.RankResult{RequestID: in.RequestID, ModelVersion: "test", Items: []mlclient.RankItem{{ProgramID: id, VersionID: "v", Score: 1, Evidence: mlclient.Evidence{Level: "beginner", SessionsPerWeek: 3, EquipmentMatch: true}, ReasonCodes: []string{"EQUIPMENT_MATCH"}}}}, nil
}
func (r testRanker) Explain(_ context.Context, id string, _ []mlclient.RankItem) (mlclient.ExplainResult, error) {
	return mlclient.ExplainResult{RequestID: id, Items: []mlclient.Explanation{{ProgramID: "p", Text: "Invented medical claims", SourceVersions: []string{"v"}}}}, nil
}
func TestRecommendationFallbackAndPublicationRecheck(t *testing.T) {
	snapshot := repository.Snapshot{Cards: []catalog.ProgramCard{{ID: "p", PublishedVersionID: "v", Level: "beginner", SessionsPerWeek: 3}}}
	for _, tc := range []struct {
		name     string
		ranker   testRanker
		visible  map[string]string
		fallback bool
		count    int
	}{{"dependency fallback", testRanker{unavailable: true}, map[string]string{"p": "v"}, true, 1}, {"reject foreign", testRanker{foreign: true}, map[string]string{"p": "v"}, true, 1}, {"valid ranking factual explanation", testRanker{}, map[string]string{"p": "v"}, false, 1}, {"withdrawn during ML", testRanker{}, map[string]string{}, false, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := NewService(testRepository{snapshot, tc.visible}, tc.ranker, 14).Recommendations(context.Background(), "101", 10)
			require.NoError(t, err)
			require.Equal(t, tc.fallback, out.Fallback)
			require.Len(t, out.Items, tc.count)
			if tc.count > 0 {
				require.NotContains(t, out.Items[0].Explanation, "Invented")
				require.Contains(t, out.Items[0].Explanation, "3")
			}
		})
	}
}
func TestDashboardStatePriorityAndTimeZone(t *testing.T) {
	now := time.Date(2026, 11, 13, 22, 0, 0, 0, time.UTC)
	plan := plans.Plan{ID: "plan", Schedule: plans.Schedule{TimeZone: "Europe/Moscow", StartDate: "2026-01-01", Entries: []plans.ScheduleEntry{{Weekday: 6, PlanDayID: "day"}}}}
	for _, tc := range []struct {
		name, state string
		plan        *plans.Plan
		finished    time.Time
	}{{"new", "new_user", nil, time.Time{}}, {"scheduled in Moscow", "workout_today", &plan, time.Time{}}, {"completed", "just_completed", &plan, now.Add(-time.Hour)}, {"inactive", "inactive", &plan, now.AddDate(0, 0, -20)}} {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := repository.DashboardSnapshot{ActivePlan: tc.plan}
			if !tc.finished.IsZero() {
				snapshot.Data.Workouts = []domain.Workout{{ID: "session", Input: domain.Input{StartedAt: tc.finished.Add(-time.Hour), FinishedAt: tc.finished}}}
			}
			out := buildDashboard(snapshot, now, 14)
			require.Equal(t, tc.state, out.State)
			require.NotEmpty(t, out.Blocks)
			if tc.state == "just_completed" {
				require.Equal(t, 120, out.Blocks[0].Priority)
			}
		})
	}
}
func TestAnalyticsPeriodBoundaries(t *testing.T) {
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	p, err := ParsePeriod("2026-11-01", "2026-11-14", now)
	require.NoError(t, err)
	require.Equal(t, time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC), p.To)
	for _, v := range [][2]string{{"bad", ""}, {"2026-11-14", "2026-11-01"}, {"2024-01-01", "2026-11-14"}, {"2026-11-01", "2027-01-01"}} {
		_, err := ParsePeriod(v[0], v[1], now)
		require.Error(t, err)
	}
}
