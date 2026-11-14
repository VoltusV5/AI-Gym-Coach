package social_service

import (
	"context"
	"slices"
	"sport_app/internal/core/transport/http/v2request"
	social_postgres_repository "sport_app/internal/features/social/repository/postgres"
	trainers_postgres_repository "sport_app/internal/features/trainers/repository/postgres"
	trainers_service "sport_app/internal/features/trainers/service"
	workouts_service "sport_app/internal/features/workouts/service"
	"strconv"
	"time"
)

type Repository interface {
	Eligibility(context.Context, int, string, int) (social_postgres_repository.Eligibility, error)
	SaveReview(context.Context, int, string, string, social_postgres_repository.ReviewInput, int) (social_postgres_repository.Review, error)
	DeleteReview(context.Context, int, string) error
	Reviews(context.Context, int, string, bool, v2request.PageFilter) (social_postgres_repository.ReviewPage, error)
	ModerateReview(context.Context, int, string, social_postgres_repository.Decision, string) (social_postgres_repository.Review, error)
	Report(context.Context, int, social_postgres_repository.ReportInput) (social_postgres_repository.Report, error)
	Reports(context.Context, int, v2request.PageFilter) (social_postgres_repository.Page[social_postgres_repository.Report], error)
	DecideReport(context.Context, int, string, social_postgres_repository.Decision, string) (social_postgres_repository.Report, error)
	Follow(context.Context, int, int, bool) error
	Trainers(context.Context, int, v2request.PageFilter) (social_postgres_repository.Page[trainers_postgres_repository.Profile], error)
	Notifications(context.Context, int, v2request.PageFilter) (social_postgres_repository.NotificationPage, error)
	ReadNotification(context.Context, int, string, bool) (social_postgres_repository.Notification, error)
	Events(context.Context, int, []social_postgres_repository.Event, []string) (social_postgres_repository.Accepted, error)
}
type Service struct {
	repo     Repository
	required int
}

