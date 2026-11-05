package profile_transport_http_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	core_auth "sport_app/internal/core/auth"
	core_errors "sport_app/internal/core/errors"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_server "sport_app/internal/core/transport/http/server"
	profile_postgres_repository "sport_app/internal/features/profile/repository/postgres"
	profile_service "sport_app/internal/features/profile/service"
	profile_transport_http "sport_app/internal/features/profile/transport/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type repositoryStub struct {
	get    func(context.Context, int) (profile_postgres_repository.Profile, error)
	update func(context.Context, int, profile_postgres_repository.Patch) (profile_postgres_repository.Profile, error)
}

// GetProfile передаёт запрос в тестовую функцию вместо реальной БД.
func (r repositoryStub) GetProfile(ctx context.Context, id int) (profile_postgres_repository.Profile, error) {
	return r.get(ctx, id)
}

// UpdateProfile передаёт валидированный patch в тестовую функцию.
func (r repositoryStub) UpdateProfile(ctx context.Context, id int, p profile_postgres_repository.Patch) (profile_postgres_repository.Profile, error) {
	return r.update(ctx, id, p)
}

// handlerFor создаёт настоящий router, middleware JWT и сервис с управляемой зависимостью.
func handlerFor(repo repositoryStub) (http.Handler, *core_auth.JWT) {
	jwt := core_auth.NewJWT(core_auth.Config{Secret: "profile-v2-test-only-secret", TokenTTL: time.Hour})
	router := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion2)
	router.RegisterRoutes(profile_transport_http.NewProfileHTTPHandler(profile_service.NewService(repo), jwt).Routes()...)
	return core_http_middleware.RequestID()(router), jwt
}

// TestProfileHTTP проверяет общий путь JWT → HTTP → service и безопасные ошибки без внешних сервисов.
func TestProfileHTTP(t *testing.T) {
	writes := 0
	repo := repositoryStub{
		get: func(_ context.Context, id int) (profile_postgres_repository.Profile, error) {
			assert.Equal(t, 42, id)
			return profile_postgres_repository.Profile{Version: 3, TrainingDays: []int{}}, nil
		},
		update: func(_ context.Context, id int, p profile_postgres_repository.Patch) (profile_postgres_repository.Profile, error) {
			assert.Equal(t, 42, id)
			writes++
			return profile_postgres_repository.Profile{Version: p.Version + 1, WeightKg: p.WeightKg, TrainingDays: []int{}}, nil
		},
	}
	handler, jwt := handlerFor(repo)
	token, err := jwt.CreateToken("42", true)
	require.NoError(t, err)
	request := func(method, body, bearer string, status int) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, "/me/profile", strings.NewReader(body))
		if bearer != "" {
			r.Header.Set("Authorization", "Bearer "+bearer)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		require.Equal(t, status, w.Code, "%s", w.Body.String())
		if status >= 400 {
			var response struct {
				Error     struct{ Code string }
				RequestID string `json:"request_id"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
			assert.NotEmpty(t, response.Error.Code)
			assert.NotEmpty(t, response.RequestID)
		}
		return w
	}
	request("GET", "", token, 200)
	for _, method := range []string{"GET", "PATCH"} {
		request(method, `{"version":3,"weight_kg":75.25}`, "", 401)
		request(method, `{"version":3,"weight_kg":75.25}`, "invalid", 401)
	}
	for _, id := range []string{"not-a-user", "0", "2147483648"} {
		badIDToken, err := jwt.CreateToken(id, false)
		require.NoError(t, err)
		request("GET", "", badIDToken, 401)
		request("PATCH", `{"version":3,"age":25}`, badIDToken, 401)
	}
	assert.Zero(t, writes)
	invalidBodies := []string{
		``, `null`, `[]`, `{}`, `{"version":3}`, `{"version":0,"age":20}`,
		`{"version":3.5,"age":20}`, `{"version":3,"user_id":7,"age":20}`,
		`{"version":3,"unknown":1}`, `{"version":3,"weight_kg":null}`,
		`{"version":3,"weight_kg":"75.5"}`, `{"version":3,"weight_kg":-1}`,
		`{"version":3,"weight_kg":1e100}`, `{"version":3,"age":0}`,
		`{"version":3,"height_cm":-10}`, `{"version":3,"gender":"  "}`,
		`{"version":3,"goal":""}`, `{"version":3,"activity_level":""}`,
		`{"version":3,"gender":"\u0000"}`, `{"version":3,"activity_level":"a\u0000"}`,
		`{"version":3,"goal":"\u0000b"}`, `{"version":3,"equipment_ids":["\u0000"]}`,
		`{"version":3,"session_duration_min":0}`, `{"version":3,"level":"expert"}`,
		`{"version":3,"training_days":[0,8]}`, `{"version":3,"training_days":[1,1]}`,
		`{"version":3,"training_days":[1.5]}`, `{"version":3,"equipment_ids":["bench","bench"]}`,
		`{"version":3,"equipment_ids":[""]}`, `{"version":3,"equipment_ids":null}`,
		`{"version":3,"excluded_exercise_ids":[0]}`, `{"version":3,"excluded_exercise_ids":[1,1]}`,
		`{"version":3,"age":20} {}`, `{"version":3,"age":20}broken`,
		`{"version":3,"goal":"` + strings.Repeat("я", 201) + `"}`,
		`{"version":3,"goal":"` + strings.Repeat("a", 1<<20) + `"}`,
	}
	for _, body := range invalidBodies {
		request("PATCH", body, token, 400)
	}
	assert.Zero(t, writes, "invalid requests must never reach the repository")
	w := request("PATCH", `{"version":3,"weight_kg":75.25,"training_days":[1,7],"equipment_ids":[],"excluded_exercise_ids":[],"injuries_notes":false}`, token, 200)
	assert.Contains(t, w.Body.String(), `"weight_kg":75.25`)
	assert.Contains(t, w.Body.String(), `"version":4`)
	assert.Equal(t, 1, writes)
}

// TestProfileHTTPDependencyErrors проверяет статусы и отсутствие SQL/личных данных в ответах v2.
func TestProfileHTTPDependencyErrors(t *testing.T) {
	for _, tt := range []struct {
		err    error
		status int
	}{
		{core_errors.ErrNotFound, 404}, {core_errors.ErrConflict, 409},
		{core_errors.ErrInvalidArgument, 400}, {errors.New("secret SQL sportapp.profile password and personal data"), 500},
	} {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			repo := repositoryStub{
				get: func(context.Context, int) (profile_postgres_repository.Profile, error) {
					return profile_postgres_repository.Profile{}, tt.err
				},
				update: func(context.Context, int, profile_postgres_repository.Patch) (profile_postgres_repository.Profile, error) {
					return profile_postgres_repository.Profile{}, tt.err
				},
			}
			handler, jwt := handlerFor(repo)
			token, err := jwt.CreateToken("42", false)
			require.NoError(t, err)
			for _, method := range []string{"GET", "PATCH"} {
				r := httptest.NewRequest(method, "/me/profile", strings.NewReader(`{"version":3,"age":25}`))
				r.Header.Set("Authorization", "Bearer "+token)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, r)
				assert.Equal(t, tt.status, w.Code)
				assert.NotContains(t, w.Body.String(), "sportapp")
				assert.NotContains(t, w.Body.String(), "password")
			}
		})
	}
}
