package v2request

import (
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStrictNestedJSON(t *testing.T) {
	type set struct {
		Reps     int     `json:"reps"`
		WeightKg float64 `json:"weight_kg"`
		IsWarmup bool    `json:"is_warmup"`
	}
	type body struct {
		Slots []struct {
			ExerciseID int   `json:"exercise_id"`
			Position   int   `json:"position"`
			Sets       []set `json:"sets"`
		} `json:"slots"`
	}
	for _, raw := range []string{`{"slots":[{"exercise_id":1,"position":0,"sets":[{"reps":10,"weight_kg":null,"is_warmup":false}]}]}`, `{"slots":[{"exercise_id":1,"position":0,"sets":[{"reps":10,"weight_kg":30}]}]}`, `{"slots":null}`, `{"slots":[],"owner_id":2}`, `{"slots":[]} {}`} {
		var b body
		_, err := Decode(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(raw)), &b, "slots")
		require.Error(t, err, raw)
	}
	var b body
	_, err := Decode(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(`{"slots":[{"exercise_id":1,"position":0,"sets":[{"reps":10,"weight_kg":0,"is_warmup":false}]}]}`)), &b, "slots")
	require.NoError(t, err)
}
func TestPageScope(t *testing.T) {
	f := PageFilter{Scope: "owner1", Limit: 20}
	cursor := Next(f, "42")
	_, err := Pagination(httptest.NewRequest("GET", "/?cursor="+*cursor, nil), "owner2")
	require.Error(t, err)
}
