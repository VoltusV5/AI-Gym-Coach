package catalog_postgres_repository_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	catalog_postgres_repository "sport_app/internal/features/catalog/repository/postgres"
	catalog_service "sport_app/internal/features/catalog/service"
	catalog_transport_http "sport_app/internal/features/catalog/transport/http"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCatalogIntegration(t *testing.T) {
	database := os.Getenv("SPORTIK_CATALOG_TEST_DB")
	if database == "" {
		t.Skip("set SPORTIK_CATALOG_TEST_DB to a new disposable database on 127.0.0.1:55439")
	}
	require.True(t, strings.HasPrefix(database, "sportik_catalog_test_"), "only a dedicated test database is allowed")
	ctx := context.Background()
	pool, err := core_pgx_pool.NewPool(ctx, core_pgx_pool.Config{
		Host: "127.0.0.1", Port: "55439", User: "sportik_test", Database: database, Timeout: 10 * time.Second,
	})
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
	files, err := filepath.Glob("../../../../../migrations/00000[1-6]_*.up.sql")
	require.NoError(t, err)
	for _, path := range files {
		apply(filepath.Base(path))
	}
	apply("000007_catalog.up.sql")
	apply("000007_catalog.down.sql")
	apply("000007_catalog.up.sql")
	apply("000007_catalog.down.sql")
	_, err = pool.Exec(ctx, `INSERT INTO sportapp.users (id, is_anonymous, created_at) VALUES
		(101, FALSE, NOW()), (102, FALSE, NOW()), (103, TRUE, NOW());
		INSERT INTO sportapp.profile (user_id, age, weight_kg, training_days_map, created_at)
		VALUES (101, 26, 75, ARRAY['Пн', 'Ср', 'Пт'], NOW()), (102, 30, 80, ARRAY['Вт'], NOW());
		INSERT INTO sportapp.user_data (user_id, completed_workouts, created_at)
		VALUES (101, '[{"day_code":"A"}]', NOW()), (102, '[]', NOW());
		INSERT INTO sportapp.notes (user_id, title, body) VALUES (101, 'fixture', 'preserve me');`)
	require.NoError(t, err)
	users := users_postgres_repository.NewUsersRepository(pool)
	require.NoError(t, users.EnsureExercisesSeeded(ctx))
	var exerciseCount int
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.exercises`).Scan(&exerciseCount))
	require.Equal(t, 67, exerciseCount)
	apply("000007_catalog.up.sql")
	repo := catalog_postgres_repository.NewRepository(pool)
	require.NoError(t, repo.EnsureReferenceData(ctx))
	require.NoError(t, repo.EnsureReferenceData(ctx))
	var roles, unmapped int
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.user_roles`).Scan(&roles))
	require.Equal(t, 2, roles)
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.exercises e
		WHERE NOT EXISTS (SELECT 1 FROM sportapp.exercise_equipment ee WHERE ee.exercise_id = e.id)`).Scan(&unmapped))
	require.Zero(t, unmapped)
	require.NoError(t, repo.SeedCatalog(ctx, 101, "Тестовый автор"))
	require.NoError(t, repo.SeedCatalog(ctx, 101, "Имя не перезаписывается"))
	require.ErrorIs(t, repo.SeedCatalog(ctx, 102, "Другой автор"), core_errors.ErrConflict)
	var otherAuthor bool
	require.NoError(t, pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sportapp.trainer_profiles WHERE user_id = 102)`).Scan(&otherAuthor))
	require.False(t, otherAuthor)
	var count int
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.programs`).Scan(&count))
	require.Equal(t, 2, count)
	exists, err := repo.EquipmentExists(ctx, []string{"missing"})
	require.NoError(t, err)
	require.False(t, exists)
	_, err = pool.Exec(ctx, `INSERT INTO sportapp.programs (id, trainer_id, status)
		VALUES ('18000000-0000-4000-8000-000000000003', 101, 'hidden'),
		('18000000-0000-4000-8000-000000000004', 101, 'active'),
		('18000000-0000-4000-8000-000000000005', 101, 'active');
		INSERT INTO sportapp.program_versions (id, program_id, number, title, description, goal, level,
			duration_weeks, sessions_per_week, price_minor, currency, status)
		SELECT id, id, 1, 'fixture', 'fixture', 'Набрать мышцы', 'новичок', 4, 2,
			CASE WHEN id::text LIKE '%5' THEN 10000 ELSE 0 END, 'RUB',
			CASE WHEN id::text LIKE '%4' THEN 'draft' ELSE 'published' END
		FROM sportapp.programs WHERE id::text IN ('18000000-0000-4000-8000-000000000003',
			'18000000-0000-4000-8000-000000000004', '18000000-0000-4000-8000-000000000005');
		UPDATE sportapp.programs SET published_version_id = id WHERE id::text IN (
			'18000000-0000-4000-8000-000000000003', '18000000-0000-4000-8000-000000000004', '18000000-0000-4000-8000-000000000005');`)
	require.NoError(t, err)
	svc := catalog_service.NewService(repo)
	jwt := core_auth.NewJWT(core_auth.Config{Secret: "catalog-test-only", TokenTTL: time.Hour})
	router := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion2)
	router.RegisterRoutes(catalog_transport_http.NewCatalogHTTPHandler(svc, jwt).Routes()...)
	handler := core_http_middleware.RequestID()(http.StripPrefix("/api/v2", router))
	token := func(id string, anonymous bool) string {
		t.Helper()
		value, err := jwt.CreateToken(id, anonymous)
		require.NoError(t, err)
		return value
	}
	memberToken := token("102", false)
	guestToken := token("103", false)
	request := func(method, path, bearer string, status int) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, "/api/v2"+path, nil)
		if bearer != "" {
			r.Header.Set("Authorization", "Bearer "+bearer)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		require.Equal(t, status, w.Code, "%s %s: %s", method, path, w.Body.String())
		if status >= 400 {
			var body struct {
				Error     struct{ Code, Message string }
				RequestID string `json:"request_id"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
			require.NotEmpty(t, body.Error.Code)
			require.NotEmpty(t, body.RequestID)
			assert.NotContains(t, w.Body.String(), "sportapp")
		}
		return w
	}
	const freeID = "18000000-0000-4000-8000-000000000001"
	const paidID = "18000000-0000-4000-8000-000000000005"
	t.Run("catalog pagination and filters", func(t *testing.T) {
		seen := map[string]bool{}
		cursor := ""
		for {
			w := request("GET", "/catalog/programs?limit=1&cursor="+url.QueryEscape(cursor), "", 200)
			var page catalog_postgres_repository.Page[catalog_postgres_repository.ProgramCard]
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &page))
			require.Len(t, page.Items, 1)
			assert.False(t, seen[page.Items[0].ID])
			seen[page.Items[0].ID] = true
			if page.NextCursor == nil {
				break
			}
			cursor = *page.NextCursor
		}
		assert.Len(t, seen, 3)
		for _, filter := range []string{"goal=unknown", "level=unknown", "duration_weeks=99", "trainer_id=999999", "equipment=barbell&price_type=free"} {
			w := request("GET", "/catalog/programs?"+filter, "", 200)
			assert.Contains(t, w.Body.String(), `"items":[]`)
		}
		w := request("GET", "/catalog/programs?equipment=dumbbell&equipment=bench&price_type=free", "", 200)
		assert.Contains(t, w.Body.String(), freeID)
		assert.NotContains(t, w.Body.String(), "Базовая программа в зале")
		for _, sort := range []string{"newest", "price_asc", "price_desc"} {
			request("GET", "/catalog/programs?limit=1&sort="+sort, "", 200)
		}
		for _, filter := range []string{"limit=0", "limit=101", "duration_weeks=-1", "trainer_id=x", "equipment=missing", "sort=invalid", "price_type=invalid", "cursor=invalid"} {
			request("GET", "/catalog/programs?"+filter, "", 400)
		}
	})
	t.Run("cursor survives publication changes", func(t *testing.T) {
		for _, sort := range []string{"price_asc", "price_desc"} {
			filter := catalog_postgres_repository.Filter{Limit: 1, Sort: sort}
			first, err := svc.ListPrograms(ctx, filter, "")
			require.NoError(t, err)
			require.NotNil(t, first.NextCursor)
			anchor := first.Items[0]
			all, err := svc.ListPrograms(ctx, catalog_postgres_repository.Filter{Sort: sort}, "")
			require.NoError(t, err)
			require.Greater(t, len(all.Items), 1)
			_, err = pool.Exec(ctx, `UPDATE sportapp.programs SET published_version_id = NULL WHERE id = $1`, anchor.ID)
			require.NoError(t, err)
			next, readErr := svc.ListPrograms(ctx, filter, *first.NextCursor)
			_, err = pool.Exec(ctx, `UPDATE sportapp.programs SET published_version_id = $2 WHERE id = $1`, anchor.ID, anchor.PublishedVersionID)
			require.NoError(t, err)
			require.NoError(t, readErr)
			if assert.Len(t, next.Items, 1, sort) {
				assert.Equal(t, all.Items[1].ID, next.Items[0].ID, sort)
			}
		}
	})
	t.Run("cursor retains original price", func(t *testing.T) {
		for _, sort := range []string{"price_asc", "price_desc"} {
			filter := catalog_postgres_repository.Filter{Limit: 1, Sort: sort}
			first, err := svc.ListPrograms(ctx, filter, "")
			require.NoError(t, err)
			require.NotNil(t, first.NextCursor)
			all, err := svc.ListPrograms(ctx, catalog_postgres_repository.Filter{Sort: sort}, "")
			require.NoError(t, err)
			anchor := first.Items[0]
			newPrice := int64(100000)
			if sort == "price_desc" {
				newPrice = 0
			}
			_, err = pool.Exec(ctx, `UPDATE sportapp.program_versions SET price_minor = $2 WHERE id = $1`, anchor.PublishedVersionID, newPrice)
			require.NoError(t, err)
			next, readErr := svc.ListPrograms(ctx, filter, *first.NextCursor)
			_, err = pool.Exec(ctx, `UPDATE sportapp.program_versions SET price_minor = $2 WHERE id = $1`, anchor.PublishedVersionID, anchor.PriceMinor)
			require.NoError(t, err)
			require.NoError(t, readErr)
			if assert.Len(t, next.Items, 1, sort) {
				assert.Equal(t, all.Items[1].ID, next.Items[0].ID, sort)
			}
		}
	})
	t.Run("UUID aliases", func(t *testing.T) {
		_, err := svc.GetProgram(ctx, "urn:uuid:"+freeID, 0)
		assert.NoError(t, err)
		_, err = svc.GrantFreeAccess(ctx, "urn:uuid:"+freeID, 101)
		assert.NoError(t, err)
		_, err = pool.Exec(ctx, `DELETE FROM sportapp.entitlements WHERE user_id = 101`)
		require.NoError(t, err)
	})
	t.Run("reference data", func(t *testing.T) {
		request("GET", "/equipment", "", 401)
		request("GET", "/exercises", "bad-token", 401)
		w := request("GET", "/equipment?limit=1", guestToken, 200)
		assert.Contains(t, w.Body.String(), `"next_cursor":"`)
		w = request("GET", "/exercises?query="+url.QueryEscape("гантел")+"&equipment=dumbbell&equipment=bench", guestToken, 200)
		assert.Contains(t, w.Body.String(), "гантел")
		request("GET", "/exercises?equipment=missing", memberToken, 400)
		_, err := pool.Exec(ctx, `UPDATE sportapp.exercises SET is_active = FALSE WHERE id = 1`)
		require.NoError(t, err)
		items, err := repo.ListExercises(ctx, catalog_postgres_repository.Filter{Limit: 100})
		require.NoError(t, err)
		assert.Len(t, items, 66)
	})
	t.Run("access and privacy", func(t *testing.T) {
		w := request("GET", "/catalog/programs/"+freeID, "", 200)
		assert.NotContains(t, w.Body.String(), "exercise_id")
		request("GET", "/catalog/programs/"+freeID, "bad-token", 401)
		for _, id := range []string{"18000000-0000-4000-8000-000000000003", "18000000-0000-4000-8000-000000000004", uuid.NewString()} {
			request("GET", "/catalog/programs/"+id, memberToken, 404)
			request("POST", "/programs/"+id+"/access", memberToken, 404)
		}
		request("GET", "/catalog/programs/invalid", "", 400)
		request("POST", "/programs/"+freeID+"/access", "", 401)
		request("POST", "/programs/"+freeID+"/access", guestToken, 403)
		request("POST", "/programs/"+freeID+"/access", token("999999", false), 401)
		request("POST", "/programs/"+paidID+"/access", memberToken, 422)
		first := request("POST", "/programs/"+freeID+"/access", memberToken, 201)
		second := request("POST", "/programs/"+freeID+"/access", memberToken, 201)
		assert.JSONEq(t, first.Body.String(), second.Body.String())
		w = request("GET", "/catalog/programs/"+freeID, memberToken, 200)
		assert.Contains(t, w.Body.String(), `"access":"free"`)
		assert.Contains(t, w.Body.String(), "exercise_id")
		w = request("GET", "/catalog/programs/"+freeID, token("101", false), 200)
		assert.NotContains(t, w.Body.String(), "exercise_id")
		var wg sync.WaitGroup
		errs := make(chan error, 8)
		for range 8 {
			wg.Go(func() { _, err := repo.GrantFreeAccess(ctx, freeID, 102); errs <- err })
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			require.NoError(t, err)
		}
		var grants int
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.entitlements WHERE user_id = 102 AND program_id = $1`, freeID).Scan(&grants))
		require.Equal(t, 1, grants)
		_, err := pool.Exec(ctx, `UPDATE sportapp.entitlements SET status = 'revoked' WHERE user_id = 102`)
		require.NoError(t, err)
		w = request("GET", "/catalog/programs/"+freeID, memberToken, 200)
		assert.NotContains(t, w.Body.String(), "exercise_id")
		request("POST", "/programs/"+freeID+"/access", memberToken, 201)
		reconnected, err := core_pgx_pool.NewPool(ctx, core_pgx_pool.Config{Host: "127.0.0.1", Port: "55439", User: "sportik_test", Database: database, Timeout: time.Second})
		require.NoError(t, err)
		defer reconnected.Close()
		detail, err := catalog_postgres_repository.NewRepository(reconnected).GetProgram(ctx, freeID, 102)
		require.NoError(t, err)
		assert.Len(t, detail.Days, 3)
		assert.Len(t, detail.Days[0].Exercises, 5)
	})
	t.Run("registration roles and v1 profile", func(t *testing.T) {
		user, err := users.CreateUser(ctx, "catalog-test@example.test", "test-hash")
		require.NoError(t, err)
		guest, err := users.CreateGuestUser(ctx)
		require.NoError(t, err)
		promoted, err := users.PromoteGuestToUser(ctx, guest, "catalog-promoted@example.test", "test-hash")
		require.NoError(t, err)
		for _, id := range []int{user.ID, promoted.ID} {
			var hasRole bool
			require.NoError(t, pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sportapp.user_roles WHERE user_id = $1 AND role = 'user')`, id).Scan(&hasRole))
			assert.True(t, hasRole)
		}
		profile, err := users.GetProfile(ctx, "101")
		require.NoError(t, err)
		assert.Equal(t, 75, *profile.WeightKg)
		assert.Equal(t, []string{"Пн", "Ср", "Пт"}, profile.TrainingDaysMap)
	})
	t.Run("rollback preserves v1 then reapply", func(t *testing.T) {
		apply("000007_catalog.down.sql")
		var body string
		require.NoError(t, pool.QueryRow(ctx, `SELECT body FROM sportapp.notes WHERE user_id = 101`).Scan(&body))
		assert.Equal(t, "preserve me", body)
		require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.exercises`).Scan(&count))
		assert.Equal(t, 67, count)
		profile, err := users.GetProfile(ctx, "102")
		require.NoError(t, err)
		assert.Equal(t, 80, *profile.WeightKg)
		apply("000007_catalog.up.sql")
		require.NoError(t, repo.EnsureReferenceData(ctx))
		require.NoError(t, repo.SeedCatalog(ctx, 101, "Тестовый автор"))
		items, err := svc.ListPrograms(ctx, catalog_postgres_repository.Filter{}, "")
		require.NoError(t, err)
		assert.Len(t, items.Items, 2)
		_, err = repo.GrantFreeAccess(ctx, freeID, 103)
		assert.True(t, errors.Is(err, core_errors.ErrForbidden))
	})
}
