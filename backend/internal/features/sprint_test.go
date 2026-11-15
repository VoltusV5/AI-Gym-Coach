package features_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	insights_repository "sport_app/internal/features/insights/repository/postgres"
	insights_service "sport_app/internal/features/insights/service"
	insights_http "sport_app/internal/features/insights/transport/http"
	users_service "sport_app/internal/features/users/service"
	users_http "sport_app/internal/features/users/transport/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	core_auth "sport_app/internal/core/auth"
	core_pgx_pool "sport_app/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_server "sport_app/internal/core/transport/http/server"
	administration_postgres_repository "sport_app/internal/features/administration/repository/postgres"
	administration_service "sport_app/internal/features/administration/service"
	administration_transport_http "sport_app/internal/features/administration/transport/http"
	catalog_postgres_repository "sport_app/internal/features/catalog/repository/postgres"
	catalog_service "sport_app/internal/features/catalog/service"
	catalog_transport_http "sport_app/internal/features/catalog/transport/http"
	commerce_postgres_repository "sport_app/internal/features/commerce/repository/postgres"
	commerce_service "sport_app/internal/features/commerce/service"
	commerce_transport_http "sport_app/internal/features/commerce/transport/http"
	favorites_postgres_repository "sport_app/internal/features/favorites/repository/postgres"
	favorites_service "sport_app/internal/features/favorites/service"
	favorites_transport_http "sport_app/internal/features/favorites/transport/http"
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
	sharing_postgres_repository "sport_app/internal/features/sharing/repository/postgres"
	sharing_service "sport_app/internal/features/sharing/service"
	sharing_transport_http "sport_app/internal/features/sharing/transport/http"
	social_postgres_repository "sport_app/internal/features/social/repository/postgres"
	social_service "sport_app/internal/features/social/service"
	social_transport_http "sport_app/internal/features/social/transport/http"
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
	_, err = pool.Exec(ctx, `INSERT INTO sportapp.exercises(exercises_name,muscular_group,muscular_subgroup,working_weights,safe_for_injuries) VALUES('Жим ногами лёжа','Ноги','квадрицепс',35,FALSE)`)
	require.NoError(t, err)
	apply("000019_*.up.sql")
	apply("000019_*.down.sql")
	apply("000019_*.up.sql")
	require.NoError(t, users_postgres_repository.NewUsersRepository(pool).EnsureExercisesSeeded(ctx))
	catalogRepo := catalog_postgres_repository.NewRepository(pool)
	require.NoError(t, catalogRepo.EnsureReferenceData(ctx))
	apply("000010_*.up.sql")
	apply("000010_*.down.sql")
	apply("000010_*.up.sql")
	apply("000011_*.up.sql")
	apply("000011_*.down.sql")
	apply("000011_*.up.sql")
	apply("000012_*.up.sql")
	apply("000012_*.down.sql")
	apply("000012_*.up.sql")
	apply("000013_*.up.sql")
	apply("000013_*.down.sql")
	apply("000013_*.up.sql")
	apply("000014_*.up.sql")
	apply("000014_*.down.sql")
	apply("000014_*.up.sql")
	apply("000015_*.up.sql")
	apply("000015_*.down.sql")
	apply("000015_*.up.sql")
	apply("000016_*.up.sql")
	apply("000016_*.down.sql")
	apply("000016_*.up.sql")
	apply("000017_*.up.sql")
	apply("000017_*.down.sql")
	apply("000017_*.up.sql")
	apply("000018_*.up.sql")
	apply("000018_*.down.sql")
	apply("000018_*.up.sql")
	var mlMode atomic.Int32
	var mlCalls atomic.Int32
	ml := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mlCalls.Add(1)
		require.Equal(t, "/v2/plans/generate", r.URL.Path)
		require.Equal(t, "Bearer sprint-test-service", r.Header.Get("Authorization"))
		var request mlclient.GenerateInput
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		require.Equal(t, "profile-v2", request.FeatureSchemaVersion)
		if os.Getenv("SPORTIK_POSTMAN_TESTS") != "1" || *request.Profile.WeightKg != 80.5 {
			require.Equal(t, 75.5, *request.Profile.WeightKg)
		}
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
		if os.Getenv("SPORTIK_POSTMAN_TESTS") == "1" && *request.Profile.WeightKg == 80.5 && len(request.Exercises) > 0 {
			for _, candidate := range request.Exercises {
				compatible := true
				for _, required := range candidate.EquipmentIDs {
					present := false
					for _, available := range request.Profile.EquipmentIDs {
						if available == required {
							present = true
						}
					}
					if !present {
						compatible = false
					}
				}
				if compatible {
					exercise = candidate.ExerciseID
					break
				}
			}
		}
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
	router.RegisterRoutes(insights_http.NewHandler(insights_service.NewService(insights_repository.NewRepository(pool), nil, 14), jwt).Routes()...)
	router.RegisterRoutes(sharing_transport_http.NewHandler(sharing_service.NewService(sharing_postgres_repository.NewRepository(pool), "http://localhost:5173"), jwt).Routes()...)
	router.RegisterRoutes(social_transport_http.NewHandler(social_service.NewService(social_postgres_repository.NewRepository(pool), 3), jwt).Routes()...)
	router.RegisterRoutes(favorites_transport_http.NewHandler(favorites_service.NewService(favorites_postgres_repository.NewRepository(pool)), jwt).Routes()...)
	router.RegisterRoutes(commerce_transport_http.NewHandler(commerce_service.NewService(commerce_postgres_repository.NewRepository(pool), commerce_service.Config{Stub: true, WebhookSecret: "sprint-stub-signing-secret"}), jwt).Routes()...)
	router.RegisterRoutes(administration_transport_http.NewHandler(administration_service.NewService(administration_postgres_repository.NewRepository(pool)), jwt).Routes()...)
	router.RegisterRoutes(catalog_transport_http.NewCatalogHTTPHandler(catalog_service.NewService(catalogRepo), jwt).Routes()...)
	router.RegisterRoutes(profile_transport_http.NewProfileHTTPHandler(profile_service.NewService(profile_postgres_repository.NewRepository(pool)), jwt).Routes()...)
	router.RegisterRoutes(plans_transport_http.NewPlansHTTPHandler(planSvc, jwt).Routes()...)
	router.RegisterRoutes(trainers_transport_http.NewHandler(trainers_service.NewService(trainers_postgres_repository.NewRepository(pool)), jwt).Routes()...)
	router.RegisterRoutes(publications_transport_http.NewHandler(publications_service.NewService(publications_postgres_repository.NewRepository(pool)), jwt).Routes()...)
	router.RegisterRoutes(workouts_transport_http.NewHandler(workouts_service.NewService(workoutRepo), jwt).Routes()...)
	router.RegisterRoutes(generation_transport_http.NewHandler(generation_service.NewService(generation_postgres_repository.NewRepository(pool), mlclient.NewV2Client(mlclient.Config{BaseURL: ml.URL, ServiceToken: "sprint-test-service"})), jwt).Routes()...)
	handler := core_http_middleware.RequestID()(core_http_middleware.AccountStatus(jwt, pool)(http.StripPrefix("/api/v2", router)))
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
	t.Run("account roles blocks and audit", func(t *testing.T) {
		created, err := users_postgres_repository.NewUsersRepository(pool).CreateUser(ctx, "registration-fixture@example.com", "synthetic-hash")
		require.NoError(t, err)
		var initialRoles int
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.user_roles WHERE user_id=$1 AND role='user'`, created.ID).Scan(&initialRoles))
		require.Equal(t, 1, initialRoles)
		_, err = pool.Exec(ctx, `INSERT INTO sportapp.users(id,is_anonymous,created_at,email) VALUES(104,FALSE,NOW(),'fixture104@example.com')`)
		require.NoError(t, err)
		tokens[104], err = jwt.CreateToken("104", false)
		require.NoError(t, err)
		check("GET", "/me/session", 0, nil, "", 401)
		w := check("GET", "/me/session", 104, nil, "", 200)
		require.Contains(t, w.Body.String(), `"roles":["user"]`)
		check("GET", "/admin/users", 104, nil, "", 403)
		check("PATCH", "/admin/users/104", 104, map[string]any{"roles": []string{"admin"}, "reason": "self"}, "", 403)
		check("PATCH", "/admin/users/301", 301, map[string]any{"blocked": true, "reason": "self"}, "", 409)
		check("PATCH", "/admin/users/104", 301, map[string]any{"roles": []string{"root"}, "reason": "bad"}, "", 400)
		check("PATCH", "/admin/users/104", 301, map[string]any{"blocked": true, "reason": "fixture block"}, "", 200)
		check("GET", "/me/profile", 104, nil, "", 403)
		v1 := core_http_middleware.AccountStatus(jwt, pool)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Fatal("blocked v1 handler executed") }))
		req := httptest.NewRequest("GET", "/api/v1/profile", nil)
		req.Header.Set("Authorization", "Bearer "+tokens[104])
		rec := httptest.NewRecorder()
		v1.ServeHTTP(rec, req)
		require.Equal(t, 403, rec.Code)
		check("GET", "/me/session", 104, nil, "", 403)
		check("PATCH", "/admin/users/104", 301, map[string]any{"blocked": false, "roles": []string{"user", "trainer"}, "reason": "restore fixture"}, "", 200)
		w = check("GET", "/me/session", 104, nil, "", 200)
		require.Contains(t, w.Body.String(), `"trainer"`)
		w = check("GET", "/admin/users?limit=1", 301, nil, "", 200)
		var page administration_postgres_repository.Page
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &page))
		require.Len(t, page.Items, 1)
		require.NotNil(t, page.NextCursor)
		check("GET", "/admin/users?limit=1&cursor="+*page.NextCursor, 104, nil, "", 400)
		check("PATCH", "/admin/users/104", 301, map[string]any{"reason": ""}, "", 400)
		var count int
		require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM sportapp.account_audit WHERE target_id=104 AND request_id<>''`).Scan(&count))
		require.Equal(t, 2, count)
		var blockedBefore, blockedAfter bool
		require.NoError(t, pool.QueryRow(ctx, `SELECT (before_state->>'blocked')::boolean,(after_state->>'blocked')::boolean FROM sportapp.account_audit WHERE target_id=104 AND reason='fixture block'`).Scan(&blockedBefore, &blockedAfter))
		require.False(t, blockedBefore)
		require.True(t, blockedAfter)
		check("PATCH", "/admin/users/999999", 301, map[string]any{"blocked": true, "reason": "missing"}, "", 404)
	})

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
	t.Run("generation rejects exercises unsafe for the current profile", func(t *testing.T) {
		_, err := pool.Exec(ctx, `UPDATE sportapp.profile SET injuries_notes=TRUE WHERE user_id=101; UPDATE sportapp.exercises SET safe_for_injuries=FALSE WHERE id=1`)
		require.NoError(t, err)
		defer func() {
			_, err := pool.Exec(ctx, `UPDATE sportapp.profile SET injuries_notes=FALSE WHERE user_id=101`)
			require.NoError(t, err)
		}()
		var before, after int
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.plan_generations WHERE user_id=101`).Scan(&before))
		check("POST", "/plans/generate", 101, map[string]int{"profile_version": 4}, "unsafe-exercise", 422)
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.plan_generations WHERE user_id=101`).Scan(&after))
		require.Equal(t, before, after)
	})

	t.Run("favorites registration duplicate owner and visibility", func(t *testing.T) {
		path := "/me/favorites/" + program.ID
		check("PUT", path, 0, nil, "", 401)
		check("PUT", path, 103, nil, "", 403)
		check("PUT", path, 101, nil, "", 200)
		check("PUT", path, 101, nil, "", 200)
		w := check("GET", "/me/favorites", 101, nil, "", 200)
		require.Contains(t, w.Body.String(), program.ID)
		w = check("GET", "/catalog/programs/"+program.ID, 101, nil, "", 200)
		require.Contains(t, w.Body.String(), `"is_favorite":true`)
		w = check("GET", "/me/favorites", 102, nil, "", 200)
		require.NotContains(t, w.Body.String(), program.ID)
		check("DELETE", path, 102, nil, "", 204)
		w = check("GET", "/me/favorites", 101, nil, "", 200)
		require.Contains(t, w.Body.String(), program.ID)
		check("DELETE", path, 101, nil, "", 204)
		check("DELETE", path, 101, nil, "", 204)
		check("PUT", "/me/favorites/"+uuid.NewString(), 101, nil, "", 404)
		check("GET", "/catalog/programs?equipment=", 101, nil, "", 200)
		w = check("GET", "/catalog/programs?sessions_per_week=1", 101, nil, "", 200)
		require.Contains(t, w.Body.String(), program.ID)
		w = check("GET", "/catalog/programs?sessions_per_week=7", 101, nil, "", 200)
		require.NotContains(t, w.Body.String(), program.ID)
		check("GET", "/catalog/programs?sessions_per_week=8", 101, nil, "", 400)
		check("GET", "/catalog/programs?sessions_per_week=0", 101, nil, "", 400)

	})

	t.Run("stub purchase ledger refunds payouts and signatures", func(t *testing.T) {
		paidID, versionID := uuid.NewString(), uuid.NewString()
		_, err := pool.Exec(ctx, `INSERT INTO sportapp.programs(id,trainer_id,status) VALUES($1,201,'active');`, paidID)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `INSERT INTO sportapp.program_versions(id,program_id,number,title,description,goal,level,duration_weeks,sessions_per_week,price_minor,currency,status) VALUES($1,$2,1,'Paid fixture','Synthetic paid fixture','muscle_gain','beginner',4,1,100000,'RUB','published')`, versionID, paidID)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `UPDATE sportapp.programs SET published_version_id=$1 WHERE id=$2`, versionID, paidID)
		require.NoError(t, err)
		input := commerce_postgres_repository.OrderInput{ProgramID: paidID, ProgramVersionID: versionID}
		check("POST", "/orders", 103, input, "guest-order", 403)
		check("POST", "/orders", 101, input, "", 400)
		replies := make(chan *httptest.ResponseRecorder, 2)
		var wg sync.WaitGroup
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() { defer wg.Done(); replies <- call("POST", "/orders", 101, input, "same-order") }()
		}
		wg.Wait()
		close(replies)
		var order commerce_postgres_repository.Order
		for w := range replies {
			require.Equal(t, 201, w.Code, w.Body.String())
			var o commerce_postgres_repository.Order
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &o))
			if order.ID != "" {
				require.Equal(t, order.ID, o.ID)
			}
			order = o
		}
		require.Equal(t, "pending", order.Status)
		require.NotNil(t, order.CheckoutURL)
		check("GET", "/orders/"+order.ID, 102, nil, "", 404)
		check("POST", "/payments/sandbox/"+order.ID, 102, map[string]string{"action": "pay"}, "", 404)
		check("POST", "/programs/"+paidID+"/access", 101, nil, "", 422)
		check("GET", "/admin/orders", 101, nil, "", 403)
		changed := input
		changed.ProgramVersionID = uuid.NewString()
		check("POST", "/orders", 101, changed, "same-order", 409)
		event := commerce_postgres_repository.Event{EventID: "fixture-payment-event", PaymentID: "stub-" + order.ID, OrderID: order.ID, Type: "payment_succeeded", AmountMinor: 100000, Currency: "RUB"}
		webhook := func(e commerce_postgres_repository.Event, valid bool) *httptest.ResponseRecorder {
			raw, err := json.Marshal(e)
			require.NoError(t, err)
			req := httptest.NewRequest("POST", "/api/v2/payments/webhook", bytes.NewReader(raw))
			if valid {
				m := hmac.New(sha256.New, []byte("sprint-stub-signing-secret"))
				_, _ = m.Write(raw)
				req.Header.Set("X-Payment-Signature", hex.EncodeToString(m.Sum(nil)))
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			return w
		}
		require.Equal(t, 401, webhook(event, false).Code)
		bad := event
		bad.AmountMinor = 1
		require.Equal(t, 400, webhook(bad, true).Code)
		w := webhook(event, true)
		require.Equal(t, 200, w.Code, w.Body.String())
		require.Equal(t, 200, webhook(event, true).Code)
		event.EventID = "second-success-event"
		require.Equal(t, 200, webhook(event, true).Code)
		w = check("GET", "/orders/"+order.ID, 101, nil, "", 200)
		require.Contains(t, w.Body.String(), `"status":"paid"`)
		check("POST", "/orders", 101, input, "already-owned", 409)
		var gross, commission, author int64
		require.NoError(t, pool.QueryRow(ctx, `SELECT SUM(gross_minor),SUM(commission_minor),SUM(author_minor) FROM sportapp.sales_ledger WHERE order_id=$1`, order.ID).Scan(&gross, &commission, &author))
		require.Equal(t, int64(100000), gross)
		require.Equal(t, int64(20000), commission)
		require.Equal(t, int64(80000), author)
		payout := commerce_postgres_repository.PayoutInput{TrainerID: 201, AmountMinor: 30000, Currency: "RUB"}
		check("POST", "/admin/payouts", 201, payout, "payout-key", 403)
		w = check("POST", "/admin/payouts", 301, payout, "payout-key", 201)
		var p commerce_postgres_repository.Payout
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &p))
		require.Equal(t, "paid", p.Status)
		repeat := check("POST", "/admin/payouts", 301, payout, "payout-key", 201)
		require.Contains(t, repeat.Body.String(), p.ID)
		check("PATCH", "/admin/payouts/"+p.ID, 301, map[string]string{"action": "retry", "reason": "duplicate"}, "", 409)
		check("PATCH", "/admin/payouts/"+p.ID, 301, map[string]string{"action": "reconcile", "reason": "test check"}, "", 200)
		oldPayout := uuid.NewString()
		_, err = pool.Exec(ctx, `INSERT INTO sportapp.payouts(id,trainer_id,amount_minor,currency,status,created_at) VALUES($1,201,100,'RUB','failed',NOW()-INTERVAL '1 year')`, oldPayout)
		require.NoError(t, err)
		check("PATCH", "/admin/payouts/"+oldPayout, 301, map[string]string{"action": "retry", "reason": "settlement period fixture"}, "", 200)
		report := check("GET", "/trainer/analytics?from="+url.QueryEscape(time.Now().Add(-time.Hour).Format(time.RFC3339))+"&to="+url.QueryEscape(time.Now().Add(time.Minute).Format(time.RFC3339)), 201, nil, "", 200)
		var settled map[string]any
		require.NoError(t, json.Unmarshal(report.Body.Bytes(), &settled))
		require.Equal(t, float64(30100), settled["paid_out_minor"])
		event.EventID = "refund-event"
		event.Type = "refund_succeeded"
		require.Equal(t, 200, webhook(event, true).Code)
		require.Equal(t, 200, webhook(event, true).Code)
		var access string
		require.NoError(t, pool.QueryRow(ctx, `SELECT status FROM sportapp.entitlements WHERE order_id=$1`, order.ID).Scan(&access))
		require.Equal(t, "revoked", access)
		require.NoError(t, pool.QueryRow(ctx, `SELECT SUM(gross_minor),SUM(commission_minor),SUM(author_minor) FROM sportapp.sales_ledger WHERE order_id=$1`, order.ID).Scan(&gross, &commission, &author))
		require.Zero(t, gross)
		require.Zero(t, commission)
		require.Zero(t, author)
		check("POST", "/admin/payouts", 301, payout, "insufficient", 422)
		failedID := uuid.NewString()
		_, err = pool.Exec(ctx, `INSERT INTO sportapp.payouts(id,trainer_id,amount_minor,currency,status) VALUES($1,201,100000,'RUB','failed')`, failedID)
		require.NoError(t, err)
		check("PATCH", "/admin/payouts/"+failedID, 301, map[string]string{"action": "retry", "reason": "insufficient"}, "", 422)
		check("PATCH", "/admin/payouts/"+failedID, 301, map[string]string{"action": "cancel", "reason": "cancel fixture"}, "", 200)
		var count int
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.payouts WHERE trainer_id=201 AND status='paid'`).Scan(&count))
		require.Equal(t, 2, count)
		check("POST", "/payments/sandbox/"+order.ID, 101, map[string]string{"action": "pay"}, "", 409)
		w = check("GET", "/admin/settings/commission", 301, nil, "", 200)
		var c commerce_postgres_repository.Commission
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &c))
		c.Bps = 2500
		c.EffectiveAt = time.Now().UTC()
		check("PATCH", "/admin/settings/commission", 101, c, "", 403)
		check("PATCH", "/admin/settings/commission", 301, c, "", 200)
		check("PATCH", "/admin/settings/commission", 301, c, "", 409)
		apply("000018_*.down.sql")
		apply("000017_*.down.sql")
		apply("000013_*.down.sql")
		apply("000013_*.up.sql")
		apply("000017_*.up.sql")
		apply("000018_*.up.sql")
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.operation_results WHERE operation IN ('order','payout')`).Scan(&count))
		require.Zero(t, count)
		w = check("POST", "/orders", 101, input, "same-order", 201)
		var recreated commerce_postgres_repository.Order
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &recreated))
		require.NotEqual(t, order.ID, recreated.ID)
		w = check("POST", "/payments/sandbox/"+recreated.ID, 101, map[string]string{"action": "fail"}, "", 200)
		require.Contains(t, w.Body.String(), `"status":"failed"`)
		check("POST", "/payments/sandbox/"+recreated.ID, 101, map[string]string{"action": "pay"}, "", 409)
	})

	t.Run("social eligibility moderation follow notification and events", func(t *testing.T) {
		check("POST", "/programs/"+program.ID+"/reviews", 101, map[string]any{"overall": 5, "text": "Early"}, "", 422)
		w := check("GET", "/programs/"+program.ID+"/review-eligibility", 101, nil, "", 200)
		require.Contains(t, w.Body.String(), `"eligible":false`)
		var reviewVersion string
		require.NoError(t, pool.QueryRow(ctx, `SELECT id::text FROM sportapp.program_versions WHERE program_id=$1 ORDER BY created_at,id LIMIT 1`, program.ID).Scan(&reviewVersion))
		_, err := pool.Exec(ctx, `UPDATE sportapp.workout_sessions SET source_program_version_id=$1 WHERE user_id=101`, reviewVersion)
		require.NoError(t, err)
		for i := 0; i < 2; i++ {
			_, err = pool.Exec(ctx, `INSERT INTO sportapp.workout_sessions(id,user_id,client_session_id,plan_id,revision_id,source_program_version_id,started_at,finished_at,request_hash,payload,result) SELECT $1,user_id,$2,plan_id,revision_id,source_program_version_id,started_at,finished_at,request_hash,payload,result FROM sportapp.workout_sessions WHERE user_id=101 ORDER BY id LIMIT 1`, uuid.NewString(), uuid.NewString())
			require.NoError(t, err)
		}
		w = check("GET", "/programs/"+program.ID+"/review-eligibility", 101, nil, "", 200)
		require.Contains(t, w.Body.String(), `"eligible":true`)
		check("POST", "/programs/"+program.ID+"/reviews", 101, map[string]any{"overall": 6}, "", 400)
		w = check("POST", "/programs/"+program.ID+"/reviews", 101, map[string]any{"overall": 5, "structure": 4, "text": "Fixture review"}, "", 201)
		var review social_postgres_repository.Review
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &review))
		require.Equal(t, reviewVersion, review.ProgramVersionID)
		check("POST", "/programs/"+program.ID+"/reviews", 101, map[string]any{"overall": 4}, "", 409)
		check("PATCH", "/reviews/"+review.ID, 102, map[string]any{"overall": 1}, "", 404)
		check("PATCH", "/admin/reviews/"+review.ID+"/moderation", 101, map[string]string{"status": "hidden", "reason": "self"}, "", 403)
		check("PATCH", "/admin/reviews/"+review.ID+"/moderation", 301, map[string]string{"status": "hidden", "reason": "fixture moderation"}, "", 200)
		w = check("GET", "/programs/"+program.ID+"/reviews", 0, nil, "", 200)
		require.NotContains(t, w.Body.String(), "Fixture review")
		require.Contains(t, w.Body.String(), `"count":0`)
		w = check("PATCH", "/reviews/"+review.ID, 101, map[string]any{"overall": 4, "text": "Revised review"}, "", 200)
		require.Contains(t, w.Body.String(), `"status":"in_review"`)
		check("PATCH", "/admin/reviews/"+review.ID+"/moderation", 301, map[string]string{"status": "published", "reason": "revision checked"}, "", 200)
		w = check("GET", "/programs/"+program.ID+"/reviews", 0, nil, "", 200)
		require.Contains(t, w.Body.String(), `"average":4`)
		w = check("POST", "/reports", 102, map[string]string{"target_type": "review", "target_id": review.ID, "reason": "Fixture complaint"}, "", 201)
		var report social_postgres_repository.Report
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &report))
		check("GET", "/admin/reports", 102, nil, "", 403)
		w = check("GET", "/admin/reports", 301, nil, "", 200)
		require.Contains(t, w.Body.String(), `"target_review"`)
		require.Contains(t, w.Body.String(), "Revised review")
		check("PATCH", "/admin/reports/"+report.ID, 301, map[string]string{"status": "resolved", "reason": "Checked fixture"}, "", 200)
		check("PATCH", "/admin/reports/"+report.ID, 301, map[string]string{"status": "dismissed", "reason": "Already closed"}, "", 409)
		check("PUT", "/trainers/201/follow", 103, nil, "", 403)
		check("PUT", "/trainers/201/follow", 101, nil, "", 200)
		check("PUT", "/trainers/201/follow", 101, nil, "", 200)
		w = check("GET", "/me/trainers", 101, nil, "", 200)
		require.Contains(t, w.Body.String(), "Fixture trainer")
		w = check("GET", "/me/trainers", 102, nil, "", 200)
		require.NotContains(t, w.Body.String(), "Fixture trainer")
		_, err = pool.Exec(ctx, `DELETE FROM sportapp.user_roles WHERE user_id=201 AND role='trainer'`)
		require.NoError(t, err)
		w = check("GET", "/me/trainers", 101, nil, "", 200)
		require.NotContains(t, w.Body.String(), "Fixture trainer")
		_, err = pool.Exec(ctx, `INSERT INTO sportapp.user_roles(user_id,role) VALUES(201,'trainer')`)
		require.NoError(t, err)

		version := check("POST", "/trainer/programs/"+program.ID+"/versions", 201, nil, "", 201)
		var draft publications_postgres_repository.Program
		require.NoError(t, json.Unmarshal(version.Body.Bytes(), &draft))
		check("POST", "/trainer/programs/"+program.ID+"/publish", 201, map[string]int64{"version": draft.Version}, "", 202)
		check("PATCH", "/admin/programs/"+program.ID+"/moderation", 301, map[string]string{"status": "approved", "reason": "notify followers"}, "", 200)
		w = check("GET", "/notifications", 101, nil, "", 200)
		var notifications social_postgres_repository.NotificationPage
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &notifications))
		require.NotEmpty(t, notifications.Items)
		id := notifications.Items[0].ID
		check("PATCH", "/notifications/"+id, 102, map[string]bool{"read": true}, "", 404)
		check("PATCH", "/notifications/"+id, 101, map[string]bool{"read": true}, "", 200)
		check("PATCH", "/notifications/"+id, 101, map[string]bool{"read": true}, "", 200)
		check("DELETE", "/trainers/201/follow", 101, nil, "", 204)
		check("DELETE", "/trainers/201/follow", 101, nil, "", 204)
		event := social_postgres_repository.Event{ID: uuid.NewString(), Type: "view", ProgramID: program.ID, OccurredAt: time.Now().UTC(), Context: map[string]string{"surface": "catalog"}}
		body := map[string]any{"events": []social_postgres_repository.Event{event}}
		check("POST", "/events", 101, body, "", 202)
		w = check("POST", "/events", 101, body, "", 202)
		require.Contains(t, w.Body.String(), `"duplicate_count":1`)
		event.Type = "purchase"
		check("POST", "/events", 101, map[string]any{"events": []social_postgres_repository.Event{event}}, "", 400)
		check("DELETE", "/reviews/"+review.ID, 102, nil, "", 404)
		check("DELETE", "/reviews/"+review.ID, 101, nil, "", 204)
		w = check("POST", "/programs/"+program.ID+"/reviews", 101, map[string]any{"overall": 4}, "", 201)
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &review))
		require.Equal(t, reviewVersion, review.ProgramVersionID)
		check("DELETE", "/reviews/"+review.ID, 101, nil, "", 204)

		_, err = pool.Exec(ctx, `DELETE FROM sportapp.workout_sessions WHERE user_id=101 AND client_session_id<>$1`, workout.ClientSessionID)
		require.NoError(t, err)
	})

	t.Run("fork provenance snapshots revocation and paid license", func(t *testing.T) {
		var version string
		require.NoError(t, pool.QueryRow(ctx, `SELECT published_version_id::text FROM sportapp.programs WHERE id=$1`, program.ID).Scan(&version))
		in := sharing_postgres_repository.ForkInput{SourceVersionID: version, Title: "Fork fixture"}
		check("POST", "/programs/"+program.ID+"/fork", 102, in, "forbidden-fork", 403)
		w := check("POST", "/programs/"+program.ID+"/fork", 101, in, "fork-once", 201)
		var fork sharing_postgres_repository.ForkResult
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &fork))
		require.Equal(t, &version, fork.SourceVersionID)
		repeat := check("POST", "/programs/"+program.ID+"/fork", 101, in, "fork-once", 201)
		require.Contains(t, repeat.Body.String(), fork.PlanID)
		in.Title = "Different"
		check("POST", "/programs/"+program.ID+"/fork", 101, in, "fork-once", 409)
		copied, err := planSvc.GetPlan(ctx, "101", fork.PlanID)
		require.NoError(t, err)
		require.Equal(t, &version, copied.SourceProgramVersionID)
		var sourceDay string
		require.NoError(t, pool.QueryRow(ctx, `SELECT id::text FROM sportapp.program_days WHERE program_version_id=$1 ORDER BY position LIMIT 1`, version).Scan(&sourceDay))
		require.NotEqual(t, sourceDay, copied.Days[0].ID)
		create := plans_postgres_repository.Input{Title: "Shared fixture", Days: []plans_postgres_repository.Day{{ID: uuid.NewString(), Title: "First day", Position: 0, Exercises: []plans_postgres_repository.Prescription{{ExerciseID: 1, Position: 0, Sets: 3, RepsMin: 8, RepsMax: 12, RestSec: 90}}}, {ID: uuid.NewString(), Title: "Second day", Position: 1, Exercises: []plans_postgres_repository.Prescription{}}}, Schedule: plans_postgres_repository.Schedule{TimeZone: "Europe/Moscow", StartDate: "2026-11-01", Entries: []plans_postgres_repository.ScheduleEntry{}}}
		own, err := planSvc.CreatePlan(ctx, "101", create)
		require.NoError(t, err)
		check("POST", "/shares", 102, map[string]any{"target_type": "plan", "target_id": own.ID, "expires_at": nil}, "", 404)
		w = check("POST", "/shares", 101, map[string]any{"target_type": "plan", "target_id": own.ID, "expires_at": nil}, "", 201)
		var share sharing_postgres_repository.ShareResult
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &share))
		u, err := url.Parse(share.URL)
		require.NoError(t, err)
		token := strings.TrimPrefix(u.Path, "/share/")
		require.Len(t, token, 43)
		w = check("GET", "/shares/"+token, 0, nil, "", 200)
		require.Contains(t, w.Body.String(), "First day")
		require.NotContains(t, w.Body.String(), "Europe/Moscow")
		require.NotContains(t, w.Body.String(), "profile")
		var public sharing_postgres_repository.PublicShare
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &public))
		require.Len(t, public.Days, 2)
		guestShare := check("GET", "/shares/"+token, 103, nil, "", 200)
		require.Contains(t, guestShare.Body.String(), `"can_fork":false`)

		change := plans_postgres_repository.Input{Version: own.Version, Title: "Changed source", Days: create.Days, Schedule: create.Schedule}
		change.Days[0].Title = "Changed first day"
		_, err = planSvc.UpdatePlan(ctx, "101", own.ID, change)
		require.NoError(t, err)
		w = check("GET", "/shares/"+token, 0, nil, "", 200)
		require.Contains(t, w.Body.String(), "First day")
		require.NotContains(t, w.Body.String(), "Changed first day")
		saved := check("POST", "/shares/"+token+"/save", 102, nil, "", 201)
		repeated := check("POST", "/shares/"+token+"/save", 102, nil, "", 201)
		require.Equal(t, saved.Body.String(), repeated.Body.String())
		check("POST", "/shares/"+token+"/fork", 102, nil, "shared-fork", 201)
		var savedItem map[string]string
		require.NoError(t, json.Unmarshal(saved.Body.Bytes(), &savedItem))
		itemID := savedItem["library_item_id"]
		check("GET", "/me/saved-shares/"+itemID, 101, nil, "", 404)
		check("GET", "/me/saved-shares/"+itemID, 102, nil, "", 200)
		w = check("GET", "/me/saved-shares", 102, nil, "", 200)
		require.Contains(t, w.Body.String(), itemID)
		check("DELETE", "/shares/"+share.ID, 102, nil, "", 404)
		check("DELETE", "/shares/"+share.ID, 101, nil, "", 204)
		check("GET", "/me/saved-shares/"+itemID, 102, nil, "", 404)
		check("POST", "/me/saved-shares/"+itemID+"/fork", 102, nil, "new-revoked-library-fork", 404)
		check("DELETE", "/me/saved-shares/"+itemID, 101, nil, "", 404)
		check("DELETE", "/me/saved-shares/"+itemID, 102, nil, "", 204)
		check("GET", "/shares/"+token, 0, nil, "", 404)
		check("POST", "/shares/"+token+"/fork", 102, nil, "shared-fork", 201)
		check("POST", "/shares/"+token+"/fork", 102, nil, "new-after-revoke", 404)
		dayShare := check("POST", "/shares", 101, map[string]string{"target_type": "plan_day", "target_id": own.Days[1].ID}, "", 201)
		require.NoError(t, json.Unmarshal(dayShare.Body.Bytes(), &share))
		u, err = url.Parse(share.URL)
		require.NoError(t, err)
		token = strings.TrimPrefix(u.Path, "/share/")
		w = check("GET", "/shares/"+token, 0, nil, "", 200)
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &public))
		require.Len(t, public.Days, 1)
		require.Equal(t, "Second day", public.Days[0].Title)
		_, err = pool.Exec(ctx, `UPDATE sportapp.public_shares SET expires_at=NOW()-INTERVAL '1 second' WHERE id=$1`, share.ID)
		require.NoError(t, err)
		check("GET", "/shares/"+token, 0, nil, "", 404)
		paid, paidVersion, paidDay := uuid.NewString(), uuid.NewString(), uuid.NewString()
		_, err = pool.Exec(ctx, `INSERT INTO sportapp.programs(id,trainer_id,status) VALUES($1,201,'active')`, paid)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `INSERT INTO sportapp.program_versions(id,program_id,number,title,description,goal,level,duration_weeks,sessions_per_week,price_minor,currency,status,allow_public_remix) VALUES($1,$2,1,'Licensed fixture','Private structure','muscle_gain','beginner',4,1,10000,'RUB','published',FALSE)`, paidVersion, paid)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `UPDATE sportapp.programs SET published_version_id=$1 WHERE id=$2`, paidVersion, paid)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `INSERT INTO sportapp.program_days(id,program_version_id,title,position) VALUES($1,$2,'Licensed day',0)`, paidDay, paidVersion)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `INSERT INTO sportapp.program_exercises(id,program_day_id,exercise_id,position,sets,reps_min,reps_max,rest_sec) VALUES($1,$2,1,0,3,8,12,90)`, uuid.NewString(), paidDay)
		require.NoError(t, err)
		w = check("POST", "/orders", 101, commerce_postgres_repository.OrderInput{ProgramID: paid, ProgramVersionID: paidVersion}, "license-order", 201)
		var order commerce_postgres_repository.Order
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &order))
		check("POST", "/payments/sandbox/"+order.ID, 101, map[string]string{"action": "pay"}, "", 200)
		w = check("POST", "/programs/"+paid+"/fork", 101, sharing_postgres_repository.ForkInput{SourceVersionID: paidVersion, Title: "Private licensed fork"}, "licensed-fork", 201)
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &fork))
		w = check("POST", "/shares", 101, map[string]string{"target_type": "plan", "target_id": fork.PlanID}, "", 201)
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &share))
		require.Equal(t, "preview", share.Visibility)
		u, err = url.Parse(share.URL)
		require.NoError(t, err)
		token = strings.TrimPrefix(u.Path, "/share/")
		for _, actor := range []int{0, 101, 102} {
			w = check("GET", "/shares/"+token, actor, nil, "", 200)
			require.NotContains(t, w.Body.String(), "Licensed day")
			require.Contains(t, w.Body.String(), `"requires_access":true`)
		}
		check("POST", "/shares/"+token+"/fork", 101, nil, "denied-licensed-share", 403)
		check("POST", "/shares/"+token+"/save", 102, nil, "", 201)
		w = check("GET", "/catalog/programs/"+paid, 102, nil, "", 200)
		require.Contains(t, w.Body.String(), `"access":"preview"`)
		check("POST", "/payments/sandbox/"+order.ID, 101, map[string]string{"action": "refund"}, "", 200)
		dispatcher := commerce_postgres_repository.NewRepository(pool)
		require.NoError(t, dispatcher.DispatchNotifications(ctx))
		require.NoError(t, dispatcher.DispatchNotifications(ctx))
		var notifications int
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.notifications WHERE user_id=101 AND event_key LIKE $1`, "order:"+order.ID+":%").Scan(&notifications))
		require.Equal(t, 2, notifications)

		check("POST", "/shares", 101, map[string]string{"target_type": "plan", "target_id": fork.PlanID}, "", 403)
		check("GET", "/me/plans/"+fork.PlanID, 101, nil, "", 200)
	})

	t.Run("manual creation retries preserve identity and reject changed payload", func(t *testing.T) {
		in := plans_postgres_repository.Input{Title: "Idempotent manual fixture", Days: []plans_postgres_repository.Day{{Title: "Server identity", Position: 0, Exercises: []plans_postgres_repository.Prescription{}}}, Schedule: plans_postgres_repository.Schedule{TimeZone: "Europe/Moscow", StartDate: "2026-11-14", Entries: []plans_postgres_repository.ScheduleEntry{}}}
		responses := make(chan *httptest.ResponseRecorder, 2)
		for i := 0; i < 2; i++ {
			go func() { responses <- call("POST", "/me/plans", 102, in, "manual-retry") }()
		}
		first, second := <-responses, <-responses
		require.Equal(t, 201, first.Code, first.Body.String())
		require.Equal(t, 201, second.Code, second.Body.String())
		require.JSONEq(t, first.Body.String(), second.Body.String())
		var plan plans_postgres_repository.Plan
		require.NoError(t, json.Unmarshal(first.Body.Bytes(), &plan))
		require.NotEmpty(t, plan.Days[0].ID)
		in.Title = "Changed request"
		check("POST", "/me/plans", 102, in, "manual-retry", 409)
	})

	t.Run("reference nullable fields reason and analytics permissions", func(t *testing.T) {
		input := map[string]any{"name": "Reference fixture", "muscle_group": "Ноги", "muscle_subgroup": "Квадрицепс", "equipment_ids": []string{}, "image_url": "https://example.test/image", "video_url": "https://example.test/video", "reason": "fixture create"}
		check("POST", "/admin/exercises", 101, input, "", 403)
		w := check("POST", "/admin/exercises", 301, input, "", 201)
		var exercise map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &exercise))
		id := fmt.Sprint(exercise["id"])
		delete(input, "reason")
		check("PATCH", "/admin/exercises/"+id, 301, input, "", 400)
		input["reason"] = "fixture nullable clear"
		input["image_url"] = nil
		input["video_url"] = nil
		input["muscle_subgroup"] = nil
		w = check("PATCH", "/admin/exercises/"+id, 301, input, "", 200)
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &exercise))
		require.Nil(t, exercise["image_url"])
		require.Nil(t, exercise["video_url"])
		require.Nil(t, exercise["muscle_subgroup"])
		input["image_url"] = "https://example.test/new"
		check("PATCH", "/admin/exercises/"+id, 301, input, "", 200)
		delete(input, "image_url")
		w = check("PATCH", "/admin/exercises/"+id, 301, input, "", 200)
		require.Contains(t, w.Body.String(), "https://example.test/new")
		var reason string
		require.NoError(t, pool.QueryRow(ctx, `SELECT reason FROM sportapp.publication_audit WHERE action=$1 ORDER BY occurred_at DESC LIMIT 1`, "exercises:"+id).Scan(&reason))
		require.Equal(t, "fixture nullable clear", reason)
		category := map[string]any{"name": "Fixture category", "slug": "fixture-category", "reason": "fixture category"}
		w = check("POST", "/admin/categories", 301, category, "", 201)
		var c map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &c))
		category["is_active"] = false
		check("PATCH", "/admin/categories/"+c["id"].(string), 301, category, "", 200)
		w = check("GET", "/catalog/categories", 0, nil, "", 200)
		require.NotContains(t, w.Body.String(), "Fixture category")
		check("GET", "/admin/analytics", 101, nil, "", 403)
		check("GET", "/trainer/analytics", 101, nil, "", 403)
		check("GET", "/admin/analytics?from=wrong", 301, nil, "", 400)
		w = check("GET", "/trainer/analytics", 201, nil, "", 200)
		var analytics map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &analytics))
		require.Equal(t, float64(10000), analytics["gross_minor"])
		require.Equal(t, float64(10000), analytics["refund_minor"])
		require.Equal(t, float64(0), analytics["author_earned_minor"])
		w = check("GET", "/trainer/analytics", 202, nil, "", 200)
		require.NotContains(t, w.Body.String(), "Licensed fixture")
		w = check("GET", "/admin/analytics", 301, nil, "", 200)
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &analytics))
		require.Equal(t, float64(0), analytics["net_minor"])
		w = check("GET", "/recommendations/programs", 101, nil, "", 200)
		require.Contains(t, w.Body.String(), `"fallback":true`)
		require.Contains(t, w.Body.String(), `"model_version":"rules-v1"`)
		check("GET", "/recommendations/programs?limit=101", 101, nil, "", 400)
		check("GET", "/dashboard", 101, nil, "", 200)
		check("GET", "/dashboard", 103, nil, "", 200)
	})

	t.Run("dashboard isolates failed progress and followed blocks", func(t *testing.T) {
		for _, fixture := range []struct{ table, block string }{{"workout_sessions", "recent_progress"}, {"trainer_follows", "followed_trainers"}} {
			t.Run(fixture.block, func(t *testing.T) {
				_, err := pool.Exec(ctx, `ALTER TABLE sportapp.`+fixture.table+` RENAME TO dashboard_unavailable_fixture`)
				require.NoError(t, err)
				defer func() {
					_, err := pool.Exec(ctx, `ALTER TABLE sportapp.dashboard_unavailable_fixture RENAME TO `+fixture.table)
					require.NoError(t, err)
				}()
				w := check("GET", "/dashboard", 101, nil, "", 200)
				var dashboard insights_service.Dashboard
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &dashboard))
				statuses := map[string]string{}
				for _, block := range dashboard.Blocks {
					statuses[block.ID] = block.Status
				}
				require.Equal(t, "unavailable", statuses[fixture.block])
				require.Equal(t, "ready", statuses["ai_message"])
				require.NotEqual(t, "unavailable", statuses["today_workout"])
			})
		}
	})

	t.Run("startup seeds preserve administrative edits", func(t *testing.T) {
		t.Chdir("../../..")
		var before int
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.exercises`).Scan(&before))
		input := map[string]any{"name": "Административное имя seed упражнения", "equipment_ids": []string{}, "image_url": "/static/images/admin-override.webp", "video_url": nil, "reason": "seed persistence fixture"}
		check("PATCH", "/admin/exercises/1", 301, input, "", 200)
		users := users_postgres_repository.NewUsersRepository(pool)
		require.NoError(t, users.EnsureExercisesSeeded(ctx))
		require.NoError(t, catalogRepo.EnsureReferenceData(ctx))
		require.NoError(t, users.EnsureExercisesSeeded(ctx))
		require.NoError(t, catalogRepo.EnsureReferenceData(ctx))
		var name, image, key string
		var video *string
		var after, equipment int
		require.NoError(t, pool.QueryRow(ctx, `SELECT exercises_name,image_url,video_url,seed_key FROM sportapp.exercises WHERE id=1`).Scan(&name, &image, &video, &key))
		require.Equal(t, input["name"], name)
		require.Equal(t, input["image_url"], image)
		require.Nil(t, video)
		require.Equal(t, "builtin:v1:1", key)
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.exercises`).Scan(&after))
		require.Equal(t, before, after)
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.exercise_equipment WHERE exercise_id=1`).Scan(&equipment))
		require.Zero(t, equipment)
		input["name"] = "Жим ногами лёжа"
		input["equipment_ids"] = []string{"machine"}
		check("PATCH", "/admin/exercises/1", 301, input, "", 200)
	})

	t.Run("reference seed serializes administrative equipment clearing", func(t *testing.T) {
		_, err := pool.Exec(ctx, `UPDATE sportapp.exercises SET admin_modified=FALSE WHERE id=1; INSERT INTO sportapp.exercise_equipment(exercise_id,equipment_id) VALUES(1,'machine') ON CONFLICT DO NOTHING`)
		require.NoError(t, err)
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		defer func() { _ = tx.Rollback(context.Background()) }()
		_, err = tx.Exec(ctx, `UPDATE sportapp.exercises SET admin_modified=TRUE WHERE id=1;DELETE FROM sportapp.exercise_equipment WHERE exercise_id=1`)
		require.NoError(t, err)
		finished := make(chan error, 1)
		go func() { finished <- catalogRepo.EnsureReferenceData(ctx) }()
		require.Eventually(t, func() bool {
			var waiting bool
			err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE datname=$1 AND wait_event_type='Lock' AND query LIKE '%jsonb_array_elements%')`, database).Scan(&waiting)
			return err == nil && waiting
		}, 2*time.Second, 10*time.Millisecond)
		require.NoError(t, tx.Commit(ctx))
		require.NoError(t, <-finished)
		var count int
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.exercise_equipment WHERE exercise_id=1`).Scan(&count))
		require.Zero(t, count)
		_, err = pool.Exec(ctx, `INSERT INTO sportapp.exercise_equipment(exercise_id,equipment_id) VALUES(1,'machine') ON CONFLICT DO NOTHING`)
		require.NoError(t, err)
	})

	t.Run("browser journeys against real Go API", func(t *testing.T) {
		if os.Getenv("SPORTIK_BROWSER_TESTS") != "1" && os.Getenv("SPORTIK_POSTMAN_TESTS") != "1" {
			t.Skip("set SPORTIK_BROWSER_TESTS=1 to run desktop and mobile Playwright against this isolated database")
		}
		v1 := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion1)
		v1.RegisterRoutes(users_http.NewUsersHTTPHandler(users_service.NewUsersService(users_postgres_repository.NewUsersRepository(pool), mlclient.NewClient(mlclient.Config{BaseURL: ml.URL}), jwt), jwt).Routes()...)
		mux := http.NewServeMux()
		mux.Handle("/api/v2/", handler)
		mux.Handle("/api/v1/", core_http_middleware.AccountStatus(jwt, pool)(http.StripPrefix("/api/v1", v1)))
		server := httptest.NewServer(mux)
		defer server.Close()
		fixtureFile, err := filepath.Abs("../../../.validation/browser-fixture.json")
		require.NoError(t, err)
		raw, err := json.Marshal(map[string]any{"token": tokens[102], "admin_token": tokens[301], "base_url": server.URL})
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(fixtureFile, raw, 0600))
		defer func() { _ = os.Remove(fixtureFile) }()
		if os.Getenv("SPORTIK_BROWSER_TESTS") == "1" {
			logFile, err := os.Create("../../../.validation/real-e2e.log")
			require.NoError(t, err)
			defer func() { _ = logFile.Close() }()
			browserCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
			defer cancel()
			command := exec.CommandContext(browserCtx, "npm", "--prefix", "../../../frontend", "run", "test:e2e")
			command.Env = append(os.Environ(), "SPORTIK_REAL_API=1", "SPORTIK_FIXTURE_FILE="+fixtureFile, "VITE_API_PROXY_TARGET="+server.URL, "PLAYWRIGHT_CHANNEL=chrome")
			command.Stdout = logFile
			command.Stderr = logFile
			require.NoError(t, command.Run(), "see .validation/real-e2e.log")
		}
		if os.Getenv("SPORTIK_POSTMAN_TESTS") == "1" {
			mlMode.Store(0)
			modules, err := filepath.Abs("../../../.validation/postman-tools/node_modules")
			require.NoError(t, err)
			report, err := filepath.Abs("../../../.validation/postman-real-report.json")
			require.NoError(t, err)
			logFile, err := os.Create("../../../.validation/postman-real.log")
			require.NoError(t, err)
			defer func() { _ = logFile.Close() }()
			runCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
			defer cancel()
			command := exec.CommandContext(runCtx, "node", "../../../docs/postman/run_isolated.cjs", modules, fixtureFile, report)
			command.Stdout = logFile
			command.Stderr = logFile
			require.NoError(t, command.Run(), "see .validation/postman-real-report.json")
		}

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
