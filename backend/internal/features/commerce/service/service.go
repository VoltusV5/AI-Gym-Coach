package commerce_service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"math"
	"slices"
	core_errors "sport_app/internal/core/errors"
	"sport_app/internal/core/transport/http/v2request"
	commerce_postgres_repository "sport_app/internal/features/commerce/repository/postgres"
	trainers_service "sport_app/internal/features/trainers/service"
	workouts_service "sport_app/internal/features/workouts/service"
	"time"
)

type Repository interface {
	Create(context.Context, int, commerce_postgres_repository.OrderInput, string, string) (commerce_postgres_repository.Order, error)
	Get(context.Context, int, string) (commerce_postgres_repository.Order, error)
	List(context.Context, int, bool, v2request.PageFilter) (commerce_postgres_repository.Page[commerce_postgres_repository.Order], error)
	Confirm(context.Context, commerce_postgres_repository.Event, string, string) error
	Commission(context.Context, int) (commerce_postgres_repository.Commission, error)
	SetCommission(context.Context, int, commerce_postgres_repository.Commission, string) (commerce_postgres_repository.Commission, error)
	Payouts(context.Context, int, bool, v2request.PageFilter) (commerce_postgres_repository.Page[commerce_postgres_repository.Payout], error)
	CreatePayout(context.Context, int, commerce_postgres_repository.PayoutInput, string, string, string) (commerce_postgres_repository.Payout, error)
	DecidePayout(context.Context, int, string, commerce_postgres_repository.Decision, string) (commerce_postgres_repository.Payout, error)
}
type Config struct {
	Stub          bool
	WebhookSecret string
}
type Service struct {
	repo   Repository
	config Config
}

