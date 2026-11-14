package mlclient

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRankRejectsUnknownVersionsAndDuplicateCandidates(t *testing.T) {
	input := RankInput{RequestID: "rank-test", Limit: 2, Candidates: []Candidate{{ProgramID: "p", VersionID: "v"}}}
	for _, tc := range []struct {
		name    string
		items   []RankItem
		success bool
	}{
		{"valid", []RankItem{{ProgramID: "p", VersionID: "v", ReasonCodes: []string{}, Evidence: Evidence{}}}, true},
		{"foreign program", []RankItem{{ProgramID: "foreign", VersionID: "v"}}, false},
		{"stale version", []RankItem{{ProgramID: "p", VersionID: "old"}}, false},
		{"duplicate", []RankItem{{ProgramID: "p", VersionID: "v"}, {ProgramID: "p", VersionID: "v"}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "Bearer fixture-token", r.Header.Get("Authorization"))
				require.Equal(t, input.RequestID, r.Header.Get("X-Request-ID"))
				require.Equal(t, "/v2/recommendations/rank", r.URL.Path)
				_ = json.NewEncoder(w).Encode(RankResult{RequestID: input.RequestID, ModelVersion: "test", Items: tc.items})
			}))
			defer server.Close()
			_, err := NewV2Client(Config{BaseURL: server.URL, ServiceToken: "fixture-token"}).Rank(context.Background(), input)
			require.Equal(t, tc.success, err == nil)
		})
	}
}
func TestRankDeadlineAndExplainPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/recommendations/rank" {
			select {
			case <-r.Context().Done():
				return
			case <-time.After(2 * time.Second):
				return
			}
		}
		var in map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&in))
		item := in["items"].([]any)[0].(map[string]any)
		require.NotContains(t, item, "score")
		require.Equal(t, "version", item["version_id"])
		_ = json.NewEncoder(w).Encode(ExplainResult{RequestID: "explain", Items: []Explanation{}})
	}))
	defer server.Close()
	client := NewV2Client(Config{BaseURL: server.URL, ServiceToken: "fixture-token"})
	start := time.Now()
	_, err := client.Rank(context.Background(), RankInput{RequestID: "slow", Limit: 1})
	require.Error(t, err)
	require.Less(t, time.Since(start), 1500*time.Millisecond)
	_, err = client.Explain(context.Background(), "explain", []RankItem{{ProgramID: "p", VersionID: "version"}})
	require.NoError(t, err)
}

func TestRankRejectsMissingOrNullRequiredFields(t *testing.T) {
	valid := `{"request_id":"r","model_version":"m","fallback":false,"items":[{"program_id":"p","version_id":"v","score":0,"reason_codes":[],"evidence":{"sessions_per_week":3,"level":"beginner","equipment_match":true,"goal_match":false}}]}`
	for _, field := range []string{`"fallback":false,`, `"score":0,`, `"reason_codes":[],`} {
		for _, null := range []bool{false, true} {
			body := strings.Replace(valid, field, "", 1)
			if null {
				body = strings.Replace(valid, field, strings.SplitN(field, ":", 2)[0]+":null,", 1)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(body)) }))
			_, err := NewV2Client(Config{BaseURL: server.URL, ServiceToken: "fixture"}).Rank(context.Background(), RankInput{RequestID: "r", Limit: 1, Candidates: []Candidate{{ProgramID: "p", VersionID: "v"}}})
			server.Close()
			require.Error(t, err)
		}
	}
}
