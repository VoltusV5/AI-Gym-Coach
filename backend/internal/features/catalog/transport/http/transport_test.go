package catalog_transport_http_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	core_auth "sport_app/internal/core/auth"
	core_errors "sport_app/internal/core/errors"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_server "sport_app/internal/core/transport/http/server"
	catalog_postgres_repository "sport_app/internal/features/catalog/repository/postgres"
	catalog_service "sport_app/internal/features/catalog/service"
	catalog_transport_http "sport_app/internal/features/catalog/transport/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testRepository struct {
	catalog_service.Repository
	err        error
	userID     int
	calls      int
	panicValue any
}

func (r *testRepository) ListEquipment(context.Context, catalog_postgres_repository.Filter) ([]catalog_postgres_repository.Equipment, error) {
	r.calls++
	return []catalog_postgres_repository.Equipment{{ID: "bench", Code: "bench", Name: "Скамья"}}, r.err
}

func (r *testRepository) ListExercises(context.Context, catalog_postgres_repository.Filter) ([]catalog_postgres_repository.Exercise, error) {
	r.calls++
	return []catalog_postgres_repository.Exercise{}, r.err
}

func (r *testRepository) ListPrograms(context.Context, catalog_postgres_repository.Filter) ([]catalog_postgres_repository.ProgramCard, error) {
	if r.panicValue != nil {
		panic(r.panicValue)
	}
	r.calls++
	return []catalog_postgres_repository.ProgramCard{}, r.err
}

func (r *testRepository) GetProgram(_ context.Context, _ string, userID int) (catalog_postgres_repository.ProgramDetail, error) {
	r.calls++
	r.userID = userID
	return catalog_postgres_repository.ProgramDetail{Access: "preview", AllowedActions: []string{"access"}}, r.err
}

func (r *testRepository) GrantFreeAccess(_ context.Context, id string, userID int) (catalog_postgres_repository.Entitlement, error) {
	r.calls++
	r.userID = userID
	return catalog_postgres_repository.Entitlement{ID: "18000000-0000-4000-8000-000000000010", ProgramID: id, Status: "active"}, r.err
}

func TestCatalogHTTP(t *testing.T) {
	const id = "18000000-0000-4000-8000-000000000001"
	for _, tt := range []struct {
		name, method, path, identity string
		err                          error
		panicValue                   any
		status, calls, userID        int
	}{
		{name: "panic redacted", method: "GET", path: "/catalog/programs", panicValue: "postgres://secret-host/password", status: 500},
		{name: "public catalog", method: "GET", path: "/catalog/programs", status: 200, calls: 1},
		{name: "public detail", method: "GET", path: "/catalog/programs/" + id, status: 200, calls: 1},
		{name: "authenticated detail", method: "GET", path: "/catalog/programs/" + id, identity: "42", status: 200, calls: 1, userID: 42},
		{name: "equipment needs JWT", method: "GET", path: "/equipment", status: 401},
		{name: "exercise invalid JWT", method: "GET", path: "/exercises", identity: "invalid-token", status: 401},
		{name: "guest reference data", method: "GET", path: "/equipment", identity: "guest", status: 200, calls: 1},
		{name: "exercises", method: "GET", path: "/exercises", identity: "42", status: 200, calls: 1},
		{name: "grant", method: "POST", path: "/programs/" + id + "/access", identity: "42", status: 201, calls: 1, userID: 42},
		{name: "grant missing JWT", method: "POST", path: "/programs/" + id + "/access", status: 401},
		{name: "bad subject", method: "POST", path: "/programs/" + id + "/access", identity: "abc", status: 401},
		{name: "bad id", method: "GET", path: "/catalog/programs/not-a-uuid", status: 400},
		{name: "bad limit", method: "GET", path: "/catalog/programs?limit=101", status: 400},
		{name: "large trainer id", method: "GET", path: "/catalog/programs?trainer_id=4294967296", status: 400},
		{name: "large duration", method: "GET", path: "/catalog/programs?duration_weeks=4294967296", status: 400},
		{name: "large subject", method: "GET", path: "/catalog/programs/" + id, identity: "4294967296", status: 401},
		{name: "bad duration", method: "GET", path: "/catalog/programs?duration_weeks=x", status: 400},
		{name: "bad cursor", method: "GET", path: "/catalog/programs?cursor=bad", status: 400},
		{name: "hidden", method: "GET", path: "/catalog/programs/" + id, err: core_errors.ErrNotFound, status: 404, calls: 1},
		{name: "guest access rejected", method: "POST", path: "/programs/" + id + "/access", identity: "guest", err: core_errors.ErrForbidden, status: 403, calls: 1, userID: 103},
		{name: "paid access rejected", method: "POST", path: "/programs/" + id + "/access", identity: "42", err: core_errors.ErrUnprocessable, status: 422, calls: 1, userID: 42},
		{name: "dependency error redacted", method: "GET", path: "/catalog/programs", err: errors.New("postgres://secret-host/password"), status: 500, calls: 1},
		{name: "timeout", method: "GET", path: "/catalog/programs", err: context.DeadlineExceeded, status: 504, calls: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testRepository{err: tt.err, panicValue: tt.panicValue}
			jwt := core_auth.NewJWT(core_auth.Config{Secret: "unit-test-only", TokenTTL: time.Hour})
			router := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion2)
			router.RegisterRoutes(catalog_transport_http.NewCatalogHTTPHandler(catalog_service.NewService(repo), jwt).Routes()...)
			r := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.identity != "" {
				bearer := tt.identity
				if tt.identity != "invalid-token" {
					sub := tt.identity
					if sub == "guest" {
						sub = "103"
					}
					var err error
					bearer, err = jwt.CreateToken(sub, tt.identity == "guest")
					require.NoError(t, err)
				}
				r.Header.Set("Authorization", "Bearer "+bearer)
			}
			w := httptest.NewRecorder()
			core_http_middleware.RequestID()(core_http_middleware.Panic()(router)).ServeHTTP(w, r)
			require.Equal(t, tt.status, w.Code, w.Body.String())
			assert.Equal(t, tt.calls, repo.calls)
			assert.Equal(t, tt.userID, repo.userID)
			if tt.status >= 400 {
				assert.Contains(t, w.Body.String(), `"request_id":"`)
				assert.NotContains(t, w.Body.String(), "secret-host")
				assert.NotContains(t, w.Body.String(), "password")
			}
		})
	}
}
