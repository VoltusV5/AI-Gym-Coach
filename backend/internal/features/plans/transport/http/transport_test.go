package plans_transport_http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	core_auth "sport_app/internal/core/auth"
	core_errors "sport_app/internal/core/errors"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_server "sport_app/internal/core/transport/http/server"
	plans_postgres_repository "sport_app/internal/features/plans/repository/postgres"
	plans_service "sport_app/internal/features/plans/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validBody = `{"title":"План","days":[{"title":"День","position":0,"exercises":[{"exercise_id":1,"position":0,"sets":3,"reps_min":8,"reps_max":12,"weight_kg":null,"rest_sec":0}]}],"schedule":{"time_zone":"UTC","start_date":"2026-10-29","entries":[]}}`

type testRepository struct {
	plans_service.Repository
	calls int
	owner int
}

func (r *testRepository) CreatePlan(_ context.Context, owner int, input plans_postgres_repository.Input) (plans_postgres_repository.Plan, error) {
	r.calls++
	r.owner = owner
	return plans_postgres_repository.Plan{ID: "18000000-0000-4000-8000-000000000001", Version: 1, Title: input.Title, Days: input.Days, Schedule: input.Schedule, SourceType: "manual"}, nil
}

func TestPlansHTTPValidation(t *testing.T) {
	for _, tt := range []struct {
		name, body, identity string
		status               int
	}{
		{"success", validBody, "101", 201},
		{"guest", validBody, "103", 201},
		{"no JWT", validBody, "", 401},
		{"invalid JWT", validBody, "invalid", 401},
		{"bad subject", validBody, "abc", 401},
		{"missing days", `{"title":"Plan","schedule":{"time_zone":"UTC","start_date":"2026-10-29","entries":[]}}`, "101", 400},
		{"null", `null`, "101", 400},
		{"null exercises", strings.Replace(validBody, `"exercises":[{"exercise_id":1,"position":0,"sets":3,"reps_min":8,"reps_max":12,"weight_kg":null,"rest_sec":0}]`, `"exercises":null`, 1), "101", 400},
		{"missing position", strings.Replace(validBody, `"position":0,`, "", 1), "101", 400},
		{"missing rest", strings.Replace(validBody, `,"rest_sec":0`, "", 1), "101", 400},
		{"null day id", strings.Replace(validBody, `"title":"День"`, `"id":null,"title":"День"`, 1), "101", 400},
		{"extra JSON", validBody + ` {}`, "101", 400},
		{"owner injection", strings.Replace(validBody, `"title":"План"`, `"owner_id":102,"title":"План"`, 1), "101", 400},
		{"source injection", strings.Replace(validBody, `"title":"План"`, `"source_type":"ai","title":"План"`, 1), "101", 400},
		{"oversized", strings.Repeat("x", (1<<20)+1), "101", 400},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testRepository{}
			jwt := core_auth.NewJWT(core_auth.Config{Secret: "plan-test-only", TokenTTL: time.Hour})
			router := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion2)
			router.RegisterRoutes(NewPlansHTTPHandler(plans_service.NewService(repo), jwt).Routes()...)
			r := httptest.NewRequest("POST", "/me/plans", bytes.NewBufferString(tt.body))
			if tt.identity != "" {
				token := tt.identity
				if token != "invalid" {
					var err error
					token, err = jwt.CreateToken(tt.identity, tt.identity == "103")
					require.NoError(t, err)
				}
				r.Header.Set("Authorization", "Bearer "+token)
			}
			w := httptest.NewRecorder()
			core_http_middleware.RequestID()(router).ServeHTTP(w, r)
			require.Equal(t, tt.status, w.Code, w.Body.String())
			if tt.status == 201 {
				assert.Equal(t, 1, repo.calls)
				assert.NotEmpty(t, w.Body.String())
				assert.NotZero(t, repo.owner)
			} else {
				assert.Zero(t, repo.calls)
				assert.Contains(t, w.Body.String(), `"request_id":`)
			}
		})
	}
}

func TestVersionBodyRequired(t *testing.T) {
	for _, body := range []string{`{}`, `{"version":null}`, `{"version":1,"owner_id":102}`, `{"version":1} {}`} {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		w := httptest.NewRecorder()
		var version struct {
			Version int64 `json:"version"`
		}
		_, err := decodeBody(w, r, &version, "version")
		assert.ErrorIs(t, err, core_errors.ErrInvalidArgument)
	}
	r := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(validBody))
	_, err := decodeInput(httptest.NewRecorder(), r, true)
	assert.ErrorIs(t, err, core_errors.ErrInvalidArgument)
	var input plans_postgres_repository.Input
	require.NoError(t, json.Unmarshal([]byte(validBody), &input))
}
