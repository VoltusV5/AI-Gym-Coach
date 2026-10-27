package catalog_service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	core_errors "sport_app/internal/core/errors"
	catalog_postgres_repository "sport_app/internal/features/catalog/repository/postgres"

	"github.com/google/uuid"
)

type Repository interface {
	ListEquipment(context.Context, catalog_postgres_repository.Filter) ([]catalog_postgres_repository.Equipment, error)
	ListExercises(context.Context, catalog_postgres_repository.Filter) ([]catalog_postgres_repository.Exercise, error)
	EquipmentExists(context.Context, []string) (bool, error)
	ListPrograms(context.Context, catalog_postgres_repository.Filter) ([]catalog_postgres_repository.ProgramCard, error)
	GetProgram(context.Context, string, int) (catalog_postgres_repository.ProgramDetail, error)
	GrantFreeAccess(context.Context, string, int) (catalog_postgres_repository.Entitlement, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type cursor struct {
	Kind       string     `json:"kind"`
	After      string     `json:"after"`
	Scope      string     `json:"scope"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	PriceMinor *int64     `json:"price_minor,omitempty"`
}

func prepareFilter(filter catalog_postgres_repository.Filter, raw, kind string) (catalog_postgres_repository.Filter, string, error) {
	if filter.Limit == 0 {
		filter.Limit = 20
	}
	if filter.Limit < 1 || filter.Limit > 100 || filter.DurationWeeks < 0 || filter.DurationWeeks > math.MaxInt32 || filter.TrainerID < 0 || filter.TrainerID > math.MaxInt32 {
		return filter, "", fmt.Errorf("invalid filter range: %w", core_errors.ErrInvalidArgument)
	}
	if !slices.Contains([]string{"", "newest", "price_asc", "price_desc"}, filter.Sort) ||
		!slices.Contains([]string{"", "free", "paid"}, filter.PriceType) {
		return filter, "", fmt.Errorf("invalid sort or price type: %w", core_errors.ErrInvalidArgument)
	}
	filter.After = ""
	filter.AfterCreatedAt = time.Time{}
	filter.AfterPrice = 0
	if kind == "programs" && filter.Sort == "" {
		filter.Sort = "newest"
	}
	filter.Equipment = slices.Clone(filter.Equipment)
	slices.Sort(filter.Equipment)
	filter.Equipment = slices.Compact(filter.Equipment)
	for _, id := range filter.Equipment {
		if strings.TrimSpace(id) == "" {
			return filter, "", fmt.Errorf("empty equipment id: %w", core_errors.ErrInvalidArgument)
		}
	}
	scopeFilter := filter
	scopeFilter.Limit = 0
	scopeRaw, _ := json.Marshal(scopeFilter)
	scope := fmt.Sprintf("%x", sha256.Sum256(scopeRaw))
	if raw == "" {
		return filter, scope, nil
	}
	if len(raw) > 1024 {
		return filter, "", fmt.Errorf("invalid cursor: %w", core_errors.ErrInvalidArgument)
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	var c cursor
	if err != nil || json.Unmarshal(decoded, &c) != nil || c.Kind != kind || c.Scope != scope || c.After == "" {
		return filter, "", fmt.Errorf("invalid cursor: %w", core_errors.ErrInvalidArgument)
	}
	switch kind {
	case "programs":
		id, err := uuid.Parse(c.After)
		if err != nil {
			return filter, "", fmt.Errorf("invalid program cursor: %w", core_errors.ErrInvalidArgument)
		}
		c.After = id.String()
		if filter.Sort == "newest" {
			if c.CreatedAt == nil || c.CreatedAt.IsZero() {
				return filter, "", fmt.Errorf("missing cursor date: %w", core_errors.ErrInvalidArgument)
			}
			filter.AfterCreatedAt = *c.CreatedAt
		} else {
			if c.PriceMinor == nil || *c.PriceMinor < 0 {
				return filter, "", fmt.Errorf("missing cursor price: %w", core_errors.ErrInvalidArgument)
			}
			filter.AfterPrice = *c.PriceMinor
		}
	case "exercises":
		if id, err := strconv.Atoi(c.After); err != nil || id <= 0 || id > math.MaxInt32 {
			return filter, "", fmt.Errorf("invalid exercise cursor: %w", core_errors.ErrInvalidArgument)
		}
	}
	filter.After = c.After
	return filter, scope, nil
}

func makePage[T any](items []T, limit int, kind, scope string, key func(T) cursor) catalog_postgres_repository.Page[T] {
	page := catalog_postgres_repository.Page[T]{Items: items}
	if page.Items == nil {
		page.Items = []T{}
	}
	if len(items) > limit {
		page.Items = items[:limit]
		c := key(items[limit-1])
		c.Kind, c.Scope = kind, scope
		raw, _ := json.Marshal(c)
		next := base64.RawURLEncoding.EncodeToString(raw)
		page.NextCursor = &next
	}
	return page
}

func (s *Service) validateEquipment(ctx context.Context, filter catalog_postgres_repository.Filter) error {
	if len(filter.Equipment) == 0 {
		return nil
	}
	exists, err := s.repo.EquipmentExists(ctx, filter.Equipment)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("unknown equipment: %w", core_errors.ErrInvalidArgument)
	}
	return nil
}

func (s *Service) ListEquipment(ctx context.Context, filter catalog_postgres_repository.Filter, rawCursor string) (catalog_postgres_repository.Page[catalog_postgres_repository.Equipment], error) {
	filter, scope, err := prepareFilter(filter, rawCursor, "equipment")
	if err != nil {
		return catalog_postgres_repository.Page[catalog_postgres_repository.Equipment]{}, err
	}
	items, err := s.repo.ListEquipment(ctx, filter)
	if err != nil {
		return catalog_postgres_repository.Page[catalog_postgres_repository.Equipment]{}, err
	}
	return makePage(items, filter.Limit, "equipment", scope, func(e catalog_postgres_repository.Equipment) cursor { return cursor{After: e.ID} }), nil
}

func (s *Service) ListExercises(ctx context.Context, filter catalog_postgres_repository.Filter, rawCursor string) (catalog_postgres_repository.Page[catalog_postgres_repository.Exercise], error) {
	filter, scope, err := prepareFilter(filter, rawCursor, "exercises")
	if err == nil {
		err = s.validateEquipment(ctx, filter)
	}
	if err != nil {
		return catalog_postgres_repository.Page[catalog_postgres_repository.Exercise]{}, err
	}
	items, err := s.repo.ListExercises(ctx, filter)
	if err != nil {
		return catalog_postgres_repository.Page[catalog_postgres_repository.Exercise]{}, err
	}
	return makePage(items, filter.Limit, "exercises", scope, func(e catalog_postgres_repository.Exercise) cursor { return cursor{After: strconv.Itoa(e.ID)} }), nil
}

func (s *Service) ListPrograms(ctx context.Context, filter catalog_postgres_repository.Filter, rawCursor string) (catalog_postgres_repository.Page[catalog_postgres_repository.ProgramCard], error) {
	filter, scope, err := prepareFilter(filter, rawCursor, "programs")
	if err == nil {
		err = s.validateEquipment(ctx, filter)
	}
	if err != nil {
		return catalog_postgres_repository.Page[catalog_postgres_repository.ProgramCard]{}, err
	}
	items, err := s.repo.ListPrograms(ctx, filter)
	if err != nil {
		return catalog_postgres_repository.Page[catalog_postgres_repository.ProgramCard]{}, err
	}
	return makePage(items, filter.Limit, "programs", scope, func(p catalog_postgres_repository.ProgramCard) cursor {
		c := cursor{After: p.ID}
		if filter.Sort == "newest" {
			c.CreatedAt = &p.CreatedAt
		} else {
			c.PriceMinor = &p.PriceMinor
		}
		return c
	}), nil
}

func (s *Service) GetProgram(ctx context.Context, programID string, userID int) (catalog_postgres_repository.ProgramDetail, error) {
	if userID < 0 || userID > math.MaxInt32 {
		return catalog_postgres_repository.ProgramDetail{}, fmt.Errorf("invalid user: %w", core_errors.ErrUnauthorized)
	}
	id, err := uuid.Parse(programID)
	if err != nil {
		return catalog_postgres_repository.ProgramDetail{}, fmt.Errorf("invalid program id: %w", core_errors.ErrInvalidArgument)
	}
	return s.repo.GetProgram(ctx, id.String(), userID)
}

func (s *Service) GrantFreeAccess(ctx context.Context, programID string, userID int) (catalog_postgres_repository.Entitlement, error) {
	if userID <= 0 || userID > math.MaxInt32 {
		return catalog_postgres_repository.Entitlement{}, fmt.Errorf("invalid user: %w", core_errors.ErrUnauthorized)
	}
	id, err := uuid.Parse(programID)
	if err != nil {
		return catalog_postgres_repository.Entitlement{}, fmt.Errorf("invalid program id: %w", core_errors.ErrInvalidArgument)
	}
	return s.repo.GrantFreeAccess(ctx, id.String(), userID)
}
