package plans_service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	core_errors "sport_app/internal/core/errors"
	plans_postgres_repository "sport_app/internal/features/plans/repository/postgres"

	"github.com/google/uuid"
)

type Repository interface {
	ListPlans(context.Context, int, plans_postgres_repository.Filter) ([]plans_postgres_repository.Plan, error)
	GetPlan(context.Context, int, string) (plans_postgres_repository.Plan, error)
	CreatePlan(context.Context, int, plans_postgres_repository.Input) (plans_postgres_repository.Plan, error)
	UpdatePlan(context.Context, int, string, plans_postgres_repository.Input) (plans_postgres_repository.Plan, error)
	ActivatePlan(context.Context, int, string, int64) (plans_postgres_repository.Plan, error)
	ArchivePlan(context.Context, int, string) error
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func userID(raw string) (int, error) {
	id, err := strconv.Atoi(raw)
	if err != nil || id < 1 || id > math.MaxInt32 {
		return 0, fmt.Errorf("invalid plan user: %w", core_errors.ErrUnauthorized)
	}
	return id, nil
}

func planID(raw string) (string, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid plan id: %w", core_errors.ErrInvalidArgument)
	}
	return id.String(), nil
}

func validTitle(title string) bool {
	return strings.TrimSpace(title) != "" && !strings.ContainsRune(title, '\x00') && utf8.RuneCountInString(title) <= 200
}

func prepareInput(input plans_postgres_repository.Input, update bool) (plans_postgres_repository.Input, error) {
	invalid := fmt.Errorf("invalid personal plan: %w", core_errors.ErrInvalidArgument)
	if !validTitle(input.Title) || input.Days == nil || input.Schedule.Entries == nil {
		return input, invalid
	}
	if update && (input.Version < 1 || input.Version == math.MaxInt64) {
		return input, invalid
	}
	if input.Schedule.TimeZone == "" || input.Schedule.TimeZone == "Local" {
		return input, invalid
	}
	if _, err := time.LoadLocation(input.Schedule.TimeZone); err != nil {
		return input, invalid
	}
	date, err := time.Parse(time.DateOnly, input.Schedule.StartDate)
	if err != nil || date.Year() < 1 {
		return input, invalid
	}
	input.Days = slices.Clone(input.Days)
	dayIDs := map[string]bool{}
	positions := map[int]bool{}
	for i, day := range input.Days {
		if day.ID == "" {
			day.ID = uuid.NewString()
		} else {
			day.ID, err = planID(day.ID)
			if err != nil {
				return input, err
			}
		}
		if dayIDs[day.ID] || positions[day.Position] || !validTitle(day.Title) || day.Position < 0 || day.Position > math.MaxInt32 || day.Exercises == nil {
			return input, invalid
		}
		dayIDs[day.ID], positions[day.Position] = true, true
		day.Exercises = slices.Clone(day.Exercises)
		exPositions := map[int]bool{}
		for _, ex := range day.Exercises {
			if ex.ExerciseID < 1 || ex.ExerciseID > math.MaxInt32 || ex.Position < 0 || ex.Position > math.MaxInt32 || exPositions[ex.Position] ||
				ex.Sets < 1 || ex.Sets > math.MaxInt32 || ex.RepsMin < 1 || ex.RepsMax < ex.RepsMin || ex.RepsMax > math.MaxInt32 || ex.RestSec < 0 || ex.RestSec > math.MaxInt32 {
				return input, invalid
			}
			if ex.WeightKg != nil && (math.IsNaN(*ex.WeightKg) || math.IsInf(*ex.WeightKg, 0) || *ex.WeightKg < 0) {
				return input, invalid
			}
			exPositions[ex.Position] = true
		}
		input.Days[i] = day
	}
	input.Schedule.Entries = slices.Clone(input.Schedule.Entries)
	weekdays := map[int]bool{}
	for i, entry := range input.Schedule.Entries {
		id, err := planID(entry.PlanDayID)
		if err != nil || !dayIDs[id] || entry.Weekday < 1 || entry.Weekday > 7 || weekdays[entry.Weekday] {
			return input, invalid
		}
		weekdays[entry.Weekday] = true
		input.Schedule.Entries[i].PlanDayID = id
	}
	return input, nil
}

