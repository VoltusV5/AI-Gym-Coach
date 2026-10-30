package generation_postgres_repository_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	core_auth "sport_app/internal/core/auth"
	core_pgx_pool "sport_app/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_server "sport_app/internal/core/transport/http/server"
	catalog_postgres_repository "sport_app/internal/features/catalog/repository/postgres"
	generation_postgres_repository "sport_app/internal/features/generation/repository/postgres"
	generation_service "sport_app/internal/features/generation/service"
	generation_transport_http "sport_app/internal/features/generation/transport/http"
	"sport_app/internal/features/mlclient"
	plans_postgres_repository "sport_app/internal/features/plans/repository/postgres"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
)

type interveningClient struct {
	inner  *mlclient.V2Client
	change func(mlclient.GenerateResult)
}

func (c interveningClient) Generate(ctx context.Context, input mlclient.GenerateInput) (mlclient.GenerateResult, error) {
	result, err := c.inner.Generate(ctx, input)
	if err == nil {
		c.change(result)
	}
	return result, err
}

// TestGenerationIntegration uses only a new fixture database in a separate PostgreSQL cluster.
// It verifies the real Python contract, atomic persistence, replay and catalog changes during inference.
func TestGenerationIntegration(t *testing.T) {
	database, mlURL := os.Getenv("SPORTIK_GENERATION_TEST_DB"), os.Getenv("SPORTIK_ML_TEST_URL")
	if database == "" || mlURL == "" {
		t.Skip("set SPORTIK_GENERATION_TEST_DB=sportik_generation_test_* on 127.0.0.1:55443 and SPORTIK_ML_TEST_URL to disposable services")
	}
	require.True(t, strings.HasPrefix(database, "sportik_generation_test_"))
	ctx := context.Background()
	pool, err := core_pgx_pool.NewPool(ctx, core_pgx_pool.Config{Host: "127.0.0.1", Port: "55443", User: "sportik_test", Database: database, Timeout: 10 * time.Second})
	require.NoError(t, err)
	defer pool.Close()
	var empty bool
	require.NoError(t, pool.QueryRow(ctx, `SELECT to_regnamespace('sportapp') IS NULL`).Scan(&empty))
	require.True(t, empty, "existing schema is never cleared")
	files, err := filepath.Glob("../../../../../migrations/*.up.sql")
	require.NoError(t, err)
	require.NotEmpty(t, files)
	for _, file := range files {
		sql, err := os.ReadFile(file)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, string(sql))
		require.NoError(t, err, file)
	}
	_, err = pool.Exec(ctx, `INSERT INTO sportapp.users(id,is_anonymous,created_at) VALUES(9201,FALSE,NOW());
		INSERT INTO sportapp.profile(user_id,age,height_cm,weight_kg,gender,activity_level,injuries_notes,goal,fitness_level,training_days_map,session_duration_min,created_at)
		VALUES(9201,26,175,75.5,'м','Средняя активность',FALSE,'muscle_gain','Новичок',ARRAY['mon','wed','fri'],60,NOW());
		INSERT INTO sportapp.profile_equipment(user_id,equipment_id) SELECT 9201,id FROM sportapp.equipment;`)
	require.NoError(t, err)
	require.NoError(t, users_postgres_repository.NewUsersRepository(pool).EnsureExercisesSeeded(ctx))
	require.NoError(t, catalog_postgres_repository.NewRepository(pool).EnsureReferenceData(ctx))
	_, err = pool.Exec(ctx, `INSERT INTO sportapp.exercises(exercises_name,muscular_group) VALUES('Fixture without muscle group',NULL)`)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO sportapp.exercises(exercises_name,muscular_group,muscular_subgroup)
		VALUES('Fixture empty subgroup','Пресс','')`)
	require.NoError(t, err)
	repo := generation_postgres_repository.NewRepository(pool)
	catalog, err := repo.Exercises(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, catalog)
	for _, exercise := range catalog {
		require.NotEmpty(t, exercise.MuscleGroup)
		if exercise.MuscleSubgroup != nil {
			require.NotEmpty(t, *exercise.MuscleSubgroup)
		}
	}
	// Keep a prior active plan to verify that generation never replaces it implicitly.
	oldID := uuid.NewString()
	_, err = pool.Exec(ctx, `INSERT INTO sportapp.training_plans(id,owner_id,title,source_type,is_active)
		VALUES($1,9201,'Fixture active plan','manual',TRUE)`, oldID)
	require.NoError(t, err)
	client := mlclient.NewV2Client(mlclient.Config{BaseURL: mlURL, ServiceToken: "synthetic-integration-token"})
	svc := generation_service.NewService(repo, client)
	jwt := core_auth.NewJWT(core_auth.Config{Secret: "generation-fixture-only", TokenTTL: time.Hour})
	token, err := jwt.CreateToken("9201", false)
	require.NoError(t, err)
	router := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion2)
	router.RegisterRoutes(generation_transport_http.NewHandler(svc, jwt).Routes()...)
	handler := core_http_middleware.RequestID()(router)
	request := func(bearer, key string, status int) generation_postgres_repository.Result {
		t.Helper()
		r := httptest.NewRequest("POST", "/plans/generate", strings.NewReader(`{"profile_version":1}`))
		r.Header.Set("Authorization", "Bearer "+bearer)
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Idempotency-Key", key)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		require.Equal(t, status, w.Code, w.Body.String())
		var result generation_postgres_repository.Result
		if status == 201 {
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
		}
		return result
	}
	request("invalid-token", "unauthorized", 401)
	result := request(token, "fixture-generate", 201)
	require.NotEmpty(t, result.PlanID)
	require.Equal(t, "catboost-v0-planner-v2", result.ModelVersion)
	require.Equal(t, result, request(token, "fixture-generate", 201))
	plan, err := plans_postgres_repository.NewRepository(pool).GetPlan(ctx, 9201, result.PlanID)
	require.NoError(t, err)
	require.Equal(t, "ai", plan.SourceType)
	require.False(t, plan.IsActive)
	require.Len(t, plan.Days, 3)
	for _, day := range plan.Days {
		require.NotEmpty(t, day.Exercises)
	}
	// A catalog edit after Python validation must still be rejected by Go's transaction.
	racing := generation_service.NewService(repo, interveningClient{inner: client, change: func(result mlclient.GenerateResult) {
		_, err := pool.Exec(ctx, `UPDATE sportapp.exercises SET is_active=FALSE WHERE id=$1`, result.Days[0].Exercises[0].ExerciseID)
		require.NoError(t, err)
	}})
	_, err = racing.Generate(ctx, "9201", "fixture-catalog-race", "req-catalog-race", 1)
	require.ErrorContains(t, err, "NO_FEASIBLE_PLAN")
	var plans, generations int
	var active bool
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.training_plans WHERE owner_id=9201`).Scan(&plans))
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.plan_generations WHERE user_id=9201`).Scan(&generations))
	require.NoError(t, pool.QueryRow(ctx, `SELECT is_active FROM sportapp.training_plans WHERE id=$1`, oldID).Scan(&active))
	require.Equal(t, 2, plans)
	require.Equal(t, 1, generations)
	require.True(t, active)
}