func NewService(repo Repository, c Config) *Service { return &Service{repo, c} }
func disabled() error {
	return &core_errors.DomainError{Code: "PAYMENTS_UNAVAILABLE", Message: "Оплата временно недоступна", Status: 503}
}
func (s *Service) decorate(o commerce_postgres_repository.Order) commerce_postgres_repository.Order {
	if s.config.Stub && o.Status == "pending" {
		url := "/checkout/" + o.ID
		o.CheckoutURL = &url
	}
	return o
}
func (s *Service) Create(ctx context.Context, raw, key string, in commerce_postgres_repository.OrderInput) (commerce_postgres_repository.Order, error) {
	id, err := v2request.UserID(raw)
	if err == nil {
		_, err = v2request.UUID(in.ProgramID)
	}
	if err == nil {
		_, err = v2request.UUID(in.ProgramVersionID)
	}
	if err == nil {
		err = workouts_service.Key(key)
	}
	if err != nil {
		return commerce_postgres_repository.Order{}, err
	}
	if !s.config.Stub {
		return commerce_postgres_repository.Order{}, disabled()
	}
	out, err := s.repo.Create(ctx, id, in, key, workouts_service.Hash(in))
	return s.decorate(out), err
}
func (s *Service) Get(ctx context.Context, raw, target string) (commerce_postgres_repository.Order, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return commerce_postgres_repository.Order{}, err
	}
	target, err = v2request.UUID(target)
	if err != nil {
		return commerce_postgres_repository.Order{}, err
	}
	out, err := s.repo.Get(ctx, id, target)
	return s.decorate(out), err
}
func (s *Service) List(ctx context.Context, raw string, admin bool, f v2request.PageFilter) (commerce_postgres_repository.Page[commerce_postgres_repository.Order], error) {
	id, err := v2request.UserID(raw)
	if err == nil {
		err = trainers_service.ValidatePage(f, false)
	}
	if err != nil {
		return commerce_postgres_repository.Page[commerce_postgres_repository.Order]{}, err
	}
	p, err := s.repo.List(ctx, id, admin, f)
	if err == nil && len(p.Items) > f.Limit {
		p.Items = p.Items[:f.Limit]
		p.NextCursor = v2request.Next(f, p.Items[f.Limit-1].ID)
	}
	for i := range p.Items {
		p.Items[i] = s.decorate(p.Items[i])
	}
	return p, err
}
func (s *Service) Confirm(ctx context.Context, e commerce_postgres_repository.Event, requestID string) error {
	if _, err := v2request.UUID(e.OrderID); err != nil {
		return err
	}
	if !v2request.Text(e.EventID, 200, true) || !v2request.Text(e.PaymentID, 200, true) || e.AmountMinor < 1 || e.AmountMinor > 1000000000000 || e.Currency != "RUB" || !slices.Contains([]string{"payment_succeeded", "payment_failed", "refund_succeeded"}, e.Type) {
		return v2request.Invalid()
	}
	return s.repo.Confirm(ctx, e, workouts_service.Hash(e), requestID)
}
func (s *Service) Verify(data []byte, signature string) bool {
	if !s.config.Stub || s.config.WebhookSecret == "" {
		return false
	}
	given, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	m := hmac.New(sha256.New, []byte(s.config.WebhookSecret))
	_, _ = m.Write(data)
	return hmac.Equal(given, m.Sum(nil))
}
func (s *Service) Sandbox(ctx context.Context, raw, target, action, requestID string) (commerce_postgres_repository.Order, error) {
	if !s.config.Stub {
		return commerce_postgres_repository.Order{}, disabled()
	}
	event := map[string]string{"pay": "payment_succeeded", "fail": "payment_failed", "refund": "refund_succeeded"}[action]
	if event == "" {
		return commerce_postgres_repository.Order{}, v2request.Invalid()
	}
	out, err := s.Get(ctx, raw, target)
	if err != nil {
		return out, err
	}
	e := commerce_postgres_repository.Event{EventID: uuid.NewString(), PaymentID: out.PaymentID, OrderID: out.ID, Type: event, AmountMinor: out.AmountMinor, Currency: out.Currency}
	if err := s.Confirm(ctx, e, requestID); err != nil {
		return out, err
	}
	return s.Get(ctx, raw, target)
}
func (s *Service) Commission(ctx context.Context, raw string) (commerce_postgres_repository.Commission, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return commerce_postgres_repository.Commission{}, err
	}
	return s.repo.Commission(ctx, id)
}
func (s *Service) SetCommission(ctx context.Context, raw string, c commerce_postgres_repository.Commission, requestID string) (commerce_postgres_repository.Commission, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return commerce_postgres_repository.Commission{}, err
	}
	if c.Bps < 0 || c.Bps > 10000 || c.Version < 1 || c.Version == math.MaxInt64 || c.EffectiveAt.IsZero() || c.EffectiveAt.Before(time.Now().Add(-time.Minute)) || c.EffectiveAt.After(time.Now().AddDate(1, 0, 0)) {
		return commerce_postgres_repository.Commission{}, v2request.Invalid()
	}
	return s.repo.SetCommission(ctx, id, c, requestID)
}
func (s *Service) Payouts(ctx context.Context, raw string, admin bool, f v2request.PageFilter) (commerce_postgres_repository.Page[commerce_postgres_repository.Payout], error) {
	id, err := v2request.UserID(raw)
	if err == nil {
		err = trainers_service.ValidatePage(f, false)
	}
	if err != nil {
		return commerce_postgres_repository.Page[commerce_postgres_repository.Payout]{}, err
	}
	p, err := s.repo.Payouts(ctx, id, admin, f)
	if err == nil && len(p.Items) > f.Limit {
		p.Items = p.Items[:f.Limit]
		p.NextCursor = v2request.Next(f, p.Items[f.Limit-1].ID)
	}
	return p, err
}
func (s *Service) CreatePayout(ctx context.Context, raw, key string, in commerce_postgres_repository.PayoutInput, requestID string) (commerce_postgres_repository.Payout, error) {
	id, err := v2request.UserID(raw)
	if err == nil {
		err = workouts_service.Key(key)
	}
	if err != nil {
		return commerce_postgres_repository.Payout{}, err
	}
	if !s.config.Stub {
		return commerce_postgres_repository.Payout{}, disabled()
	}
	if in.TrainerID < 1 || in.TrainerID > math.MaxInt32 || in.AmountMinor < 1 || in.AmountMinor > 1000000000000 || in.Currency != "RUB" {
		return commerce_postgres_repository.Payout{}, v2request.Invalid()
	}
	return s.repo.CreatePayout(ctx, id, in, key, workouts_service.Hash(in), requestID)
}
func (s *Service) DecidePayout(ctx context.Context, raw, target string, in commerce_postgres_repository.Decision, requestID string) (commerce_postgres_repository.Payout, error) {
	id, err := v2request.UserID(raw)
	if err == nil {
		target, err = v2request.UUID(target)
	}
	if err != nil {
		return commerce_postgres_repository.Payout{}, err
	}
	if !s.config.Stub {
		return commerce_postgres_repository.Payout{}, disabled()
	}
	if !slices.Contains([]string{"retry", "cancel", "reconcile"}, in.Action) || !v2request.Text(in.Reason, 4000, true) {
		return commerce_postgres_repository.Payout{}, fmt.Errorf("invalid payout decision: %w", core_errors.ErrInvalidArgument)
	}
	return s.repo.DecidePayout(ctx, id, target, in, requestID)
}