func NewService(r Repository, required int) *Service {
	if required < 1 || required > 100 {
		required = 3
	}
	return &Service{r, required}
}
func identity(raw, target string) (int, string, error) {
	id, err := v2request.UserID(raw)
	if err == nil {
		target, err = v2request.UUID(target)
	}
	return id, target, err
}
func validPage(raw string, f v2request.PageFilter, integer bool) (int, error) {
	id, err := v2request.UserID(raw)
	if err == nil {
		err = trainers_service.ValidatePage(f, integer)
	}
	return id, err
}
func (s *Service) Eligibility(ctx context.Context, raw, target string) (social_postgres_repository.Eligibility, error) {
	id, target, err := identity(raw, target)
	if err != nil {
		return social_postgres_repository.Eligibility{}, err
	}
	return s.repo.Eligibility(ctx, id, target, s.required)
}
func (s *Service) SaveReview(ctx context.Context, raw, program, target string, in social_postgres_repository.ReviewInput) (social_postgres_repository.Review, error) {
	id, err := v2request.UserID(raw)
	if err == nil {
		if target != "" {
			target, err = v2request.UUID(target)
		} else {
			program, err = v2request.UUID(program)
		}
	}
	if err != nil {
		return social_postgres_repository.Review{}, err
	}
	if in.Overall < 1 || in.Overall > 5 || !v2request.Text(in.Text, 10000, false) {
		return social_postgres_repository.Review{}, v2request.Invalid()
	}
	for _, v := range []*int{in.Difficulty, in.Structure, in.DescriptionMatch, in.Usability, in.Result} {
		if v != nil && (*v < 1 || *v > 5) {
			return social_postgres_repository.Review{}, v2request.Invalid()
		}
	}
	return s.repo.SaveReview(ctx, id, program, target, in, s.required)
}
func (s *Service) DeleteReview(ctx context.Context, raw, target string) error {
	id, target, err := identity(raw, target)
	if err != nil {
		return err
	}
	return s.repo.DeleteReview(ctx, id, target)
}
func (s *Service) Reviews(ctx context.Context, raw, program string, admin bool, f v2request.PageFilter) (social_postgres_repository.ReviewPage, error) {
	id := 0
	var err error
	if admin {
		id, err = v2request.UserID(raw)
	} else {
		program, err = v2request.UUID(program)
	}
	if err == nil {
		err = trainers_service.ValidatePage(f, false)
	}
	if err != nil {
		return social_postgres_repository.ReviewPage{}, err
	}
	p, err := s.repo.Reviews(ctx, id, program, admin, f)
	if err == nil && len(p.Items) > f.Limit {
		p.Items = p.Items[:f.Limit]
		p.NextCursor = v2request.Next(f, p.Items[f.Limit-1].ID)
	}
	return p, err
}
func (s *Service) ModerateReview(ctx context.Context, raw, target string, in social_postgres_repository.Decision, requestID string) (social_postgres_repository.Review, error) {
	id, target, err := identity(raw, target)
	if err != nil {
		return social_postgres_repository.Review{}, err
	}
	if !slices.Contains([]string{"published", "hidden"}, in.Status) || !v2request.Text(in.Reason, 4000, true) {
		return social_postgres_repository.Review{}, v2request.Invalid()
	}
	return s.repo.ModerateReview(ctx, id, target, in, requestID)
}
func (s *Service) Report(ctx context.Context, raw string, in social_postgres_repository.ReportInput) (social_postgres_repository.Report, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return social_postgres_repository.Report{}, err
	}
	if !slices.Contains([]string{"program", "review", "trainer"}, in.TargetType) || !v2request.Text(in.Reason, 4000, true) {
		return social_postgres_repository.Report{}, v2request.Invalid()
	}
	if in.TargetType == "trainer" {
		_, err = v2request.UserID(in.TargetID)
	} else {
		in.TargetID, err = v2request.UUID(in.TargetID)
	}
	if err != nil {
		return social_postgres_repository.Report{}, v2request.Invalid()
	}
	return s.repo.Report(ctx, id, in)
}
func (s *Service) Reports(ctx context.Context, raw string, f v2request.PageFilter) (social_postgres_repository.Page[social_postgres_repository.Report], error) {
	id, err := validPage(raw, f, false)
	if err != nil {
		return social_postgres_repository.Page[social_postgres_repository.Report]{}, err
	}
	p, err := s.repo.Reports(ctx, id, f)
	if err == nil && len(p.Items) > f.Limit {
		p.Items = p.Items[:f.Limit]
		p.NextCursor = v2request.Next(f, p.Items[f.Limit-1].ID)
	}
	return p, err
}
func (s *Service) DecideReport(ctx context.Context, raw, target string, in social_postgres_repository.Decision, requestID string) (social_postgres_repository.Report, error) {
	id, target, err := identity(raw, target)
	if err != nil {
		return social_postgres_repository.Report{}, err
	}
	if !slices.Contains([]string{"in_progress", "resolved", "dismissed"}, in.Status) || !v2request.Text(in.Reason, 4000, true) {
		return social_postgres_repository.Report{}, v2request.Invalid()
	}
	return s.repo.DecideReport(ctx, id, target, in, requestID)
}
func (s *Service) Follow(ctx context.Context, raw, target string, follow bool) (int, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return 0, err
	}
	trainer, err := v2request.UserID(target)
	if err != nil {
		return 0, v2request.Invalid()
	}
	if id == trainer {
		return 0, v2request.Invalid()
	}
	return trainer, s.repo.Follow(ctx, id, trainer, follow)
}
func (s *Service) Trainers(ctx context.Context, raw string, f v2request.PageFilter) (social_postgres_repository.Page[trainers_postgres_repository.Profile], error) {
	id, err := validPage(raw, f, true)
	if err != nil {
		return social_postgres_repository.Page[trainers_postgres_repository.Profile]{}, err
	}
	p, err := s.repo.Trainers(ctx, id, f)
	if err == nil && len(p.Items) > f.Limit {
		p.Items = p.Items[:f.Limit]
		p.NextCursor = v2request.Next(f, strconv.Itoa(p.Items[f.Limit-1].ID))
	}
	return p, err
}
func (s *Service) Notifications(ctx context.Context, raw string, f v2request.PageFilter) (social_postgres_repository.NotificationPage, error) {
	id, err := validPage(raw, f, false)
	if err != nil {
		return social_postgres_repository.NotificationPage{}, err
	}
	p, err := s.repo.Notifications(ctx, id, f)
	if err == nil && len(p.Items) > f.Limit {
		p.Items = p.Items[:f.Limit]
		p.NextCursor = v2request.Next(f, p.Items[f.Limit-1].ID)
	}
	return p, err
}
func (s *Service) ReadNotification(ctx context.Context, raw, target string, read bool) (social_postgres_repository.Notification, error) {
	id, target, err := identity(raw, target)
	if err != nil {
		return social_postgres_repository.Notification{}, err
	}
	return s.repo.ReadNotification(ctx, id, target, read)
}
func (s *Service) Events(ctx context.Context, raw string, in []social_postgres_repository.Event) (social_postgres_repository.Accepted, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return social_postgres_repository.Accepted{}, err
	}
	if len(in) < 1 || len(in) > 100 {
		return social_postgres_repository.Accepted{}, v2request.Invalid()
	}
	hashes := []string{}
	now := time.Now()
	for _, e := range in {
		if _, err := v2request.UUID(e.ID); err != nil {
			return social_postgres_repository.Accepted{}, err
		}
		if _, err := v2request.UUID(e.ProgramID); err != nil {
			return social_postgres_repository.Accepted{}, err
		}
		if !slices.Contains([]string{"impression", "view", "save", "unsave", "recommendation_click"}, e.Type) || e.OccurredAt.Before(now.AddDate(0, 0, -30)) || e.OccurredAt.After(now.Add(5*time.Minute)) {
			return social_postgres_repository.Accepted{}, v2request.Invalid()
		}
		for k, v := range e.Context {
			if !slices.Contains([]string{"surface", "recommendation_request_id"}, k) || !v2request.Text(v, 200, false) {
				return social_postgres_repository.Accepted{}, v2request.Invalid()
			}
		}
		hashes = append(hashes, workouts_service.Hash(e))
	}
	return s.repo.Events(ctx, id, in, hashes)
}
