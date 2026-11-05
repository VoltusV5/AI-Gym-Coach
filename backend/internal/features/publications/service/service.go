package publications_service

import (
	"context"
	"encoding/json"
	"math"
	"regexp"
	"strings"

	"sport_app/internal/core/transport/http/v2request"
	plans_postgres_repository "sport_app/internal/features/plans/repository/postgres"
	plans_service "sport_app/internal/features/plans/service"
	publications_postgres_repository "sport_app/internal/features/publications/repository/postgres"
	trainers_service "sport_app/internal/features/trainers/service"
)

type Repository interface {
	Get(context.Context, int, string) (publications_postgres_repository.Program, error)
	List(context.Context, int, bool, publications_postgres_repository.Filter) (publications_postgres_repository.Page, error)
	Save(context.Context, int, string, publications_postgres_repository.Draft) (publications_postgres_repository.Program, error)
	NewVersion(context.Context, int, string) (publications_postgres_repository.Program, error)
	Publish(context.Context, int, string, int64) (publications_postgres_repository.Workflow, error)
	Moderate(context.Context, int, string, publications_postgres_repository.Moderation, string) (publications_postgres_repository.Workflow, error)
}
type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

var currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)

func prepare(d publications_postgres_repository.Draft) (publications_postgres_repository.Draft, error) {
	p, err := plans_service.PrepareInput(plans_postgres_repository.Input{Title: d.Title, Days: d.Days, Schedule: plans_postgres_repository.Schedule{TimeZone: "UTC", StartDate: "2026-01-01", Entries: []plans_postgres_repository.ScheduleEntry{}}}, false)
	if err != nil {
		return d, err
	}
	d.Days = p.Days
	if !v2request.Text(d.Description, 50000, false) || !v2request.Text(d.Goal, 200, true) || d.DurationWeeks < 1 || d.DurationWeeks > math.MaxInt32 || d.SessionsPerWeek < 1 || d.SessionsPerWeek > 7 || d.PriceMinor < 0 || !currencyPattern.MatchString(d.Currency) || d.EquipmentIDs == nil {
		return d, v2request.Invalid()
	}
	if d.Level != "beginner" && d.Level != "intermediate" && d.Level != "advanced" {
		return d, v2request.Invalid()
	}
	if d.CategoryIDs == nil {
		d.CategoryIDs = []string{}
	}
	seen := map[string]bool{}
	for _, id := range d.EquipmentIDs {
		if !v2request.Text(id, 100, true) || seen[id] {
			return d, v2request.Invalid()
		}
		seen[id] = true
	}
	seen = map[string]bool{}
	for i, id := range d.CategoryIDs {
		u, err := v2request.UUID(id)
		if err != nil || seen[u] {
			return d, v2request.Invalid()
		}
		seen[u] = true
		d.CategoryIDs[i] = u
	}
	return d, nil
}
func (s *Service) Get(ctx context.Context, raw, target string) (publications_postgres_repository.Program, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return publications_postgres_repository.Program{}, err
	}
	target, err = v2request.UUID(target)
	if err != nil {
		return publications_postgres_repository.Program{}, err
	}
	return s.repo.Get(ctx, id, target)
}
func (s *Service) Create(ctx context.Context, raw string, d publications_postgres_repository.Draft) (publications_postgres_repository.Program, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return publications_postgres_repository.Program{}, err
	}
	d, err = prepare(d)
	if err != nil {
		return publications_postgres_repository.Program{}, err
	}
	return s.repo.Save(ctx, id, "", d)
}
func (s *Service) Update(ctx context.Context, raw, target string, fields map[string]json.RawMessage) (publications_postgres_repository.Program, error) {
	current, err := s.Get(ctx, raw, target)
	if err != nil {
		return current, err
	}
	rawDraft, _ := json.Marshal(current.Draft)
	var merged map[string]json.RawMessage
	_ = json.Unmarshal(rawDraft, &merged)
	for k, v := range fields {
		merged[k] = v
	}
	rawDraft, _ = json.Marshal(merged)
	var d publications_postgres_repository.Draft
	if json.Unmarshal(rawDraft, &d) != nil || d.Version < 1 || d.Version == math.MaxInt64 {
		return current, v2request.Invalid()
	}
	d, err = prepare(d)
	if err != nil {
		return current, err
	}
	id, _ := v2request.UserID(raw)
	return s.repo.Save(ctx, id, current.ID, d)
}
func (s *Service) NewVersion(ctx context.Context, raw, target string) (publications_postgres_repository.Program, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return publications_postgres_repository.Program{}, err
	}
	target, err = v2request.UUID(target)
	if err != nil {
		return publications_postgres_repository.Program{}, err
	}
	return s.repo.NewVersion(ctx, id, target)
}
func (s *Service) List(ctx context.Context, raw string, admin bool, f v2request.PageFilter) (publications_postgres_repository.Page, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return publications_postgres_repository.Page{}, err
	}
	if err := trainers_service.ValidatePage(f, false); err != nil {
		return publications_postgres_repository.Page{}, err
	}
	page, err := s.repo.List(ctx, id, admin, publications_postgres_repository.Filter{Limit: f.Limit, After: f.After})
	if err == nil && len(page.Items) > f.Limit {
		page.Items = page.Items[:f.Limit]
		page.NextCursor = v2request.Next(f, page.Items[f.Limit-1].ID)
	}
	return page, err
}
func (s *Service) Publish(ctx context.Context, raw, target string, version int64) (publications_postgres_repository.Workflow, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return publications_postgres_repository.Workflow{}, err
	}
	target, err = v2request.UUID(target)
	if err != nil {
		return publications_postgres_repository.Workflow{}, err
	}
	if version < 1 || version == math.MaxInt64 {
		return publications_postgres_repository.Workflow{}, v2request.Invalid()
	}
	return s.repo.Publish(ctx, id, target, version)
}
func (s *Service) Moderate(ctx context.Context, raw, target string, m publications_postgres_repository.Moderation, requestID string) (publications_postgres_repository.Workflow, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return publications_postgres_repository.Workflow{}, err
	}
	target, err = v2request.UUID(target)
	if err != nil {
		return publications_postgres_repository.Workflow{}, err
	}
	if m.ProgramVersionID != "" {
		m.ProgramVersionID, err = v2request.UUID(m.ProgramVersionID)
		if err != nil {
			return publications_postgres_repository.Workflow{}, err
		}
	}
	if !v2request.Text(m.Reason, 4000, true) || !strings.Contains("|approved|published|rejected|", "|"+m.Status+"|") {
		return publications_postgres_repository.Workflow{}, v2request.Invalid()
	}
	return s.repo.Moderate(ctx, id, target, m, requestID)
}
