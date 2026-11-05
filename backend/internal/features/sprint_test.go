package features_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	core_auth "sport_app/internal/core/auth"
	core_pgx_pool "sport_app/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_server "sport_app/internal/core/transport/http/server"
	catalog_postgres_repository "sport_app/internal/features/catalog/repository/postgres"
	catalog_service "sport_app/internal/features/catalog/service"
	catalog_transport_http "sport_app/internal/features/catalog/transport/http"
	generation_postgres_repository "sport_app/internal/features/generation/repository/postgres"
	generation_service "sport_app/internal/features/generation/service"
	generation_transport_http "sport_app/internal/features/generation/transport/http"
	"sport_app/internal/features/mlclient"
	plans_postgres_repository "sport_app/internal/features/plans/repository/postgres"
	plans_service "sport_app/internal/features/plans/service"
	plans_transport_http "sport_app/internal/features/plans/transport/http"
	profile_postgres_repository "sport_app/internal/features/profile/repository/postgres"
	profile_service "sport_app/internal/features/profile/service"
	profile_transport_http "sport_app/internal/features/profile/transport/http"
	publications_postgres_repository "sport_app/internal/features/publications/repository/postgres"
	publications_service "sport_app/internal/features/publications/service"
	publications_transport_http "sport_app/internal/features/publications/transport/http"
	trainers_postgres_repository "sport_app/internal/features/trainers/repository/postgres"
	trainers_service "sport_app/internal/features/trainers/service"
	trainers_transport_http "sport_app/internal/features/trainers/transport/http"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
	workouts_domain "sport_app/internal/features/workouts/domain"
	workouts_postgres_repository "sport_app/internal/features/workouts/repository/postgres"
	workouts_service "sport_app/internal/features/workouts/service"
	workouts_transport_http "sport_app/internal/features/workouts/transport/http"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSprintIntegration(t *testing.T) {
	database := os.Getenv("SPORTIK_SPRINT_TEST_DB")
	if database == "" {
		t.Skip("set SPORTIK_SPRINT_TEST_DB to a new sportik_sprint_test_* database on 127.0.0.1:55443")
	}
	require.True(t, strings.HasPrefix(database, "sportik_sprint_test_"))
	ctx := context.Background()
	pool, err := core_pgx_pool.NewPool(ctx, core_pgx_pool.Config{Host: "127.0.0.1", Port: "55443", User: "sportik_test", Database: database, Timeout: 10 * time.Second})
	require.NoError(t, err)
	defer pool.Close()
	var empty bool
	require.NoError(t, pool.QueryRow(ctx, `SELECT to_regnamespace('sportapp') IS NULL`).Scan(&empty))
	require.True(t, empty, "existing schemas are never cleared")
	apply := func(pattern string) {
		t.Helper()
		files, err := filepath.Glob(filepath.Join("../../migrations", pattern))
		require.NoError(t, err)
		require.Len(t, files, 1)
		raw, err := os.ReadFile(files[0])
		require.NoError(t, err)
		_, err = pool.Exec(ctx, string(raw))
		require.NoError(t, err, pattern)
	}
	for n := 1; n <= 9; n++ {
		apply(fmt.Sprintf("%06d_*.up.sql", n))
	}
	_, err = pool.Exec(ctx, `INSERT INTO sportapp.users(id,is_anonymous,created_at) VALUES(101,FALSE,NOW()),(102,FALSE,NOW()),(103,TRUE,NOW()),(201,FALSE,NOW()),(202,FALSE,NOW()),(301,FALSE,NOW());
 INSERT INTO sportapp.user_roles(user_id,role) VALUES(101,'user'),(102,'user'),(201,'trainer'),(202,'trainer'),(301,'admin');
 INSERT INTO sportapp.profile(user_id,age,height_cm,weight_kg,gender,activity_level,goal,fitness_level,training_days_map,session_duration_min,created_at)
 VALUES(101,26,175,75.5,'м','Средняя активность','Набрать мышцы','Новичок',ARRAY['mon','wed'],60,NOW()),(102,20,180,80,'м','Средняя активность','Набрать мышцы','Новичок',ARRAY['mon'],60,NOW());
 INSERT INTO sportapp.profile_equipment(user_id,equipment_id) SELECT 101,id FROM sportapp.equipment;
 INSERT INTO sportapp.notes(user_id,title,body) VALUES(101,'sprint fixture','preserve');
 INSERT INTO sportapp.user_programs(user_id,is_active,plan_template,plan_exercises) VALUES(101,TRUE,'{"legacy":true}','{"plan":[]}');`)
	require.NoError(t, err)
	require.NoError(t, users_postgres_repository.NewUsersRepository(pool).EnsureExercisesSeeded(ctx))
	catalogRepo := catalog_postgres_repository.NewRepository(pool)
	require.NoError(t, catalogRepo.EnsureReferenceData(ctx))
	apply("000010_*.up.sql")
	apply("000010_*.down.sql")
	apply("000010_*.up.sql")
	apply("000011_*.up.sql")
	apply("000011_*.down.sql")
	apply("000011_*.up.sql")
	var mlMode atomic.Int32
	var mlCalls atomic.Int32
	ml := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mlCalls.Add(1)
		require.Equal(t, "/v2/plans/generate", r.URL.Path)
		require.Equal(t, "Bearer sprint-test-service", r.Header.Get("Authorization"))
		var request mlclient.GenerateInput
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		require.Equal(t, "profile-v2", request.FeatureSchemaVersion)
		require.Equal(t, 75.5, *request.Profile.WeightKg)
		if mlMode.Load() == 1 {
			w.WriteHeader(503)
			return
		}
		if mlMode.Load() == 3 {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"request_id":"wrong","model_version":"test","days":[],"fallback":false}`))
			return
		}
		if mlMode.Load() == 4 {
			_, err := pool.Exec(ctx, `UPDATE sportapp.profile SET version=version+1 WHERE user_id=101`)
			require.NoError(t, err)
		}
		exercise := 1
		if mlMode.Load() == 2 {
			exercise = 2
		}
		days := []mlclient.GeneratedDay{}
		for _, day := range request.Profile.TrainingDays {
			days = append(days, mlclient.GeneratedDay{Weekday: day, Title: "AI day", Exercises: []plans_postgres_repository.Prescription{{ExerciseID: exercise, Position: 0, Sets: 3, RepsMin: 8, RepsMax: 12, RestSec: 90}}})
		}
		require.NoError(t, json.NewEncoder(w).Encode(mlclient.GenerateResult{RequestID: request.RequestID, ModelVersion: "test-transformer-v1", FeatureSchemaVersion: "profile-v2", Split: "Full body", Days: days, Warnings: []string{}, Fallback: false}))
	}))
	defer ml.Close()
	jwt := core_auth.NewJWT(core_auth.Config{Secret: "sprint-tests-only", TokenTTL: time.Hour})
	router := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion2)
	planSvc := plans_service.NewService(plans_postgres_repository.NewRepository(pool))
	workoutRepo := workouts_postgres_repository.NewRepository(pool)
	router.RegisterRoutes(catalog_transport_http.NewCatalogHTTPHandler(catalog_service.NewService(catalogRepo), jwt).Routes()...)
	router.RegisterRoutes(profile_transport_http.NewProfileHTTPHandler(profile_service.NewService(profile_postgres_repository.NewRepository(pool)), jwt).Routes()...)
	router.RegisterRoutes(plans_transport_http.NewPlansHTTPHandler(planSvc, jwt).Routes()...)
	router.RegisterRoutes(trainers_transport_http.NewHandler(trainers_service.NewService(trainers_postgres_repository.NewRepository(pool)), jwt).Routes()...)
	router.RegisterRoutes(publications_transport_http.NewHandler(publications_service.NewService(publications_postgres_repository.NewRepository(pool)), jwt).Routes()...)
	router.RegisterRoutes(workouts_transport_http.NewHandler(workouts_service.NewService(workoutRepo), jwt).Routes()...)
	router.RegisterRoutes(generation_transport_http.NewHandler(generation_service.NewService(generation_postgres_repository.NewRepository(pool), mlclient.NewV2Client(mlclient.Config{BaseURL: ml.URL, ServiceToken: "sprint-test-service"})), jwt).Routes()...)
	handler := core_http_middleware.RequestID()(http.StripPrefix("/api/v2", router))
	tokens := map[int]string{}
	for _, id := range []int{101, 102, 103, 201, 202, 301} {
		tokens[id], err = jwt.CreateToken(fmt.Sprint(id), id == 103)
		require.NoError(t, err)
	}
	call := func(method, path string, user int, body any, key string) *httptest.ResponseRecorder {
		var raw []byte
		if body != nil {
			var err error
			raw, err = json.Marshal(body)
			require.NoError(t, err)
		}
		req := httptest.NewRequest(method, "/api/v2"+path, bytes.NewReader(raw))
		if user != 0 {
			req.Header.Set("Authorization", "Bearer "+tokens[user])
		}
		if key != "" {
			req.Header.Set("Idempotency-Key", key)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		return w
	}
	check := func(method, path string, user int, body any, key string, status int) *httptest.ResponseRecorder {
		t.Helper()
		w := call(method, path, user, body, key)
		require.Equal(t, status, w.Code, w.Body.String())
		if status >= 400 {
			require.Contains(t, w.Body.String(), `"request_id"`)
			require.NotContains(t, w.Body.String(), "sportapp")
		}
		return w
	}
	var program publications_postgres_repository.Program
	t.Run("trainer role verification privacy and concurrent decision", func(t *testing.T) {
		for _, id := range []int{0, 101, 103} {
			status := 403
			if id == 0 {
				status = 401
			}
			check("PATCH", "/trainer/profile", id, map[string]any{"display_name": "Denied"}, "", status)
		}
		w := check("PATCH", "/trainer/profile", 201, map[string]any{"display_name": "Fixture trainer", "specializations": []string{"strength"}, "experience_years": 3, "qualifications": "Public certificate"}, "", 200)
		var p trainers_postgres_repository.Profile
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &p))
		require.False(t, p.Verified)
		check("PATCH", "/trainer/profile", 201, map[string]any{"display_name": "Fixture trainer", "verified": true}, "", 400)
		check("POST", "/trainer/verification", 201, map[string]any{"qualification": "Private certificate", "document_refs": []string{"private/test-document"}}, "", 201)
		w = check("GET", "/trainer/verification", 201, nil, "", 200)
		var v trainers_postgres_repository.Verification
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &v))
		require.Equal(t, "pending", v.Status)
		check("POST", "/trainer/verification", 201, map[string]any{"qualification": "Again", "document_refs": []string{"another"}}, "", 409)
		check("GET", "/admin/verifications", 101, nil, "", 403)
		check("PATCH", "/admin/verifications/"+v.ID, 201, map[string]any{"status": "approved", "reason": "self"}, "", 403)
		check("PATCH", "/admin/verifications/"+v.ID, 301, map[string]any{"status": "rejected", "reason": " "}, "", 400)
		statuses := make(chan int, 2)
		var wg sync.WaitGroup
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				statuses <- call("PATCH", "/admin/verifications/"+v.ID, 301, map[string]any{"status": "rejected", "reason": "Missing evidence"}, "").Code
			}()
		}
		wg.Wait()
		close(statuses)
		codes := map[int]int{}
		for c := range statuses {
			codes[c]++
		}
		require.Equal(t, map[int]int{200: 1, 409: 1}, codes)
		check("POST", "/trainer/verification", 201, map[string]any{"qualification": "Updated certificate", "document_refs": []string{"private/new-document"}}, "", 201)
		w = check("GET", "/trainer/verification", 201, nil, "", 200)
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &v))
		check("PATCH", "/admin/verifications/"+v.ID, 301, map[string]any{"status": "approved", "reason": "Evidence checked"}, "", 200)
		w = check("GET", "/trainers/201", 0, nil, "", 200)
		require.NotContains(t, w.Body.String(), "document_refs")
		require.NotContains(t, w.Body.String(), "private/")
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &p))
		require.True(t, p.Verified)
		check("PATCH", "/trainer/profile", 202, map[string]any{"display_name": "Other trainer"}, "", 200)
		pageResponse := check("GET", "/admin/trainers?limit=1", 301, nil, "", 200)
		var trainersPage trainers_postgres_repository.Page[trainers_postgres_repository.Profile]
		require.NoError(t, json.Unmarshal(pageResponse.Body.Bytes(), &trainersPage))
		require.Len(t, trainersPage.Items, 1)
		require.NotNil(t, trainersPage.NextCursor)
		nextResponse := check("GET", "/admin/trainers?limit=1&cursor="+*trainersPage.NextCursor, 301, nil, "", 200)
		var nextTrainers trainers_postgres_repository.Page[trainers_postgres_repository.Profile]
		require.NoError(t, json.Unmarshal(nextResponse.Body.Bytes(), &nextTrainers))
		require.Len(t, nextTrainers.Items, 1)
		require.NotEqual(t, trainersPage.Items[0].ID, nextTrainers.Items[0].ID)
		verResponse := check("GET", "/admin/verifications?limit=1", 301, nil, "", 200)
		var verPage trainers_postgres_repository.Page[trainers_postgres_repository.Verification]
		require.NoError(t, json.Unmarshal(verResponse.Body.Bytes(), &verPage))
		require.Len(t, verPage.Items, 1)
		require.NotNil(t, verPage.NextCursor)
		check("GET", "/admin/verifications?limit=1&cursor="+*verPage.NextCursor, 101, nil, "", 400)
		var count int
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.publication_audit WHERE action='trainer_verification' AND actor_id=301 AND request_id<>''`).Scan(&count))
		require.Equal(t, 2, count)
	})
	t.Run("program publication immutable version and access", func(t *testing.T) {
		draft := publications_postgres_repository.Draft{Title: "Fixture program", Description: "Safe program", Goal: "muscle_gain", Level: "beginner", DurationWeeks: 4, SessionsPerWeek: 1, EquipmentIDs: []string{"barbell", "dumbbell", "bench", "machine", "cable", "pull_up_bar", "parallel_bars", "kettlebell"}, CategoryIDs: []string{}, Currency: "RUB", Days: []plans_postgres_repository.Day{{Title: "Day A", Position: 0, Exercises: []plans_postgres_repository.Prescription{{ExerciseID: 1, Position: 0, Sets: 3, RepsMin: 8, RepsMax: 12, RestSec: 90}}}}}
		w := check("POST", "/trainer/programs", 201, draft, "", 201)
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &program))
		require.Equal(t, "draft", program.Status)
		check("GET", "/catalog/programs/"+program.ID, 101, nil, "", 404)
		check("GET", "/trainer/programs/"+program.ID, 202, nil, "", 404)
		check("PATCH", "/trainer/programs/"+program.ID, 202, map[string]any{"version": program.Version, "title": "hijack"}, "", 404)
		w = check("PATCH", "/trainer/programs/"+program.ID, 201, map[string]any{"version": program.Version, "title": "New title"}, "", 200)
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &program))
		check("PATCH", "/trainer/programs/"+program.ID, 201, map[string]any{"version": 1, "title": "stale"}, "", 409)
		check("POST", "/trainer/programs/"+program.ID+"/publish", 201, map[string]any{"version": program.Version}, "", 202)
		check("PATCH", "/trainer/programs/"+program.ID, 201, map[string]any{"version": program.Version + 1, "title": "change in review"}, "", 409)
		check("PATCH", "/admin/programs/"+program.ID+"/moderation", 101, map[string]any{"status": "approved", "reason": "hijack"}, "", 403)
		check("PATCH", "/admin/programs/"+program.ID+"/moderation", 301, map[string]any{"status": "approved", "reason": "Checked"}, "", 200)
		w = check("GET", "/catalog/programs/"+program.ID, 101, nil, "", 200)
		var published map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &published))
		require.Equal(t, "New title", published["title"])
		check("POST", "/programs/"+program.ID+"/access", 101, nil, "", 201)
		w = check("POST", "/trainer/programs/"+program.ID+"/versions", 201, nil, "", 201)
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &program))
		require.Equal(t, "draft", program.Status)
		w = check("PATCH", "/trainer/programs/"+program.ID, 201, map[string]any{"version": program.Version, "title": "Version two"}, "", 200)
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &program))
		w = check("GET", "/catalog/programs/"+program.ID, 101, nil, "", 200)
		var unchanged map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &unchanged))
		require.Equal(t, published["title"], unchanged["title"])
		require.Equal(t, published["published_version_id"], unchanged["published_version_id"])
		check("POST", "/trainer/programs", 201, draft, "", 201)
		pageResponse := check("GET", "/admin/programs?limit=1", 301, nil, "", 200)
		var programsPage publications_postgres_repository.Page
		require.NoError(t, json.Unmarshal(pageResponse.Body.Bytes(), &programsPage))
		require.Len(t, programsPage.Items, 1)
		require.NotNil(t, programsPage.NextCursor)
		pageResponse = check("GET", "/admin/programs?limit=1&cursor="+*programsPage.NextCursor, 301, nil, "", 200)
		var nextPrograms publications_postgres_repository.Page
		require.NoError(t, json.Unmarshal(pageResponse.Body.Bytes(), &nextPrograms))
		require.Len(t, nextPrograms.Items, 1)
		require.NotEqual(t, programsPage.Items[0].ID, nextPrograms.Items[0].ID)
		check("POST", "/trainer/programs/"+program.ID+"/publish", 201, map[string]any{"version": program.Version}, "", 202)
		check("PATCH", "/admin/programs/"+program.ID+"/moderation", 301, map[string]any{"program_version_id": published["published_version_id"], "status": "approved", "reason": "Wrong version"}, "", 409)
		check("PATCH", "/admin/programs/"+program.ID+"/moderation", 301, map[string]any{"program_version_id": program.VersionID, "status": "approved", "reason": "New version checked"}, "", 200)
		var oldTitle string
		require.NoError(t, pool.QueryRow(ctx, `SELECT title FROM sportapp.program_versions WHERE id=$1`, published["published_version_id"]).Scan(&oldTitle))
		require.Equal(t, "New title", oldTitle)
	})
	var plan plans_postgres_repository.Plan
	var workout workouts_domain.Input
	t.Run("workout revision idempotency concurrency metrics owner", func(t *testing.T) {
		day := uuid.NewString()
		input := plans_postgres_repository.Input{Title: "Workout plan", Days: []plans_postgres_repository.Day{{ID: day, Title: "Day A", Position: 0, Exercises: []plans_postgres_repository.Prescription{{ExerciseID: 1, Position: 0, Sets: 3, RepsMin: 8, RepsMax: 12, RestSec: 90}}}}, Schedule: plans_postgres_repository.Schedule{TimeZone: "Europe/Moscow", StartDate: "2026-10-01", Entries: []plans_postgres_repository.ScheduleEntry{{Weekday: 1, PlanDayID: day}}}}
		plan, err = planSvc.CreatePlan(ctx, "101", input)
		require.NoError(t, err)
		_, err = planSvc.ActivatePlan(ctx, "101", plan.ID, plan.Version)
		require.NoError(t, err)
		input.Version = 2
		input.Days[0].Exercises[0].ExerciseID = 2
		_, err = planSvc.UpdatePlan(ctx, "101", plan.ID, input)
		require.NoError(t, err)
		workout = workouts_domain.Input{ClientSessionID: uuid.NewString(), PlanID: plan.ID, PlanVersion: 1, PlanDayID: day, StartedAt: time.Now().UTC().Add(-time.Hour), FinishedAt: time.Now().UTC().Add(-time.Minute), Slots: []workouts_domain.Slot{{ExerciseID: 1, Position: 0, Sets: []workouts_domain.Set{{Reps: 20, WeightKg: 20, IsWarmup: true}, {Reps: 10, WeightKg: 60}, {Reps: 8, WeightKg: 60}}}}}
		check("POST", "/workouts/complete", 101, workout, "", 400)
		check("POST", "/workouts/complete", 102, workout, "foreign", 404)
		statuses := make(chan *httptest.ResponseRecorder, 2)
		var wg sync.WaitGroup
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() { defer wg.Done(); statuses <- call("POST", "/workouts/complete", 101, workout, "session-key") }()
		}
		wg.Wait()
		close(statuses)
		var result workouts_domain.Result
		for w := range statuses {
			require.Equal(t, 201, w.Code, w.Body.String())
			var value workouts_domain.Result
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &value))
			if result.SessionID != "" {
				require.Equal(t, result, value)
			}
			result = value
		}
		require.Equal(t, 1080.0, result.Metrics.VolumeKg)
		require.Equal(t, 2, result.Metrics.WorkingSets)
		w := check("POST", "/workouts/complete", 101, workout, "another-key", 201)
		var repeated workouts_domain.Result
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &repeated))
		require.Equal(t, result, repeated)
		other := workout
		other.PlanVersion = 2
		check("POST", "/workouts/complete", 101, other, "session-key", 409)
		check("POST", "/workouts/complete", 101, other, "different-key", 409)
		check("GET", "/workouts/"+result.SessionID, 102, nil, "", 404)
		check("GET", "/workouts/"+result.SessionID, 101, nil, "", 200)
		check("GET", "/workouts?limit=1", 101, nil, "", 200)
		check("GET", "/progress/exercises/1", 101, nil, "", 200)
		check("POST", "/progress/weight", 101, map[string]any{"day": "2026-10-29", "weight_kg": 80.5}, "", 201)
		check("POST", "/progress/weight", 101, map[string]any{"day": "2026-10-30", "weight_kg": 79.25}, "", 201)
		w = check("GET", "/progress?period=all&time_zone=Europe/Moscow", 101, nil, "", 200)
		var p workouts_domain.Progress
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &p))
		require.Equal(t, 1, p.WorkoutCount)
		require.Equal(t, 1080.0, p.VolumeKg)
		require.Equal(t, -1.25, *p.WeightChangeKg)
		check("GET", "/progress/summaries?period=week", 101, nil, "", 200)
		check("GET", "/progress/summaries?period=month", 101, nil, "", 200)
		check("GET", "/progress?time_zone=Local", 101, nil, "", 400)
		require.NoError(t, planSvc.ArchivePlan(ctx, "101", plan.ID))
		check("GET", "/workouts/"+result.SessionID, 101, nil, "", 200)
	})
	t.Run("generation profile constraints cached result and failure isolation", func(t *testing.T) {
		activeDay := uuid.NewString()
		activeInput := plans_postgres_repository.Input{Title: "Keep active", Days: []plans_postgres_repository.Day{{ID: activeDay, Title: "Keep day", Position: 0, Exercises: []plans_postgres_repository.Prescription{{ExerciseID: 1, Position: 0, Sets: 3, RepsMin: 8, RepsMax: 12, RestSec: 90}}}}, Schedule: plans_postgres_repository.Schedule{TimeZone: "Europe/Moscow", StartDate: "2026-10-01", Entries: []plans_postgres_repository.ScheduleEntry{}}}
		active, err := planSvc.CreatePlan(ctx, "101", activeInput)
		require.NoError(t, err)
		_, err = planSvc.ActivatePlan(ctx, "101", active.ID, active.Version)
		require.NoError(t, err)

		check("POST", "/plans/generate", 101, map[string]int{"profile_version": 1}, "", 400)
		w := check("POST", "/plans/generate", 101, map[string]int{"profile_version": 1}, "generate-key", 201)
		var result generation_postgres_repository.Result
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
		calls := mlCalls.Load()
		w = check("POST", "/plans/generate", 101, map[string]int{"profile_version": 1}, "generate-key", 201)
		var repeated generation_postgres_repository.Result
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &repeated))
		require.Equal(t, result, repeated)
		require.Equal(t, calls, mlCalls.Load())
		check("POST", "/plans/generate", 101, map[string]int{"profile_version": 2}, "generate-key", 409)
		w = check("GET", "/me/plans/"+result.PlanID, 101, nil, "", 200)
		var generated plans_postgres_repository.Plan
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &generated))
		require.Equal(t, "ai", generated.SourceType)
		require.False(t, generated.IsActive)
		require.Len(t, generated.Days, 2)
		patch := profile_postgres_repository.Patch{Version: 1, ExcludedExerciseIDs: &[]int{2}}
		profileRepo := profile_postgres_repository.NewRepository(pool)
		_, err = profileRepo.UpdateProfile(ctx, 101, patch)
		require.NoError(t, err)
		mlMode.Store(2)
		check("POST", "/plans/generate", 101, map[string]int{"profile_version": 2}, "excluded", 422)
		_, err = pool.Exec(ctx, `DELETE FROM sportapp.profile_equipment WHERE user_id=101; UPDATE sportapp.profile SET version=version+1 WHERE user_id=101; INSERT INTO sportapp.exercise_equipment(exercise_id,equipment_id) VALUES(1,'barbell') ON CONFLICT DO NOTHING`)
		require.NoError(t, err)
		mlMode.Store(0)
		check("POST", "/plans/generate", 101, map[string]int{"profile_version": 3}, "equipment", 422)
		mlMode.Store(1)
		check("POST", "/plans/generate", 101, map[string]int{"profile_version": 3}, "unavailable", 503)
		mlMode.Store(3)
		check("POST", "/plans/generate", 101, map[string]int{"profile_version": 3}, "malformed", 422)
		check("GET", "/me/plans/"+result.PlanID, 101, nil, "", 200)
		var count int
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.plan_generations WHERE user_id=101 AND model_version='test-transformer-v1' AND request_id<>''`).Scan(&count))
		require.Equal(t, 1, count)
		mlMode.Store(4)
		check("POST", "/plans/generate", 101, map[string]int{"profile_version": 3}, "changed-profile", 409)
		_, err = pool.Exec(ctx, `INSERT INTO sportapp.profile_equipment(user_id,equipment_id) SELECT 101,id FROM sportapp.equipment`)
		require.NoError(t, err)
		mlMode.Store(0)
		replies := make(chan *httptest.ResponseRecorder, 2)
		var wg sync.WaitGroup
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				replies <- call("POST", "/plans/generate", 101, map[string]int{"profile_version": 4}, "concurrent-generation")
			}()
		}
		wg.Wait()
		close(replies)
		savedID := ""
		for response := range replies {
			require.Equal(t, 201, response.Code, response.Body.String())
			var value generation_postgres_repository.Result
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &value))
			if savedID != "" {
				require.Equal(t, savedID, value.PlanID)
			}
			savedID = value.PlanID
		}
		current, err := planSvc.GetPlan(ctx, "101", active.ID)
		require.NoError(t, err)
		require.True(t, current.IsActive)
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.plan_generations WHERE user_id=101`).Scan(&count))
		require.Equal(t, 2, count)
	})
	t.Run("reconnect and rollback preserve legacy data", func(t *testing.T) {
		second, err := core_pgx_pool.NewPool(ctx, core_pgx_pool.Config{Host: "127.0.0.1", Port: "55443", User: "sportik_test", Database: database, Timeout: 10 * time.Second})
		require.NoError(t, err)
		data, err := workouts_postgres_repository.NewRepository(second).Data(ctx, 101)
		require.NoError(t, err)
		require.Len(t, data.Workouts, 1)
		second.Close()
		apply("000011_*.down.sql")
		apply("000011_*.up.sql")
		apply("000010_*.down.sql")
		apply("000010_*.up.sql")
		var weight float64
		require.NoError(t, pool.QueryRow(ctx, `SELECT weight_kg::double precision FROM sportapp.profile WHERE user_id=101`).Scan(&weight))
		require.Equal(t, 75.5, weight)
		var count int
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.notes WHERE user_id=101 AND body='preserve'`).Scan(&count))
		require.Equal(t, 1, count)
		var legacy bool
		require.NoError(t, pool.QueryRow(ctx, `SELECT (plan_template->>'legacy')::boolean FROM sportapp.user_programs WHERE user_id=101`).Scan(&legacy))
		require.True(t, legacy)
		var title string
		require.NoError(t, pool.QueryRow(ctx, `SELECT v.title FROM sportapp.programs p JOIN sportapp.program_versions v ON v.id=p.published_version_id WHERE p.id=$1`, program.ID).Scan(&title))
		require.Equal(t, "Version two", title)
	})
}
