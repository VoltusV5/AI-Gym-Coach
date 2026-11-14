package favorites_service

import (
	"context"
	"sport_app/internal/core/transport/http/v2request"
	catalog_postgres_repository "sport_app/internal/features/catalog/repository/postgres"
	trainers_service "sport_app/internal/features/trainers/service"
)

type Repository interface {
	Set(context.Context, int, string, bool) error
	List(context.Context, int, v2request.PageFilter) (catalog_postgres_repository.Page[catalog_postgres_repository.ProgramCard], error)
}
type Service struct{ repo Repository }

func NewService(r Repository) *Service { return &Service{r} }
func (s *Service) Set(ctx context.Context, raw, target string, add bool) error {
	id, err := v2request.UserID(raw)
	if err == nil {
		target, err = v2request.UUID(target)
	}
	if err != nil {
		return err
	}
	return s.repo.Set(ctx, id, target, add)
}
func (s *Service) List(ctx context.Context, raw string, f v2request.PageFilter) (catalog_postgres_repository.Page[catalog_postgres_repository.ProgramCard], error) {
	id, err := v2request.UserID(raw)
	if err == nil {
		err = trainers_service.ValidatePage(f, false)
	}
	if err != nil {
		return catalog_postgres_repository.Page[catalog_postgres_repository.ProgramCard]{}, err
	}
	p, err := s.repo.List(ctx, id, f)
	if err == nil && len(p.Items) > f.Limit {
		p.Items = p.Items[:f.Limit]
		p.NextCursor = v2request.Next(f, p.Items[f.Limit-1].ID)
	}
	return p, err
}
