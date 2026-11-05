package mlclient

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	core_errors "sport_app/internal/core/errors"
	plans_postgres_repository "sport_app/internal/features/plans/repository/postgres"
)

func responseV2(id string) GenerateResult {
	return GenerateResult{RequestID: id, ModelVersion: "transformer-test", Split: "Full body", Days: []GeneratedDay{{Weekday: 1, Title: "Day A", Exercises: []plans_postgres_repository.Prescription{{ExerciseID: 1, Position: 0, Sets: 3, RepsMin: 8, RepsMax: 12, RestSec: 90}}}}, Warnings: []string{}, Fallback: false}
}
func TestV2ClientContractAndAuthorization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer synthetic-service-token", r.Header.Get("Authorization"))
		require.Equal(t, "/v2/plans/generate", r.URL.Path)
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, "req-test", payload["request_id"])
		require.Equal(t, float64(808), payload["exercises"].([]any)[0].(map[string]any)["exercise_id"])
		require.Equal(t, []any{}, payload["profile"].(map[string]any)["excluded_exercise_ids"])
		require.NoError(t, json.NewEncoder(w).Encode(responseV2("req-test")))
	}))
	defer server.Close()
	result, err := NewV2Client(Config{BaseURL: server.URL, ServiceToken: "synthetic-service-token"}).Generate(context.Background(), GenerateInput{RequestID: "req-test", FeatureSchemaVersion: "profile-v2", Exercises: []ExerciseCandidate{{ExerciseID: 808, MuscleGroup: "Ноги", EquipmentIDs: []string{}}}})
	require.NoError(t, err)
	require.Equal(t, "transformer-test", result.ModelVersion)
	_, err = NewV2Client(Config{BaseURL: server.URL}).Generate(context.Background(), GenerateInput{})
	var domain *core_errors.DomainError
	require.ErrorAs(t, err, &domain)
	require.Equal(t, "ML_UNAVAILABLE", domain.Code)
}
func TestV2ClientRejectsInvalidAndDoesNotForwardRedirectCredentials(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{{"missing fields", 200, `{"request_id":"req-test","model_version":"m","split":"x","days":[{"weekday":1,"title":"d","exercises":[{"exercise_id":1,"sets":3,"reps_min":8,"reps_max":12,"rest_sec":90}]}],"fallback":false}`}, {"wrong request", 200, `{"request_id":"wrong","model_version":"m","split":"x","days":[],"fallback":false}`}, {"dependency", 503, `sensitive internal text`}, {"infeasible", 422, `{}`}, {"trailing JSON", 200, `{} {}`}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			_, err := NewV2Client(Config{BaseURL: server.URL, ServiceToken: "test"}).Generate(context.Background(), GenerateInput{RequestID: "req-test", FeatureSchemaVersion: "profile-v2"})
			require.Error(t, err)
			require.NotContains(t, err.Error(), "sensitive")
		})
	}
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("redirect must not be followed") }))
	defer target.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	_, err := NewV2Client(Config{BaseURL: server.URL, ServiceToken: "test"}).Generate(context.Background(), GenerateInput{})
	require.Error(t, err)
}
func TestV2ClientDeadlineBeforeAndAfterHeaders(t *testing.T) {
	for _, flush := range []bool{false, true} {
		t.Run(map[bool]string{false: "headers", true: "body"}[flush], func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.Copy(io.Discard, r.Body)
				if flush {
					w.WriteHeader(200)
					w.(http.Flusher).Flush()
				}
				select {
				case <-r.Context().Done():
				case <-time.After(200 * time.Millisecond):
				}
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
			defer cancel()
			_, err := NewV2Client(Config{BaseURL: server.URL, ServiceToken: "test"}).Generate(ctx, GenerateInput{})
			require.True(t, errors.Is(err, context.DeadlineExceeded), "%v", err)
		})
	}
}
