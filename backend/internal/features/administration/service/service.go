package administration_service

import (
	"context"
	"slices"
	"sport_app/internal/core/transport/http/v2request"
	administration_postgres_repository "sport_app/internal/features/administration/repository/postgres"
	trainers_service "sport_app/internal/features/trainers/service"
	"strconv"
)

type Repository interface {
	Session(context.Context, int) (administration_postgres_repository.User, error)
	List(context.Context, int, v2request.PageFilter, string) (administration_postgres_repository.Page, error)
	Update(context.Context, int, int, administration_postgres_repository.Change, string) (administration_postgres_repository.User, error)
}
type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo} }
func (s *Service) Session(ctx context.Context, raw string) (administration_postgres_repository.User, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return administration_postgres_repository.User{}, err
	}
	return s.repo.Session(ctx, id)
}
func (s *Service) List(ctx context.Context, raw string, f v2request.PageFilter, q string) (administration_postgres_repository.Page, error) {
	id, err := v2request.UserID(raw)
	if err == nil {
		err = trainers_service.ValidatePage(f, true)
	}
	if err == nil && !v2request.Text(q, 200, false) {
		err = v2request.Invalid()
	}
	if err != nil {
		return administration_postgres_repository.Page{}, err
	}
	p, err := s.repo.List(ctx, id, f, q)
	if err == nil && len(p.Items) > f.Limit {
		p.Items = p.Items[:f.Limit]
		p.NextCursor = v2request.Next(f, strconv.Itoa(p.Items[f.Limit-1].ID))
	}
	return p, err
}
func (s *Service) Update(ctx context.Context, raw, target string, p administration_postgres_repository.Change, requestID string) (administration_postgres_repository.User, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return administration_postgres_repository.User{}, err
	}
	other, err := v2request.UserID(target)
	if err != nil {
		return administration_postgres_repository.User{}, v2request.Invalid()
	}
	valid := v2request.Text(p.Reason, 4000, true) && (p.Blocked != nil || p.Roles != nil)
	if p.Roles != nil {
		seen := map[string]bool{}
		valid = valid && len(*p.Roles) > 0
		for _, role := range *p.Roles {
			valid = valid && slices.Contains([]string{"user", "trainer", "admin"}, role) && !seen[role]
			seen[role] = true
		}
	}
	if !valid {
		return administration_postgres_repository.User{}, v2request.Invalid()
	}
	return s.repo.Update(ctx, id, other, p, requestID)
}
