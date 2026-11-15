package mlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"sport_app/internal/core/transport/http/v2request"
	"time"
)

type RankingContext struct {
	Level           *string  `json:"level,omitempty"`
	Goal            *string  `json:"goal,omitempty"`
	SessionsPerWeek int      `json:"sessions_per_week"`
	EquipmentIDs    []string `json:"equipment_ids"`
	Progress        History  `json:"progress"`
	Behavior        Behavior `json:"behavior"`
}
type Rating struct {
	ProgramID string `json:"program_id"`
	Overall   int    `json:"overall"`
}
type Behavior struct {
	Viewed    []string `json:"viewed_program_ids"`
	Saved     []string `json:"saved_program_ids"`
	Purchased []string `json:"purchased_program_ids"`
	Ratings   []Rating `json:"ratings"`
}
type Candidate struct {
	ProgramID       string   `json:"program_id"`
	VersionID       string   `json:"version_id"`
	Level           string   `json:"level"`
	Goal            string   `json:"goal"`
	SessionsPerWeek int      `json:"sessions_per_week"`
	EquipmentIDs    []string `json:"equipment_ids"`
}
type Evidence struct {
	SessionsPerWeek int    `json:"sessions_per_week"`
	Level           string `json:"level"`
	EquipmentMatch  bool   `json:"equipment_match"`
	GoalMatch       bool   `json:"goal_match"`
}
type RankItem struct {
	ProgramID   string   `json:"program_id"`
	VersionID   string   `json:"version_id"`
	Score       float64  `json:"score"`
	ReasonCodes []string `json:"reason_codes"`
	Evidence    Evidence `json:"evidence"`
}
type RankInput struct {
	RequestID  string         `json:"request_id"`
	Schema     string         `json:"feature_schema_version"`
	Context    RankingContext `json:"context"`
	Candidates []Candidate    `json:"candidates"`
	Limit      int            `json:"limit"`
}
type RankResult struct {
	RequestID    string     `json:"request_id"`
	ModelVersion string     `json:"model_version"`
	Items        []RankItem `json:"items"`
	Fallback     bool       `json:"fallback"`
}
type Explanation struct {
	ProgramID      string   `json:"program_id"`
	Text           string   `json:"text"`
	SourceVersions []string `json:"source_version_ids"`
	Mode           string   `json:"explanation_mode"`
}
type ExplainResult struct {
	RequestID string        `json:"request_id"`
	Items     []Explanation `json:"items"`
}

func (c *V2Client) recommendationRequest(ctx context.Context, path, requestID string, input, result any) error {
	if c.baseURL == "" || c.token == "" {
		return Unavailable()
	}
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return Unavailable()
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("X-Request-ID", requestID)
	req.Header.Set("Content-Type", "application/json")
	response, err := c.client.Do(req)
	if err != nil {
		return Unavailable()
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != 200 {
		return Unavailable()
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, (1<<20)+1))
	if err != nil || len(raw) > 1<<20 {
		return Unavailable()
	}
	required := []string{"request_id", "items"}
	if _, ok := result.(*RankResult); ok {
		required = append(required, "model_version", "fallback")
	}
	fields, err := v2request.Object(raw, required...)
	if err != nil {
		return Unavailable()
	}
	if _, ok := result.(*RankResult); ok {
		var items []json.RawMessage
		if json.Unmarshal(fields["items"], &items) != nil {
			return Unavailable()
		}
		for _, item := range items {
			if _, err := v2request.Object(item, "program_id", "version_id", "score", "reason_codes", "evidence"); err != nil {
				return Unavailable()
			}
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(result) != nil {
		return Unavailable()
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return Unavailable()
	}
	return nil
}
func (c *V2Client) Rank(ctx context.Context, input RankInput) (RankResult, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var out RankResult
	if err := c.recommendationRequest(ctx, "/v2/recommendations/rank", input.RequestID, input, &out); err != nil {
		return out, err
	}
	if out.RequestID != input.RequestID || len(out.ModelVersion) == 0 || len(out.ModelVersion) > 200 || out.Items == nil || len(out.Items) > input.Limit {
		return out, Unavailable()
	}
	allowed := map[string]string{}
	for _, v := range input.Candidates {
		allowed[v.ProgramID] = v.VersionID
	}
	seen := map[string]bool{}
	for _, v := range out.Items {
		if allowed[v.ProgramID] != v.VersionID || v.VersionID == "" || seen[v.ProgramID] || math.IsNaN(v.Score) || math.IsInf(v.Score, 0) || len(v.ReasonCodes) > 20 {
			return out, Unavailable()
		}
		seen[v.ProgramID] = true
	}
	return out, nil
}
func (c *V2Client) Explain(ctx context.Context, requestID string, items []RankItem) (ExplainResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var out ExplainResult
	explainItems := []map[string]any{}
	for _, item := range items {
		explainItems = append(explainItems, map[string]any{"program_id": item.ProgramID, "version_id": item.VersionID, "reason_codes": item.ReasonCodes, "evidence": item.Evidence})
	}
	input := map[string]any{"request_id": requestID, "locale": "ru-RU", "items": explainItems}
	if err := c.recommendationRequest(ctx, "/v2/recommendations/explain", requestID, input, &out); err != nil {
		return out, err
	}
	if out.RequestID != requestID || out.Items == nil {
		return out, Unavailable()
	}
	return out, nil
}
