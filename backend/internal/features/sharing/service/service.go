package sharing_service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"github.com/google/uuid"
	"net/url"
	"sport_app/internal/core/transport/http/v2request"
	sharing_postgres_repository "sport_app/internal/features/sharing/repository/postgres"
	workouts_service "sport_app/internal/features/workouts/service"
	"strings"
	"time"
)

type Repository interface {
	Saved(context.Context, int, v2request.PageFilter) (sharing_postgres_repository.SavedPage, error)
	SavedHash(context.Context, int, string) (string, error)
	RemoveSaved(context.Context, int, string) error
	ForkProgram(context.Context, int, string, sharing_postgres_repository.ForkInput, string, string) (sharing_postgres_repository.ForkResult, error)
	Create(context.Context, int, sharing_postgres_repository.ShareInput, string, string) (sharing_postgres_repository.ShareResult, error)
	Get(context.Context, int, string) (sharing_postgres_repository.PublicShare, error)
	Revoke(context.Context, int, string) error
	Save(context.Context, int, string) (string, error)
	ForkShare(context.Context, int, string, string, string) (sharing_postgres_repository.ForkResult, error)
}
type Service struct {
	repo Repository
	base string
}

func NewService(r Repository, base string) *Service {
	u, err := url.Parse(base)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") {
		base = "http://localhost:3000"
	}
	return &Service{r, strings.TrimRight(base, "/")}
}
func tokenHash(token string) (string, error) {
	data, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(data) != 32 {
		return "", v2request.Invalid()
	}
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:]), nil
}
func (s *Service) ForkProgram(ctx context.Context, raw, target, key string, in sharing_postgres_repository.ForkInput) (sharing_postgres_repository.ForkResult, error) {
	id, err := v2request.UserID(raw)
	if err == nil {
		target, err = v2request.UUID(target)
	}
	if err == nil {
		in.SourceVersionID, err = v2request.UUID(in.SourceVersionID)
	}
	if err == nil {
		err = workouts_service.Key(key)
	}
	if err != nil {
		return sharing_postgres_repository.ForkResult{}, err
	}
	if !v2request.Text(in.Title, 200, true) {
		return sharing_postgres_repository.ForkResult{}, v2request.Invalid()
	}
	return s.repo.ForkProgram(ctx, id, target, in, key, workouts_service.Hash(in))
}
func (s *Service) Create(ctx context.Context, raw string, in sharing_postgres_repository.ShareInput) (sharing_postgres_repository.ShareResult, error) {
	id, err := v2request.UserID(raw)
	if err == nil {
		in.TargetID, err = v2request.UUID(in.TargetID)
	}
	if err != nil {
		return sharing_postgres_repository.ShareResult{}, err
	}
	if (in.TargetType != "plan" && in.TargetType != "plan_day") || (in.ExpiresAt != nil && (!in.ExpiresAt.After(time.Now()) || in.ExpiresAt.After(time.Now().AddDate(1, 0, 0)))) {
		return sharing_postgres_repository.ShareResult{}, v2request.Invalid()
	}
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return sharing_postgres_repository.ShareResult{}, err
	}
	token := base64.RawURLEncoding.EncodeToString(data)
	hash, _ := tokenHash(token)
	out, err := s.repo.Create(ctx, id, in, uuid.NewString(), hash)
	if err == nil {
		out.URL = s.base + "/share/" + token
	}
	return out, err
}
func (s *Service) Get(ctx context.Context, raw, token string) (sharing_postgres_repository.PublicShare, error) {
	id := 0
	var err error
	if raw != "" {
		id, err = v2request.UserID(raw)
	}
	hash, tokenErr := tokenHash(token)
	if err == nil {
		err = tokenErr
	}
	if err != nil {
		return sharing_postgres_repository.PublicShare{}, err
	}
	return s.repo.Get(ctx, id, hash)
}
func (s *Service) Revoke(ctx context.Context, raw, target string) error {
	id, err := v2request.UserID(raw)
	if err == nil {
		target, err = v2request.UUID(target)
	}
	if err != nil {
		return err
	}
	return s.repo.Revoke(ctx, id, target)
}
func (s *Service) Save(ctx context.Context, raw, token string) (string, error) {
	id, err := v2request.UserID(raw)
	hash, e := tokenHash(token)
	if err == nil {
		err = e
	}
	if err != nil {
		return "", err
	}
	return s.repo.Save(ctx, id, hash)
}
func (s *Service) ForkShare(ctx context.Context, raw, token, key string) (sharing_postgres_repository.ForkResult, error) {
	id, err := v2request.UserID(raw)
	hash, e := tokenHash(token)
	if err == nil {
		err = e
	}
	if err == nil {
		err = workouts_service.Key(key)
	}
	if err != nil {
		return sharing_postgres_repository.ForkResult{}, err
	}
	return s.repo.ForkShare(ctx, id, hash, key, hash)
}