type cursor struct {
	OwnerID   int       `json:"owner_id"`
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Service) ListPlans(ctx context.Context, rawID string, limit int, rawCursor string) (plans_postgres_repository.Page, error) {
	id, err := userID(rawID)
	if err != nil {
		return plans_postgres_repository.Page{}, err
	}
	if limit == 0 {
		limit = 20
	}
	if limit < 1 || limit > 100 || len(rawCursor) > 1024 {
		return plans_postgres_repository.Page{}, fmt.Errorf("invalid plan page: %w", core_errors.ErrInvalidArgument)
	}
	filter := plans_postgres_repository.Filter{Limit: limit}
	if rawCursor != "" {
		data, err := base64.RawURLEncoding.DecodeString(rawCursor)
		var c cursor
		if err != nil || json.Unmarshal(data, &c) != nil || c.OwnerID != id || c.CreatedAt.IsZero() {
			return plans_postgres_repository.Page{}, fmt.Errorf("invalid plan cursor: %w", core_errors.ErrInvalidArgument)
		}
		filter.AfterID, err = planID(c.ID)
		if err != nil {
			return plans_postgres_repository.Page{}, err
		}
		filter.AfterCreatedAt = c.CreatedAt
	}
	items, err := s.repo.ListPlans(ctx, id, filter)
	if err != nil {
		return plans_postgres_repository.Page{}, err
	}
	if items == nil {
		items = []plans_postgres_repository.Plan{}
	}
	page := plans_postgres_repository.Page{Items: items}
	if len(items) > limit {
		page.Items = items[:limit]
		last := items[limit-1]
		raw, _ := json.Marshal(cursor{OwnerID: id, ID: last.ID, CreatedAt: last.CreatedAt})
		next := base64.RawURLEncoding.EncodeToString(raw)
		page.NextCursor = &next
	}
	return page, nil
}

func (s *Service) GetPlan(ctx context.Context, rawID, rawPlanID string) (plans_postgres_repository.Plan, error) {
	id, err := userID(rawID)
	if err != nil {
		return plans_postgres_repository.Plan{}, err
	}
	pid, err := planID(rawPlanID)
	if err != nil {
		return plans_postgres_repository.Plan{}, err
	}
	return s.repo.GetPlan(ctx, id, pid)
}

func (s *Service) CreatePlan(ctx context.Context, rawID string, input plans_postgres_repository.Input) (plans_postgres_repository.Plan, error) {
	id, err := userID(rawID)
	if err != nil {
		return plans_postgres_repository.Plan{}, err
	}
	input, err = prepareInput(input, false)
	if err != nil {
		return plans_postgres_repository.Plan{}, err
	}
	return s.repo.CreatePlan(ctx, id, input)
}

func (s *Service) UpdatePlan(ctx context.Context, rawID, rawPlanID string, input plans_postgres_repository.Input) (plans_postgres_repository.Plan, error) {
	id, err := userID(rawID)
	if err != nil {
		return plans_postgres_repository.Plan{}, err
	}
	pid, err := planID(rawPlanID)
	if err != nil {
		return plans_postgres_repository.Plan{}, err
	}
	input, err = prepareInput(input, true)
	if err != nil {
		return plans_postgres_repository.Plan{}, err
	}
	return s.repo.UpdatePlan(ctx, id, pid, input)
}

func (s *Service) ActivatePlan(ctx context.Context, rawID, rawPlanID string, version int64) (plans_postgres_repository.Plan, error) {
	id, err := userID(rawID)
	if err != nil {
		return plans_postgres_repository.Plan{}, err
	}
	pid, err := planID(rawPlanID)
	if err != nil {
		return plans_postgres_repository.Plan{}, err
	}
	if version < 1 || version == math.MaxInt64 {
		return plans_postgres_repository.Plan{}, fmt.Errorf("invalid plan version: %w", core_errors.ErrInvalidArgument)
	}
	return s.repo.ActivatePlan(ctx, id, pid, version)
}

func (s *Service) ArchivePlan(ctx context.Context, rawID, rawPlanID string) error {
	id, err := userID(rawID)
	if err != nil {
		return err
	}
	pid, err := planID(rawPlanID)
	if err != nil {
		return err
	}
	return s.repo.ArchivePlan(ctx, id, pid)
}
