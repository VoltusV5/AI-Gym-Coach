package insights_service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"math"
	"slices"
	"sort"
	"sport_app/internal/core/transport/http/v2request"
	catalog "sport_app/internal/features/catalog/repository/postgres"
	repository "sport_app/internal/features/insights/repository/postgres"
	"sport_app/internal/features/mlclient"
	"strings"
)

type Repository interface {
	Snapshot(context.Context, int) (repository.Snapshot, error)
	Visible(context.Context, []string) (map[string]string, error)
	Analytics(context.Context, int, bool, Period) (map[string]any, error)
	Dashboard(context.Context, int) (repository.DashboardSnapshot, error)
}
type Ranker interface {
	Rank(context.Context, mlclient.RankInput) (mlclient.RankResult, error)
	Explain(context.Context, string, []mlclient.RankItem) (mlclient.ExplainResult, error)
}
type Service struct {
	repo         Repository
	ml           Ranker
	inactiveDays int
}

func NewService(repo Repository, ml Ranker, inactiveDays int) *Service {
	if inactiveDays < 1 {
		inactiveDays = 14
	}
	return &Service{repo, ml, inactiveDays}
}

type Recommendation struct {
	mlclient.RankItem
	Program         catalog.ProgramCard `json:"program"`
	Explanation     string              `json:"explanation"`
	ExplanationMode string              `json:"explanation_mode"`
}
type RecommendationPage struct {
	Items        []Recommendation `json:"items"`
	ModelVersion string           `json:"model_version"`
	Fallback     bool             `json:"fallback"`
	NextCursor   *string          `json:"next_cursor"`
	RequestID    string           `json:"request_id"`
}

func facts(card catalog.ProgramCard, ctx mlclient.RankingContext) (mlclient.Evidence, []string) {
	e := mlclient.Evidence{Level: card.Level, SessionsPerWeek: card.SessionsPerWeek, EquipmentMatch: true, GoalMatch: ctx.Goal != nil && *ctx.Goal == card.Goal}
	codes := []string{"EQUIPMENT_MATCH"}
	if e.GoalMatch {
		codes = append(codes, "GOAL_MATCH")
	}
	if ctx.Level != nil && *ctx.Level == card.Level {
		codes = append(codes, "LEVEL_MATCH")
	}
	if ctx.SessionsPerWeek > 0 && ctx.SessionsPerWeek == card.SessionsPerWeek {
		codes = append(codes, "FREQUENCY_MATCH")
	}
	return e, codes
}
func explanation(e mlclient.Evidence, codes []string) string {
	parts := []string{fmt.Sprintf("Занятий в неделю: %d", e.SessionsPerWeek)}
	if slices.Contains(codes, "LEVEL_MATCH") {
		parts = append(parts, "уровень соответствует вашему профилю")
	}
	if slices.Contains(codes, "GOAL_MATCH") {
		parts = append(parts, "цель совпадает с вашей")
	}
	if e.EquipmentMatch {
		parts = append(parts, "подходит доступное оборудование")
	}
	return strings.Join(parts, "; ") + "."
}
func (s *Service) Recommendations(ctx context.Context, raw string, limit int) (RecommendationPage, error) {
	out := RecommendationPage{Items: []Recommendation{}, ModelVersion: "rules-v1", Fallback: true, RequestID: uuid.NewString()}
	user, err := v2request.UserID(raw)
	if err != nil {
		return out, err
	}
	if limit < 1 || limit > 100 {
		return out, v2request.Invalid()
	}
	snapshot, err := s.repo.Snapshot(ctx, user)
	if err != nil {
		return out, err
	}
	sort.Slice(snapshot.Context.Progress.ExerciseMetrics, func(i, j int) bool {
		return snapshot.Context.Progress.ExerciseMetrics[i].ExerciseID < snapshot.Context.Progress.ExerciseMetrics[j].ExerciseID
	})
	cards := map[string]catalog.ProgramCard{}
	candidates := []mlclient.Candidate{}
	ranked := []mlclient.RankItem{}
	for _, card := range snapshot.Cards {
		cards[card.ID] = card
		equipment := []string{}
		for _, e := range card.Equipment {
			equipment = append(equipment, e.ID)
		}
		candidates = append(candidates, mlclient.Candidate{ProgramID: card.ID, VersionID: card.PublishedVersionID, Level: card.Level, Goal: card.Goal, SessionsPerWeek: card.SessionsPerWeek, EquipmentIDs: equipment})
		e, codes := facts(card, snapshot.Context)
		ranked = append(ranked, mlclient.RankItem{ProgramID: card.ID, VersionID: card.PublishedVersionID, Score: float64(len(codes)), ReasonCodes: codes, Evidence: e})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Score == ranked[j].Score {
			return ranked[i].ProgramID < ranked[j].ProgramID
		}
		return ranked[i].Score > ranked[j].Score
	})
	if len(ranked) > limit {
		ranked = ranked[:limit]
	}
	if len(candidates) > 0 && s.ml != nil {
		result, err := s.ml.Rank(ctx, mlclient.RankInput{RequestID: out.RequestID, Schema: "ranking-v1", Context: snapshot.Context, Candidates: candidates, Limit: limit})
		valid := err == nil && result.RequestID == out.RequestID && len(result.Items) > 0 && len(result.Items) <= limit && v2request.Text(result.ModelVersion, 200, true)
		seen := map[string]bool{}
		for _, item := range result.Items {
			card, ok := cards[item.ProgramID]
			e, codes := facts(card, snapshot.Context)
			if !ok || seen[item.ProgramID] || item.VersionID != card.PublishedVersionID || item.Evidence != e || math.IsNaN(item.Score) || math.IsInf(item.Score, 0) {
				valid = false
			}
			seen[item.ProgramID] = true
			for _, code := range item.ReasonCodes {
				if !slices.Contains(codes, code) {
					valid = false
				}
			}
		}
		if valid {
			ranked = result.Items
			out.ModelVersion = result.ModelVersion
			out.Fallback = result.Fallback
		}
	}
	explanations := map[string]string{}
	if len(ranked) > 0 && s.ml != nil {
		if response, err := s.ml.Explain(ctx, out.RequestID, ranked); err == nil && response.RequestID == out.RequestID {
			for _, v := range response.Items {
				card, ok := cards[v.ProgramID]
				if !ok || len(v.SourceVersions) != 1 || v.SourceVersions[0] != card.PublishedVersionID {
					continue
				}
				for _, item := range ranked {
					if item.ProgramID == v.ProgramID && v.Text == explanation(item.Evidence, item.ReasonCodes) {
						explanations[v.ProgramID] = v.Text
					}
				}
			}
		}
	}
	ids := []string{}
	for _, v := range ranked {
		ids = append(ids, v.ProgramID)
	}
	visible, err := s.repo.Visible(ctx, ids)
	if err != nil {
		return out, err
	}
	for _, v := range ranked {
		if visible[v.ProgramID] != v.VersionID {
			continue
		}
		text := explanation(v.Evidence, v.ReasonCodes)
		if explanations[v.ProgramID] != "" {
			text = explanations[v.ProgramID]
		}
		out.Items = append(out.Items, Recommendation{RankItem: v, Program: cards[v.ProgramID], Explanation: text, ExplanationMode: "template"})
	}
	return out, nil
}
