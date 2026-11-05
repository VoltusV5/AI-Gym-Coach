package v2request

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	core_errors "sport_app/internal/core/errors"
)

func Invalid() error { return fmt.Errorf("invalid input: %w", core_errors.ErrInvalidArgument) }
func UserID(raw string) (int, error) {
	id, err := strconv.Atoi(raw)
	if err != nil || id < 1 || id > math.MaxInt32 {
		return 0, fmt.Errorf("invalid identity: %w", core_errors.ErrUnauthorized)
	}
	return id, nil
}
func UUID(raw string) (string, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return "", Invalid()
	}
	return id.String(), nil
}
func Text(value string, max int, required bool) bool {
	return (!required || strings.TrimSpace(value) != "") && !strings.ContainsRune(value, '\x00') && len([]rune(value)) <= max
}
func Decode(w http.ResponseWriter, r *http.Request, dest any, required ...string) (map[string]json.RawMessage, error) {
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		return nil, Invalid()
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(dest) != nil {
		return nil, Invalid()
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return nil, Invalid()
	}
	fields, err := Object(data, required...)
	if err != nil {
		return nil, err
	}
	return fields, validateNested(data, "")
}
func Object(data []byte, required ...string) (map[string]json.RawMessage, error) {
	var m map[string]json.RawMessage
	if json.Unmarshal(data, &m) != nil || m == nil {
		return nil, Invalid()
	}
	for _, k := range required {
		if _, ok := m[k]; !ok {
			return nil, Invalid()
		}
	}
	for k, v := range m {
		if k != "weight_kg" && k != "photo_ref" && bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
			return nil, Invalid()
		}
	}
	return m, nil
}
func validateNested(data []byte, key string) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return Invalid()
	}
	switch data[0] {
	case '{':
		required := map[string][]string{"days": {"title", "position", "exercises"}, "exercises": {"exercise_id", "position", "sets", "reps_min", "reps_max", "rest_sec"}, "slots": {"exercise_id", "position", "sets"}, "sets": {"reps", "weight_kg", "is_warmup"}, "entries": {"weekday", "plan_day_id"}}[key]
		m, err := Object(data, required...)
		if err != nil {
			return err
		}
		if key == "sets" && bytes.Equal(bytes.TrimSpace(m["weight_kg"]), []byte("null")) {
			return Invalid()
		}
		for k, v := range m {
			if err := validateNested(v, k); err != nil {
				return err
			}
		}
	case '[':
		var items []json.RawMessage
		if json.Unmarshal(data, &items) != nil {
			return Invalid()
		}
		for _, v := range items {
			if err := validateNested(v, key); err != nil {
				return err
			}
		}
	}
	return nil
}

type PageFilter struct {
	Limit int
	After string
	Scope string
}

func Pagination(r *http.Request, scope string) (PageFilter, error) {
	f := PageFilter{Limit: 20, Scope: scope}
	if r.URL.Query().Has("limit") {
		var err error
		f.Limit, err = strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil {
			return f, Invalid()
		}
	}
	if f.Limit < 1 || f.Limit > 100 {
		return f, Invalid()
	}
	return PageValues(f.Limit, r.URL.Query().Get("cursor"), scope)
}
func PageValues(limit int, raw, scope string) (PageFilter, error) {
	if limit == 0 {
		limit = 20
	}
	f := PageFilter{Limit: limit, Scope: scope}
	if limit < 1 || limit > 100 {
		return f, Invalid()
	}

	if raw != "" {
		var c struct {
			Scope string
			After string
		}
		data, err := base64.RawURLEncoding.DecodeString(raw)
		if len(raw) > 2048 || err != nil || json.Unmarshal(data, &c) != nil || c.Scope != scope || c.After == "" {
			return f, Invalid()
		}
		f.After = c.After
	}
	return f, nil
}
func Next(f PageFilter, after string) *string {
	data, _ := json.Marshal(struct {
		Scope string
		After string
	}{f.Scope, after})
	s := base64.RawURLEncoding.EncodeToString(data)
	return &s
}
