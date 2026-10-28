package profile_postgres_repository_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	core_auth "sport_app/internal/core/auth"
	core_errors "sport_app/internal/core/errors"
	core_pgx_pool "sport_app/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_server "sport_app/internal/core/transport/http/server"
	catalog_postgres_repository "sport_app/internal/features/catalog/repository/postgres"
	profile_postgres_repository "sport_app/internal/features/profile/repository/postgres"
	profile_service "sport_app/internal/features/profile/service"
	profile_transport_http "sport_app/internal/features/profile/transport/http"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProfileIntegration применяет миграции и фикстуры только к новой одноразовой БД на выделенном порту.
// Рабочая .env не используется, существующая схема никогда не очищается.
func TestProfileIntegration(t *testing.T) {
	database := os.Getenv("SPORTIK_PROFILE_TEST_DB")
	if database == "" {
		t.Skip("set SPORTIK_PROFILE_TEST_DB to a new disposable sportik_profile_test_* database on 127.0.0.1:55443")
	}
	require.True(t, strings.HasPrefix(database, "sportik_profile_test_"), "only a dedicated test database is allowed")
	ctx := context.Background()
	config := core_pgx_pool.Config{Host: "127.0.0.1", Port: "55443", User: "sportik_test", Database: database, Timeout: 10 * time.Second}
	pool, err := core_pgx_pool.NewPool(ctx, config)
	require.NoError(t, err)
	defer pool.Close()
	var empty bool
	require.NoError(t, pool.QueryRow(ctx, `SELECT to_regnamespace('sportapp') IS NULL`).Scan(&empty))
	require.True(t, empty, "test database must be new; existing data is never cleared")
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
	// Verify rollback on an empty schema, then migrate a populated v1 database.
	apply("000008_profile_v2.down.sql")
	_, err = pool.Exec(ctx, `INSERT INTO sportapp.users (id,is_anonymous,created_at) VALUES
		(101,FALSE,NOW()),(102,TRUE,NOW()),(103,FALSE,NOW()),(104,FALSE,NOW());
		INSERT INTO sportapp.profile(user_id,age,weight_kg,fitness_level,training_days_map,injuries_notes,created_at)
		VALUES (101,26,75,'Новичок',ARRAY['mon','wed'],TRUE,NOW()),
		(102,30,80,'Любитель',ARRAY['Вт'],FALSE,NOW()),(103,20,90,NULL,NULL,NULL,NOW());
		INSERT INTO sportapp.profile(user_id,age,height_cm,weight_kg,gender,activity_level,goal,fitness_level,created_at)
		VALUES (104,0,-1,0,'',' ','','legacy-label',NOW());
		INSERT INTO sportapp.profile_equipment (user_id,equipment_id) VALUES (101,'dumbbell');
		INSERT INTO sportapp.notes(user_id,title,body) VALUES (101,'fixture','must survive');`)
	require.NoError(t, err)
	users := users_postgres_repository.NewUsersRepository(pool)
	require.NoError(t, users.EnsureExercisesSeeded(ctx))
	require.NoError(t, catalog_postgres_repository.NewRepository(pool).EnsureReferenceData(ctx))
	apply("000008_profile_v2.up.sql")
	repo := profile_postgres_repository.NewRepository(pool)
	svc := profile_service.NewService(repo)
	initial, err := svc.GetProfile(ctx, "101")
	require.NoError(t, err)
	assert.Equal(t, 75.0, *initial.WeightKg)
	assert.Equal(t, []int{1, 3}, initial.TrainingDays)
	assert.Equal(t, "beginner", *initial.Level)
	assert.Equal(t, []string{"dumbbell"}, initial.EquipmentIDs)
	assert.Empty(t, initial.ExcludedExerciseIDs, "injuries_notes must not infer exclusions")
	assert.True(t, *initial.InjuriesNotes)
	legacyIncomplete, err := svc.GetProfile(ctx, "104")
	require.NoError(t, err)
	assert.Zero(t, *legacyIncomplete.Age)
	assert.Equal(t, -1, *legacyIncomplete.HeightCm)
	assert.Zero(t, *legacyIncomplete.WeightKg)
	assert.Empty(t, *legacyIncomplete.Gender)
	assert.Empty(t, *legacyIncomplete.Goal)
	assert.Equal(t, " ", *legacyIncomplete.ActivityLevel)
	assert.Nil(t, legacyIncomplete.Level, "unknown legacy labels must not be guessed or overwritten")
	jwt := core_auth.NewJWT(core_auth.Config{Secret: "profile-integration-test-only", TokenTTL: time.Hour})
	router := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion2)
	router.RegisterRoutes(profile_transport_http.NewProfileHTTPHandler(svc, jwt).Routes()...)
	handler := core_http_middleware.RequestID()(router)
	request := func(method, body, user string, status int) profile_postgres_repository.Profile {
		t.Helper()
		r := httptest.NewRequest(method, "/me/profile", strings.NewReader(body))
		token, err := jwt.CreateToken(user, user == "102")
		require.NoError(t, err)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		require.Equal(t, status, w.Code, "%s", w.Body.String())
		var p profile_postgres_repository.Profile
		if status == 200 {
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &p))
		}
		return p
	}
	p := request("PATCH", `{"version":1,"weight_kg":75.25,"training_days":[7,1],"equipment_ids":["bench","dumbbell"],"excluded_exercise_ids":[1],"session_duration_min":45,"level":"intermediate"}`, "101", 200)
	assert.Equal(t, int64(2), p.Version)
	assert.Equal(t, 75.25, *p.WeightKg)
	assert.Equal(t, []int{1, 7}, p.TrainingDays)
	assert.Equal(t, []int{1}, p.ExcludedExerciseIDs)
	assert.Equal(t, 45, *p.SessionDurationMin)
	assert.Equal(t, 26, *p.Age, "omitted fields are preserved")
	assert.True(t, *p.InjuriesNotes)
	assert.Equal(t, "intermediate", *p.Level)
	for _, body := range []string{
		`{"version":2,"age":99,"equipment_ids":["missing"]}`,
		`{"version":2,"age":99,"excluded_exercise_ids":[999999]}`,
	} {
		request("PATCH", body, "101", 400)
	}
	request("PATCH", `{"version":1,"weight_kg":10}`, "101", 409)
	after := request("GET", "", "101", 200)
	assert.Equal(t, p, after, "invalid references and conflicts must leave all fields unchanged")
	request("PATCH", `{"version":2,"user_id":103,"age":99}`, "101", 400)
	other := request("GET", "", "103", 200)
	assert.Equal(t, 90.0, *other.WeightKg)
	guest := request("PATCH", `{"version":1,"weight_kg":80.5}`, "102", 200)
	assert.Equal(t, 80.5, *guest.WeightKg)
	assert.Equal(t, []int{2}, guest.TrainingDays)
	request("GET", "", "999999", 404)
	_, err = pool.Exec(ctx, `UPDATE sportapp.exercises SET is_active = FALSE WHERE id = 2`)
	require.NoError(t, err)
	request("PATCH", `{"version":2,"excluded_exercise_ids":[2]}`, "101", 400)
	// Competing requests share an expected version: exactly one succeeds.
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, weight := range []float64{76.125, 77.875} {
		wg.Add(1)
		go func(weight float64) {
			defer wg.Done()
			_, err := svc.UpdateProfile(ctx, "101", profile_postgres_repository.Patch{Version: 2, WeightKg: &weight})
			results <- err
		}(weight)
	}
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else {
			require.ErrorIs(t, err, core_errors.ErrConflict)
			conflict++
		}
	}
	assert.Equal(t, 1, success)
	assert.Equal(t, 1, conflict)
	current, err := svc.GetProfile(ctx, "101")
	require.NoError(t, err)
	assert.Equal(t, int64(3), current.Version)
	assert.Contains(t, []float64{76.125, 77.875}, *current.WeightKg)
	// v1 still reads integers, legacy day codes and fitness labels.
	legacy, err := users.GetProfile(ctx, "101")
	require.NoError(t, err)
	assert.NotNil(t, legacy.WeightKg)
	assert.Equal(t, []string{"sun", "mon"}, legacy.TrainingDaysMap)
	assert.Equal(t, "Любитель", *legacy.FitnessLevel)
	_, err = pool.Exec(ctx, `SELECT 1`)
	require.NoError(t, err)
	// A recreated connection/repository reads persisted values without a cache.
	reconnected, err := core_pgx_pool.NewPool(ctx, config)
	require.NoError(t, err)
	restored, err := profile_postgres_repository.NewRepository(reconnected).GetProfile(ctx, 101)
	reconnected.Close()
	require.NoError(t, err)
	assert.Equal(t, current, restored)
	// Empty arrays explicitly clear lists, while untouched boolean restrictions stay intact.
	cleared := request("PATCH", `{"version":3,"training_days":[],"equipment_ids":[],"excluded_exercise_ids":[]}`, "101", 200)
	assert.Equal(t, int64(4), cleared.Version)
	assert.Empty(t, cleared.TrainingDays)
	assert.Empty(t, cleared.EquipmentIDs)
	assert.Empty(t, cleared.ExcludedExerciseIDs)
	assert.True(t, *cleared.InjuriesNotes)
	// The rollback guard must fail atomically instead of rounding fractional weights.
	down, err := os.ReadFile("../../../../../migrations/000008_profile_v2.down.sql")
	require.NoError(t, err)
	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	_, err = tx.Exec(ctx, string(down))
	require.Error(t, err)
	require.NoError(t, tx.Rollback(ctx))
	var columnType string
	require.NoError(t, pool.QueryRow(ctx, `SELECT data_type FROM information_schema.columns
		WHERE table_schema='sportapp' AND table_name='profile' AND column_name='weight_kg'`).Scan(&columnType))
	assert.Equal(t, "numeric", columnType)
	// Explicitly resolve synthetic fractional fixtures, then check down/up on populated data.
	_, err = pool.Exec(ctx, `UPDATE sportapp.profile SET weight_kg = trunc(weight_kg)`)
	require.NoError(t, err)
	apply("000008_profile_v2.down.sql")
	legacy, err = users.GetProfile(ctx, "101")
	require.NoError(t, err)
	assert.Equal(t, 26, *legacy.Age)
	var note string
	require.NoError(t, pool.QueryRow(ctx, `SELECT body FROM sportapp.notes WHERE user_id=101`).Scan(&note))
	assert.Equal(t, "must survive", note)
	apply("000008_profile_v2.up.sql")
	for _, id := range []int{101, 102, 103} {
		_, err := svc.GetProfile(ctx, strconv.Itoa(id))
		require.NoError(t, err, fmt.Sprint(id))
	}
}
