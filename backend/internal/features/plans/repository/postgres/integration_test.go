package plans_postgres_repository_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	core_auth "sport_app/internal/core/auth"
	core_errors "sport_app/internal/core/errors"
	core_pgx_pool "sport_app/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_server "sport_app/internal/core/transport/http/server"
	plans_postgres_repository "sport_app/internal/features/plans/repository/postgres"
	plans_service "sport_app/internal/features/plans/service"
	plans_transport_http "sport_app/internal/features/plans/transport/http"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleInput(title string) plans_postgres_repository.Input {
	id := uuid.NewString()
	weight := 12.5
	return plans_postgres_repository.Input{
		Title: title,
		Days: []plans_postgres_repository.Day{{ID: id, Title: "День A", Position: 0, Exercises: []plans_postgres_repository.Prescription{
			{ExerciseID: 1, Position: 0, Sets: 3, RepsMin: 8, RepsMax: 12, WeightKg: &weight, RestSec: 90},
		}}},
		Schedule: plans_postgres_repository.Schedule{TimeZone: "Europe/Moscow", StartDate: "2026-10-29", Entries: []plans_postgres_repository.ScheduleEntry{{Weekday: 1, PlanDayID: id}}},
	}
}

func inputOf(plan plans_postgres_repository.Plan) plans_postgres_repository.Input {
	data, _ := json.Marshal(plans_postgres_repository.Input{Version: plan.Version, Title: plan.Title, Days: plan.Days, Schedule: plan.Schedule})
	var input plans_postgres_repository.Input
	_ = json.Unmarshal(data, &input)
	return input
}

