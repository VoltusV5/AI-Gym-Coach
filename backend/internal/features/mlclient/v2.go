package mlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	core_errors "sport_app/internal/core/errors"
	"sport_app/internal/core/transport/http/v2request"
	plans_postgres_repository "sport_app/internal/features/plans/repository/postgres"
	profile_postgres_repository "sport_app/internal/features/profile/repository/postgres"
)

type HistoryMetric struct {
	ExerciseID int       `json:"exercise_id"`
	WeightKg   float64   `json:"weight_kg"`
	Reps       int       `json:"reps"`
	VolumeKg   float64   `json:"volume_kg"`
	MeasuredAt time.Time `json:"measured_at"`
}
type History struct {
	Completed28d    int             `json:"completed_28d"`
	Adherence28d    *float64        `json:"adherence_28d"`
	ExerciseMetrics []HistoryMetric `json:"exercise_metrics"`
}
type GenerateInput struct {
	RequestID            string                              `json:"request_id"`
	FeatureSchemaVersion string                              `json:"feature_schema_version"`
	Profile              profile_postgres_repository.Profile `json:"profile"`
	History              *History                            `json:"history"`
}
type GeneratedDay struct {
	Weekday   int                                      `json:"weekday"`
	Title     string                                   `json:"title"`
	Exercises []plans_postgres_repository.Prescription `json:"exercises"`
}
type GenerateResult struct {
	RequestID            string         `json:"request_id"`
	ModelVersion         string         `json:"model_version"`
	FeatureSchemaVersion string         `json:"feature_schema_version,omitempty"`
	Split                string         `json:"split"`
	Days                 []GeneratedDay `json:"days"`
	Warnings             []string       `json:"warnings"`
	Fallback             bool           `json:"fallback"`
}
type V2Client struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewV2Client(config Config) *V2Client {
	return &V2Client{baseURL: strings.TrimRight(config.BaseURL, "/"), token: config.ServiceToken, client: &http.Client{Timeout: 8 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}
func Unavailable() error {
	return &core_errors.DomainError{Code: "ML_UNAVAILABLE", Message: "Plan generator unavailable", Status: 503}
}
func NoFeasiblePlan() error {
	return &core_errors.DomainError{Code: "NO_FEASIBLE_PLAN", Message: "No valid plan satisfies the constraints", Status: 422}
}
func (c *V2Client) Generate(ctx context.Context, input GenerateInput) (GenerateResult, error) {
	var result GenerateResult
	if c.baseURL == "" || c.token == "" {
		return result, Unavailable()
	}
	profile := map[string]any{"age": input.Profile.Age, "height_cm": input.Profile.HeightCm, "weight_kg": input.Profile.WeightKg, "level": input.Profile.Level, "goal": input.Profile.Goal, "training_days": input.Profile.TrainingDays, "equipment_ids": input.Profile.EquipmentIDs, "excluded_exercise_ids": input.Profile.ExcludedExerciseIDs, "session_duration_min": input.Profile.SessionDurationMin, "gender": input.Profile.Gender, "activity_level": input.Profile.ActivityLevel, "injuries_notes": input.Profile.InjuriesNotes}
	data, err := json.Marshal(map[string]any{"request_id": input.RequestID, "feature_schema_version": input.FeatureSchemaVersion, "profile": profile, "history": input.History})
	if err != nil {
		return result, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v2/plans/generate", bytes.NewReader(data))
	if err != nil {
		return result, Unavailable()
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", input.RequestID)
	req.Header.Set("Authorization", "Bearer "+c.token)
	response, err := c.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		var timeout interface{ Timeout() bool }
		if errors.As(err, &timeout) && timeout.Timeout() {
			return result, context.DeadlineExceeded
		}
		return result, Unavailable()
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode == 422 {
		return result, NoFeasiblePlan()
	}
	if response.StatusCode != 200 {
		return result, Unavailable()
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, (1<<20)+1))
	if err != nil {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		var timeout interface{ Timeout() bool }
		if errors.As(err, &timeout) && timeout.Timeout() {
			return result, context.DeadlineExceeded
		}
		return result, Unavailable()
	}
	if len(raw) > 1<<20 {
		return result, Unavailable()
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&result) != nil {
		return result, NoFeasiblePlan()
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return result, NoFeasiblePlan()
	}
	fields, err := v2request.Object(raw, "request_id", "model_version", "split", "days", "fallback")
	if err != nil {
		return result, NoFeasiblePlan()
	}
	var days []json.RawMessage
	if json.Unmarshal(fields["days"], &days) != nil {
		return result, NoFeasiblePlan()
	}
	for _, day := range days {
		d, err := v2request.Object(day, "weekday", "title", "exercises")
		if err != nil {
			return result, NoFeasiblePlan()
		}
		var exercises []json.RawMessage
		if json.Unmarshal(d["exercises"], &exercises) != nil {
			return result, NoFeasiblePlan()
		}
		for _, ex := range exercises {
			if _, err := v2request.Object(ex, "exercise_id", "position", "sets", "reps_min", "reps_max", "rest_sec"); err != nil {
				return result, NoFeasiblePlan()
			}
		}
	}
	if result.RequestID != input.RequestID || !v2request.Text(result.ModelVersion, 200, true) || !v2request.Text(result.Split, 200, true) || (result.FeatureSchemaVersion != "" && result.FeatureSchemaVersion != input.FeatureSchemaVersion) {
		return result, NoFeasiblePlan()
	}
	if result.Warnings == nil {
		result.Warnings = []string{}
	}
	for _, warning := range result.Warnings {
		if !v2request.Text(warning, 200, true) {
			return result, fmt.Errorf("invalid warning: %w", NoFeasiblePlan())
		}
	}
	return result, nil
}
