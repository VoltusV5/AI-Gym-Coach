// Package profile_service проверяет частичные обновления анкеты v2 перед записью.
package profile_service

import (
	"context"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	core_errors "sport_app/internal/core/errors"
	profile_postgres_repository "sport_app/internal/features/profile/repository/postgres"
)

// Repository задаёт чтение и атомарное обновление анкеты одного пользователя.
type Repository interface {
	GetProfile(context.Context, int) (profile_postgres_repository.Profile, error)
	UpdateProfile(context.Context, int, profile_postgres_repository.Patch) (profile_postgres_repository.Profile, error)
}

// Service применяет правила профиля v2 без изменения поведения v1.
type Service struct{ repo Repository }

// NewService создаёт сервис профиля с внедрённым хранилищем.
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// userID проверяет числовой идентификатор из JWT до обращения в PostgreSQL.
func userID(raw string) (int, error) {
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 || id > math.MaxInt32 {
		return 0, fmt.Errorf("invalid profile user: %w", core_errors.ErrUnauthorized)
	}
	return id, nil
}

// GetProfile возвращает только анкету пользователя из JWT; допускает гостевую анкету.
func (s *Service) GetProfile(ctx context.Context, rawID string) (profile_postgres_repository.Profile, error) {
	id, err := userID(rawID)
	if err != nil {
		return profile_postgres_repository.Profile{}, err
	}
	return s.repo.GetProfile(ctx, id)
}

// UpdateProfile проверяет типизированный patch и атомарно сохраняет изменения для пользователя из JWT.
// Возвращает ErrInvalidArgument без вызова БД при неверных значениях; ошибки версий и БД сохраняет в цепочке.
func (s *Service) UpdateProfile(ctx context.Context, rawID string, patch profile_postgres_repository.Patch) (profile_postgres_repository.Profile, error) {
	id, err := userID(rawID)
	if err != nil {
		return profile_postgres_repository.Profile{}, err
	}
	if err := validatePatch(patch); err != nil {
		return profile_postgres_repository.Profile{}, err
	}
	return s.repo.UpdateProfile(ctx, id, patch)
}

// validatePatch проверяет диапазоны хранения, обязательную version, ISO-дни и отсутствие дублей.
// Точные ID оборудования и активность упражнений проверяются репозиторием в транзакции.
func validatePatch(p profile_postgres_repository.Patch) error {
	invalid := func() error { return fmt.Errorf("invalid profile patch: %w", core_errors.ErrInvalidArgument) }
	if p.Version < 1 || p.Version == math.MaxInt64 {
		return invalid()
	}
	if p.Age == nil && p.HeightCm == nil && p.WeightKg == nil && p.Gender == nil &&
		p.ActivityLevel == nil && p.InjuriesNotes == nil && p.Goal == nil && p.Level == nil &&
		p.TrainingDays == nil && p.EquipmentIDs == nil && p.SessionDurationMin == nil && p.ExcludedExerciseIDs == nil {
		return invalid()
	}
	for _, value := range []*int{p.Age, p.HeightCm, p.SessionDurationMin} {
		if value != nil && (*value < 1 || *value > math.MaxInt32) {
			return invalid()
		}
	}
	if p.WeightKg != nil && (math.IsNaN(*p.WeightKg) || math.IsInf(*p.WeightKg, 0) || *p.WeightKg <= 0 || *p.WeightKg > math.MaxInt32) {
		return invalid()
	}
	for _, field := range []struct {
		value *string
		limit int
	}{
		{p.Gender, 100}, {p.ActivityLevel, 200}, {p.Goal, 200},
	} {
		if field.value != nil && (strings.TrimSpace(*field.value) == "" || strings.ContainsRune(*field.value, '\x00') || utf8.RuneCountInString(*field.value) > field.limit) {
			return invalid()
		}
	}
	if p.Level != nil && !slices.Contains([]string{"beginner", "intermediate", "advanced"}, *p.Level) {
		return invalid()
	}
	if p.TrainingDays != nil {
		seen := map[int]bool{}
		for _, day := range *p.TrainingDays {
			if day < 1 || day > 7 || seen[day] {
				return invalid()
			}
			seen[day] = true
		}
	}
	if p.EquipmentIDs != nil {
		seen := map[string]bool{}
		for _, id := range *p.EquipmentIDs {
			if strings.TrimSpace(id) == "" || strings.ContainsRune(id, '\x00') || seen[id] {
				return invalid()
			}
			seen[id] = true
		}
	}
	if p.ExcludedExerciseIDs != nil {
		seen := map[int]bool{}
		for _, id := range *p.ExcludedExerciseIDs {
			if id < 1 || id > math.MaxInt32 || seen[id] {
				return invalid()
			}
			seen[id] = true
		}
	}
	return nil
}