func TestPlansIntegration(t *testing.T) {
	database := os.Getenv("SPORTIK_PLANS_TEST_DB")
	if database == "" {
		t.Skip("set SPORTIK_PLANS_TEST_DB to a new sportik_plans_test_* database on 127.0.0.1:55443")
	}
	require.True(t, strings.HasPrefix(database, "sportik_plans_test_"))
	ctx := context.Background()
	config := core_pgx_pool.Config{Host: "127.0.0.1", Port: "55443", User: "sportik_test", Database: database, Timeout: 10 * time.Second}
	pool, err := core_pgx_pool.NewPool(ctx, config)
	require.NoError(t, err)
	defer pool.Close()
	var empty bool
	require.NoError(t, pool.QueryRow(ctx, `SELECT to_regnamespace('sportapp') IS NULL`).Scan(&empty))
	require.True(t, empty, "existing schema is never cleared")
	apply := func(name string) {
		t.Helper()
		sql, err := os.ReadFile(filepath.Join("../../../../../migrations", name))
		require.NoError(t, err)
		_, err = pool.Exec(ctx, string(sql))
		require.NoError(t, err, name)
	}
	files, err := filepath.Glob("../../../../../migrations/00000[1-8]_*.up.sql")
	require.NoError(t, err)
	for _, file := range files {
		apply(filepath.Base(file))
	}
	apply("000009_personal_plans.up.sql")
	apply("000009_personal_plans.down.sql")
	_, err = pool.Exec(ctx, `INSERT INTO sportapp.users (id,is_anonymous,created_at) VALUES
		(101,FALSE,NOW()),(102,FALSE,NOW()),(103,TRUE,NOW());
		INSERT INTO sportapp.profile (user_id,weight_kg,created_at) VALUES (101,75.5,NOW()),(102,80,NOW());
		INSERT INTO sportapp.notes (user_id,title,body) VALUES (101,'fixture','keep');
		INSERT INTO sportapp.user_programs (user_id,is_active,plan_template,plan_exercises)
		VALUES (101,TRUE,'{"legacy":true}','{"plan":[]}');`)
	require.NoError(t, err)
	require.NoError(t, users_postgres_repository.NewUsersRepository(pool).EnsureExercisesSeeded(ctx))
	apply("000009_personal_plans.up.sql")
	repo := plans_postgres_repository.NewRepository(pool)
	svc := plans_service.NewService(repo)
	jwt := core_auth.NewJWT(core_auth.Config{Secret: "plans-test-only", TokenTTL: time.Hour})
	router := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion2)
	router.RegisterRoutes(plans_transport_http.NewPlansHTTPHandler(svc, jwt).Routes()...)
	handler := core_http_middleware.RequestID()(http.StripPrefix("/api/v2", router))
	token := func(user string) string {
		value, err := jwt.CreateToken(user, user == "103")
		require.NoError(t, err)
		return value
	}
	request := func(t *testing.T, method, path, user string, body any, status int) *httptest.ResponseRecorder {
		t.Helper()
		var data []byte
		if body != nil {
			data, err = json.Marshal(body)
			require.NoError(t, err)
		}
		r := httptest.NewRequest(method, "/api/v2"+path, bytes.NewReader(data))
		if user != "" {
			r.Header.Set("Authorization", "Bearer "+token(user))
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		require.Equal(t, status, w.Code, w.Body.String())
		if status >= 400 {
			assert.Contains(t, w.Body.String(), `"request_id":`)
			assert.NotContains(t, w.Body.String(), "sportapp")
		}
		return w
	}
	var first, second plans_postgres_repository.Plan
	t.Run("create and owner isolation", func(t *testing.T) {
		request(t, "GET", "/me/plans", "", nil, 401)
		request(t, "GET", "/me/plans?limit=0", "101", nil, 400)
		w := request(t, "GET", "/me/plans", "101", nil, 200)
		assert.Contains(t, w.Body.String(), `"items":[]`)
		w = request(t, "POST", "/me/plans", "101", sampleInput("Мой план"), 201)
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &first))
		assert.Equal(t, int64(1), first.Version)
		assert.Equal(t, "manual", first.SourceType)
		assert.False(t, first.IsActive)
		assert.Equal(t, 12.5, *first.Days[0].Exercises[0].WeightKg)
		request(t, "GET", "/me/plans/"+first.ID, "102", nil, 404)
		request(t, "PATCH", "/me/plans/"+first.ID, "102", inputOf(first), 404)
		request(t, "POST", "/me/plans/"+first.ID+"/activate", "102", map[string]any{"version": 1}, 404)
		request(t, "DELETE", "/me/plans/"+first.ID, "102", nil, 404)
		request(t, "POST", "/me/plans", "103", sampleInput("Гостевой"), 201)
		request(t, "POST", "/me/plans", "999999", sampleInput("Missing"), 401)
	})
	t.Run("aggregate editing and rollback", func(t *testing.T) {
		input := inputOf(first)
		copyDay := input.Days[0]
		copyDay.ID, copyDay.Title, copyDay.Position = uuid.NewString(), "Копия дня", 1
		input.Days = append(input.Days, copyDay)
		input.Days[0].Position, input.Days[1].Position = 1, 0
		input.Days[0].Exercises = []plans_postgres_repository.Prescription{{ExerciseID: 2, Position: 0, Sets: 2, RepsMin: 6, RepsMax: 10, RestSec: 120}}
		input.Schedule.Entries = []plans_postgres_repository.ScheduleEntry{{Weekday: 2, PlanDayID: copyDay.ID}}
		input.Title = "Изменённый план"
		w := request(t, "PATCH", "/me/plans/"+first.ID, "101", input, 200)
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &first))
		assert.Equal(t, int64(2), first.Version)
		assert.Equal(t, copyDay.ID, first.Days[0].ID)
		assert.Equal(t, 2, first.Days[1].Exercises[0].ExerciseID)
		request(t, "PATCH", "/me/plans/"+first.ID, "101", input, 409)
		invalid := inputOf(first)
		invalid.Days[0].Exercises[0].ExerciseID = 999999
		request(t, "PATCH", "/me/plans/"+first.ID, "101", invalid, 400)
		stored, err := svc.GetPlan(ctx, "101", first.ID)
		require.NoError(t, err)
		assert.Equal(t, first.Version, stored.Version)
		assert.Equal(t, first.Days, stored.Days)
		input = inputOf(first)
		input.Days = input.Days[:1]
		w = request(t, "PATCH", "/me/plans/"+first.ID, "101", input, 200)
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &first))
		assert.Len(t, first.Days, 1)
		w = request(t, "POST", "/me/plans", "101", sampleInput("Второй"), 201)
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &second))
		bad := inputOf(first)
		bad.Days[0].ID = second.Days[0].ID
		bad.Schedule.Entries = []plans_postgres_repository.ScheduleEntry{}
		request(t, "PATCH", "/me/plans/"+first.ID, "101", bad, 400)
	})
	t.Run("concurrent updates", func(t *testing.T) {
		input := inputOf(first)
		var wg sync.WaitGroup
		errs := make(chan error, 2)
		for range 2 {
			wg.Go(func() { _, err := svc.UpdatePlan(ctx, "101", first.ID, input); errs <- err })
		}
		wg.Wait()
		close(errs)
		success, conflict := 0, 0
		for err := range errs {
			if err == nil {
				success++
			} else if errors.Is(err, core_errors.ErrConflict) {
				conflict++
			} else {
				t.Error(err)
			}
		}
		assert.Equal(t, 1, success)
		assert.Equal(t, 1, conflict)
		first, err = svc.GetPlan(ctx, "101", first.ID)
		require.NoError(t, err)
	})
	t.Run("activation serializes per owner", func(t *testing.T) {
		var wg sync.WaitGroup
		errs := make(chan error, 2)
		for _, plan := range []plans_postgres_repository.Plan{first, second} {
			wg.Go(func() { _, err := svc.ActivatePlan(ctx, "101", plan.ID, plan.Version); errs <- err })
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			require.NoError(t, err)
		}
		var active int
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.training_plans WHERE owner_id = 101 AND is_active AND archived_at IS NULL`).Scan(&active))
		assert.Equal(t, 1, active)
		first, err = svc.GetPlan(ctx, "101", first.ID)
		require.NoError(t, err)
		second, err = svc.GetPlan(ctx, "101", second.ID)
		require.NoError(t, err)
		request(t, "POST", "/me/plans/"+first.ID+"/activate", "101", map[string]any{"version": 1}, 409)
	})
	t.Run("pagination and AI provenance", func(t *testing.T) {
		page, err := svc.ListPlans(ctx, "101", 1, "")
		require.NoError(t, err)
		require.Len(t, page.Items, 1)
		require.NotNil(t, page.NextCursor)
		last, err := svc.ListPlans(ctx, "101", 1, *page.NextCursor)
		require.NoError(t, err)
		require.Len(t, last.Items, 1)
		assert.Nil(t, last.NextCursor)
		assert.NotEqual(t, page.Items[0].ID, last.Items[0].ID)
		_, err = svc.ListPlans(ctx, "102", 1, *page.NextCursor)
		assert.ErrorIs(t, err, core_errors.ErrInvalidArgument)
		_, err = pool.Exec(ctx, `UPDATE sportapp.training_plans SET source_type = 'ai' WHERE id = $1`, second.ID)
		require.NoError(t, err)
		second, err = svc.UpdatePlan(ctx, "101", second.ID, inputOf(second))
		require.NoError(t, err)
		assert.Equal(t, "ai", second.SourceType)
	})
	t.Run("reconnect and archive preserve history", func(t *testing.T) {
		reconnected, err := core_pgx_pool.NewPool(ctx, config)
		require.NoError(t, err)
		defer reconnected.Close()
		stored, err := plans_postgres_repository.NewRepository(reconnected).GetPlan(ctx, 101, second.ID)
		require.NoError(t, err)
		assert.Equal(t, second.Days, stored.Days)
		request(t, "DELETE", "/me/plans/"+second.ID, "101", nil, 204)
		request(t, "GET", "/me/plans/"+second.ID, "101", nil, 404)
		request(t, "POST", "/me/plans/"+second.ID+"/activate", "101", map[string]any{"version": second.Version}, 404)
		var revisions, days int
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.plan_revisions WHERE plan_id = $1`, second.ID).Scan(&revisions))
		require.GreaterOrEqual(t, revisions, 3)
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.plan_days WHERE plan_id = $1`, second.ID).Scan(&days))
		assert.Equal(t, 1, days)
	})
	t.Run("migration rollback preserves legacy", func(t *testing.T) {
		apply("000009_personal_plans.down.sql")
		var weight float64
		var body string
		require.NoError(t, pool.QueryRow(ctx, `SELECT weight_kg FROM sportapp.profile WHERE user_id = 101`).Scan(&weight))
		assert.Equal(t, 75.5, weight)
		require.NoError(t, pool.QueryRow(ctx, `SELECT body FROM sportapp.notes WHERE user_id = 101`).Scan(&body))
		assert.Equal(t, "keep", body)
		var legacy bool
		require.NoError(t, pool.QueryRow(ctx, `SELECT (plan_template->>'legacy')::boolean FROM sportapp.user_programs WHERE user_id = 101`).Scan(&legacy))
		assert.True(t, legacy)
		apply("000009_personal_plans.up.sql")
		page, err := svc.ListPlans(ctx, "101", 20, "")
		require.NoError(t, err)
		assert.Empty(t, page.Items)
	})
}
