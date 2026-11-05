package trainers_service

import (
	"context"
	"math"
	"strconv"
	"strings"

	"sport_app/internal/core/transport/http/v2request"
	trainers_postgres_repository "sport_app/internal/features/trainers/repository/postgres"
)

type Repository interface {
	GetProfile(context.Context, int, int, bool) (trainers_postgres_repository.Profile, error)
	UpdateProfile(context.Context, int, trainers_postgres_repository.Patch) (trainers_postgres_repository.Profile, error)
	GetVerification(context.Context, int) (trainers_postgres_repository.Verification, error)
	Submit(context.Context, int, trainers_postgres_repository.VerificationInput) (trainers_postgres_repository.Verification, error)
	Decide(context.Context, int, string, trainers_postgres_repository.Decision, string) (trainers_postgres_repository.Verification, error)
	ListVerifications(context.Context, int, trainers_postgres_repository.Filter) (trainers_postgres_repository.Page[trainers_postgres_repository.Verification], error)
	ListTrainers(context.Context, int, trainers_postgres_repository.Filter) (trainers_postgres_repository.Page[trainers_postgres_repository.Profile], error)
}
type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) Profile(ctx context.Context, raw, target string) (trainers_postgres_repository.Profile, error) {
	own := target == ""
	id := 0
	if own {
		var err error
		id, err = v2request.UserID(raw)
		if err != nil {
			return trainers_postgres_repository.Profile{}, err
		}
		return s.repo.GetProfile(ctx, id, id, true)
	}
	n, err := strconv.Atoi(target)
	if err != nil || n < 1 || n > math.MaxInt32 {
		return trainers_postgres_repository.Profile{}, v2request.Invalid()
	}
	return s.repo.GetProfile(ctx, id, n, false)
}
func (s *Service) Update(ctx context.Context, raw string, p trainers_postgres_repository.Patch) (trainers_postgres_repository.Profile, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return trainers_postgres_repository.Profile{}, err
	}
	valid := v2request.Text(p.DisplayName, 200, true)
	if p.PhotoRef != nil {
		valid = valid && v2request.Text(*p.PhotoRef, 2048, false)
	}
	if p.Bio != nil {
		valid = valid && v2request.Text(*p.Bio, 10000, false)
	}
	if p.Qualifications != nil {
		valid = valid && v2request.Text(*p.Qualifications, 10000, false)
	}
	if p.ExperienceYears != nil {
		valid = valid && *p.ExperienceYears >= 0 && *p.ExperienceYears <= math.MaxInt32
	}
	if p.Specializations != nil {
		seen := map[string]bool{}
		valid = valid && len(*p.Specializations) <= 100
		for _, x := range *p.Specializations {
			valid = valid && v2request.Text(x, 200, true) && !seen[x]
			seen[x] = true
		}
	}
	if !valid {
		return trainers_postgres_repository.Profile{}, v2request.Invalid()
	}
	return s.repo.UpdateProfile(ctx, id, p)
}
func (s *Service) Verification(ctx context.Context, raw string) (trainers_postgres_repository.Verification, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return trainers_postgres_repository.Verification{}, err
	}
	return s.repo.GetVerification(ctx, id)
}
func (s *Service) Submit(ctx context.Context, raw string, p trainers_postgres_repository.VerificationInput) (trainers_postgres_repository.Verification, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return trainers_postgres_repository.Verification{}, err
	}
	if !v2request.Text(p.Qualification, 10000, true) || len(p.DocumentRefs) == 0 || len(p.DocumentRefs) > 20 {
		return trainers_postgres_repository.Verification{}, v2request.Invalid()
	}
	seen := map[string]bool{}
	for _, ref := range p.DocumentRefs {
		if !v2request.Text(ref, 2048, true) || seen[ref] {
			return trainers_postgres_repository.Verification{}, v2request.Invalid()
		}
		seen[ref] = true
	}
	return s.repo.Submit(ctx, id, p)
}
func (s *Service) Decide(ctx context.Context, raw, target string, p trainers_postgres_repository.Decision, requestID string) (trainers_postgres_repository.Verification, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return trainers_postgres_repository.Verification{}, err
	}
	target, err = v2request.UUID(target)
	if err != nil {
		return trainers_postgres_repository.Verification{}, err
	}
	if (p.Status != "approved" && p.Status != "rejected") || !v2request.Text(p.Reason, 4000, true) {
		return trainers_postgres_repository.Verification{}, v2request.Invalid()
	}
	return s.repo.Decide(ctx, id, target, p, requestID)
}
func ValidatePage(f v2request.PageFilter, integer bool) error {
	if f.Limit < 1 || f.Limit > 100 {
		return v2request.Invalid()
	}
	if f.After == "" {
		return nil
	}
	if integer {
		id, err := strconv.Atoi(f.After)
		if err != nil || id < 1 || id > math.MaxInt32 {
			return v2request.Invalid()
		}
	} else {
		_, err := v2request.UUID(f.After)
		if err != nil || strings.Contains(f.After, ":") {
			return v2request.Invalid()
		}
	}
	return nil
}
func (s *Service) Verifications(ctx context.Context, raw string, f v2request.PageFilter) (trainers_postgres_repository.Page[trainers_postgres_repository.Verification], error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return trainers_postgres_repository.Page[trainers_postgres_repository.Verification]{}, err
	}
	if err := ValidatePage(f, false); err != nil {
		return trainers_postgres_repository.Page[trainers_postgres_repository.Verification]{}, err
	}
	page, err := s.repo.ListVerifications(ctx, id, trainers_postgres_repository.Filter{Limit: f.Limit, After: f.After})
	if err == nil && len(page.Items) > f.Limit {
		page.Items = page.Items[:f.Limit]
		page.NextCursor = v2request.Next(f, page.Items[f.Limit-1].ID)
	}
	return page, err
}
func (s *Service) Trainers(ctx context.Context, raw string, f v2request.PageFilter) (trainers_postgres_repository.Page[trainers_postgres_repository.Profile], error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return trainers_postgres_repository.Page[trainers_postgres_repository.Profile]{}, err
	}
	if err := ValidatePage(f, true); err != nil {
		return trainers_postgres_repository.Page[trainers_postgres_repository.Profile]{}, err
	}
	page, err := s.repo.ListTrainers(ctx, id, trainers_postgres_repository.Filter{Limit: f.Limit, After: f.After})
	if err == nil && len(page.Items) > f.Limit {
		page.Items = page.Items[:f.Limit]
		page.NextCursor = v2request.Next(f, strconv.Itoa(page.Items[f.Limit-1].ID))
	}
	return page, err
}
