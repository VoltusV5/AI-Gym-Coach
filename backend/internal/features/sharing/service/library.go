package sharing_service

import (
	"context"
	"sport_app/internal/core/transport/http/v2request"
	repository "sport_app/internal/features/sharing/repository/postgres"
	trainers "sport_app/internal/features/trainers/service"
	workouts "sport_app/internal/features/workouts/service"
)

func (s *Service) Saved(ctx context.Context, raw string, f v2request.PageFilter) (repository.SavedPage, error) {
	user, err := v2request.UserID(raw)
	if err == nil {
		err = trainers.ValidatePage(f, false)
	}
	if err != nil {
		return repository.SavedPage{}, err
	}
	out, err := s.repo.Saved(ctx, user, f)
	if err == nil && len(out.Items) > f.Limit {
		out.Items = out.Items[:f.Limit]
		out.NextCursor = v2request.Next(f, out.Items[f.Limit-1].ID)
	}
	return out, err
}
func (s *Service) savedHash(ctx context.Context, raw, target string) (int, string, error) {
	user, err := v2request.UserID(raw)
	if err == nil {
		target, err = v2request.UUID(target)
	}
	if err != nil {
		return 0, "", err
	}
	hash, err := s.repo.SavedHash(ctx, user, target)
	return user, hash, err
}
func (s *Service) GetSaved(ctx context.Context, raw, target string) (repository.PublicShare, error) {
	user, hash, err := s.savedHash(ctx, raw, target)
	if err != nil {
		return repository.PublicShare{}, err
	}
	return s.repo.Get(ctx, user, hash)
}
func (s *Service) ForkSaved(ctx context.Context, raw, target, key string) (repository.ForkResult, error) {
	user, hash, err := s.savedHash(ctx, raw, target)
	if err == nil {
		err = workouts.Key(key)
	}
	if err != nil {
		return repository.ForkResult{}, err
	}
	return s.repo.ForkShare(ctx, user, hash, key, hash)
}
func (s *Service) RemoveSaved(ctx context.Context, raw, target string) error {
	user, err := v2request.UserID(raw)
	if err == nil {
		target, err = v2request.UUID(target)
	}
	if err != nil {
		return err
	}
	return s.repo.RemoveSaved(ctx, user, target)
}
