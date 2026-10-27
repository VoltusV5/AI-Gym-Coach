package catalog_service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

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
	Kind  string `json:"kind"`
	After string `json:"after"`
	Scope string `json:"scope"`
}

func prepareFilter(filter catalog_postgres_repository.Filter, raw, kind string) (catalog_postgres_repository.Filter, string, error) {
	if filter.Limit == 0 {
		filter.Limit = 20
	}
	if filter.Limit < 1 || filter.Limit > 100 || filter.DurationWeeks < 0 || filter.TrainerID < 0 {
		return filter, "", fmt.Errorf("invalid filter range: %w", core_errors.ErrInvalidArgument)
	}
	if !slices.Contains([]string{"", "newest", "price_asc", "price_desc"}, filter.Sort) ||
		!slices.Contains([]string{"", "free", "paid"}, filter.PriceType) {
		return filter, "", fmt.Errorf("invalid sort or price type: %w", core_errors.ErrInvalidArgument)
	}
	filter.After = ""
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
		if _, err := uuid.Parse(c.After); err != nil {
			return filter, "", fmt.Errorf("invalid program cursor: %w", core_errors.ErrInvalidArgument)
		}
	case "exercises":
		if id, err := strconv.Atoi(c.After); err != nil || id <= 0 {
			return filter, "", fmt.Errorf("invalid exercise cursor: %w", core_errors.ErrInvalidArgument)
		}
	}
	filter.After = c.After
	return filter, scope, nil
}

func makePage[T any](items []T, limit int, kind, scope string, id func(T) string) catalog_postgres_repository.Page[T] {
	page := catalog_postgres_repository.Page[T]{Items: items}
	if page.Items == nil {
		page.Items = []T{}
	}
	if len(items) > limit {
		page.Items = items[:limit]
		raw, _ := json.Marshal(cursor{Kind: kind, After: id(items[limit-1]), Scope: scope})
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
	return makePage(items, filter.Limit, "equipment", scope, func(e catalog_postgres_repository.Equipment) string { return e.ID }), nil
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
	return makePage(items, filter.Limit, "exercises", scope, func(e catalog_postgres_repository.Exercise) string { return strconv.Itoa(e.ID) }), nil
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
	return makePage(items, filter.Limit, "programs", scope, func(p catalog_postgres_repository.ProgramCard) string { return p.ID }), nil
}

func (s *Service) GetProgram(ctx context.Context, programID string, userID int) (catalog_postgres_repository.ProgramDetail, error) {
	if _, err := uuid.Parse(programID); err != nil {
		return catalog_postgres_repository.ProgramDetail{}, fmt.Errorf("invalid program id: %w", core_errors.ErrInvalidArgument)
	}
	return s.repo.GetProgram(ctx, programID, userID)
}

func (s *Service) GrantFreeAccess(ctx context.Context, programID string, userID int) (catalog_postgres_repository.Entitlement, error) {
	if userID <= 0 {
		return catalog_postgres_repository.Entitlement{}, fmt.Errorf("invalid user: %w", core_errors.ErrUnauthorized)
	}
	if _, err := uuid.Parse(programID); err != nil {
		return catalog_postgres_repository.Entitlement{}, fmt.Errorf("invalid program id: %w", core_errors.ErrInvalidArgument)
	}
	return s.repo.GrantFreeAccess(ctx, programID, userID)
}
