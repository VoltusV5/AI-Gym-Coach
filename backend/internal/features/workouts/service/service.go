package workouts_service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"
	"sport_app/internal/core/transport/http/v2request"
	workouts_domain "sport_app/internal/features/workouts/domain"
)

type Repository interface {
	Complete(context.Context, int, string, string, workouts_domain.Input) (workouts_domain.Result, error)
	Get(context.Context, int, string) (workouts_domain.Workout, error)
	List(context.Context, int, workouts_domain.Filter) ([]workouts_domain.Workout, error)
	Weight(context.Context, int, workouts_domain.Weight) error
	Data(context.Context, int) (workouts_domain.Data, error)
}
type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service { return &Service{repo: repo, now: time.Now} }
func Key(key string) error {
	if !v2request.Text(key, 200, true) {
		return v2request.Invalid()
	}
	return nil
}
func Hash(value any) string {
	raw, _ := json.Marshal(value)
	h := sha256.Sum256(raw)
	return hex.EncodeToString(h[:])
}
func (s *Service) Complete(ctx context.Context, raw, key string, input workouts_domain.Input) (workouts_domain.Result, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return workouts_domain.Result{}, err
	}
	if err := Key(key); err != nil {
		return workouts_domain.Result{}, err
	}
	input, err = workouts_domain.Validate(input, s.now())
	if err != nil {
		return workouts_domain.Result{}, err
	}
	return s.repo.Complete(ctx, id, key, Hash(input), input)
}
func (s *Service) Get(ctx context.Context, raw, target string) (workouts_domain.Workout, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return workouts_domain.Workout{}, err
	}
	target, err = v2request.UUID(target)
	if err != nil {
		return workouts_domain.Workout{}, err
	}
	return s.repo.Get(ctx, id, target)
}
func parseTime(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		t, err = time.Parse(time.DateOnly, raw)
	}
	if err != nil || t.Year() < 1 {
		return t, v2request.Invalid()
	}
	return t, nil
}

type Query struct {
	From, To, PlanID, Cursor string
	Limit                    int
}
type SummaryQuery struct {
	Period, TimeZone, Cursor string
	Limit                    int
}

func (s *Service) List(ctx context.Context, raw string, q Query) (workouts_domain.Page[workouts_domain.Workout], error) {
	page := workouts_domain.Page[workouts_domain.Workout]{}
	id, err := v2request.UserID(raw)
	if err != nil {
		return page, err
	}
	scope := fmt.Sprintf("workouts:%d:%s:%s:%s", id, q.From, q.To, q.PlanID)
	f, err := v2request.PageValues(q.Limit, q.Cursor, scope)
	if err != nil {
		return page, err
	}
	filter := workouts_domain.Filter{Limit: f.Limit}
	filter.From, err = parseTime(q.From)
	if err != nil {
		return page, err
	}
	filter.To, err = parseTime(q.To)
	if err != nil {
		return page, err
	}
	if !filter.From.IsZero() && !filter.To.IsZero() && !filter.To.After(filter.From) {
		return page, v2request.Invalid()
	}
	if q.PlanID != "" {
		filter.PlanID, err = v2request.UUID(q.PlanID)
		if err != nil {
			return page, err
		}
	}
	if f.After != "" {
		var c struct {
			ID string
			At time.Time
		}
		if json.Unmarshal([]byte(f.After), &c) != nil || c.At.IsZero() {
			return page, v2request.Invalid()
		}
		filter.AfterID, err = v2request.UUID(c.ID)
		if err != nil {
			return page, err
		}
		filter.AfterTime = c.At
	}
	items, err := s.repo.List(ctx, id, filter)
	if err != nil {
		return page, err
	}
	page.Items = items
	if len(items) > f.Limit {
		page.Items = items[:f.Limit]
		last := page.Items[f.Limit-1]
		data, _ := json.Marshal(struct {
			ID string
			At time.Time
		}{last.ID, last.FinishedAt})
		page.NextCursor = v2request.Next(f, string(data))
	}
	return page, nil
}
func (s *Service) Weight(ctx context.Context, raw string, w workouts_domain.Weight) error {
	id, err := v2request.UserID(raw)
	if err != nil {
		return err
	}
	date, err := time.Parse(time.DateOnly, w.Day)
	if err != nil || date.Year() < 1 || w.WeightKg <= 0 || math.IsInf(w.WeightKg, 0) || math.IsNaN(w.WeightKg) {
		return v2request.Invalid()
	}
	return s.repo.Weight(ctx, id, w)
}
func location(raw string) (*time.Location, error) {
	if raw == "" {
		raw = "Europe/Moscow"
	}
	if raw == "Local" {
		return nil, v2request.Invalid()
	}
	loc, err := time.LoadLocation(raw)
	if err != nil {
		return nil, v2request.Invalid()
	}
	return loc, nil
}
func (s *Service) Progress(ctx context.Context, raw, period, anchor, zone string) (workouts_domain.Progress, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return workouts_domain.Progress{}, err
	}
	loc, err := location(zone)
	if err != nil {
		return workouts_domain.Progress{}, err
	}
	if period == "" {
		period = "week"
	}
	now := s.now()
	at := now
	if anchor != "" {
		at, err = time.ParseInLocation(time.DateOnly, anchor, loc)
		if err != nil || at.Year() < 1 {
			return workouts_domain.Progress{}, v2request.Invalid()
		}
	}
	start, end, err := workouts_domain.Bounds(period, at, loc)
	if err != nil {
		return workouts_domain.Progress{}, err
	}
	data, err := s.repo.Data(ctx, id)
	if err != nil {
		return workouts_domain.Progress{}, err
	}
	return workouts_domain.Aggregate(data, period, start, end, now, loc), nil
}
func (s *Service) Exercise(ctx context.Context, raw, target string) (workouts_domain.ExerciseProgress, error) {
	id, err := v2request.UserID(raw)
	if err != nil {
		return workouts_domain.ExerciseProgress{}, err
	}
	exercise, err := strconv.Atoi(target)
	if err != nil || exercise < 1 || exercise > math.MaxInt32 {
		return workouts_domain.ExerciseProgress{}, v2request.Invalid()
	}
	data, err := s.repo.Data(ctx, id)
	if err != nil {
		return workouts_domain.ExerciseProgress{}, err
	}
	return workouts_domain.ExerciseHistory(data.Workouts, exercise), nil
}
func (s *Service) Summaries(ctx context.Context, raw string, q SummaryQuery) (workouts_domain.Page[workouts_domain.Summary], error) {
	page := workouts_domain.Page[workouts_domain.Summary]{Items: []workouts_domain.Summary{}}
	id, err := v2request.UserID(raw)
	if err != nil {
		return page, err
	}
	period := q.Period
	if period == "" {
		period = "week"
	}
	if period != "week" && period != "month" {
		return page, v2request.Invalid()
	}
	loc, err := location(q.TimeZone)
	if err != nil {
		return page, err
	}
	f, err := v2request.PageValues(q.Limit, q.Cursor, fmt.Sprintf("summaries:%d:%s:%s", id, period, loc))
	if err != nil {
		return page, err
	}
	now := s.now()
	anchor := now
	if f.After != "" {
		anchor, err = time.ParseInLocation(time.DateOnly, f.After, loc)
		if err != nil || anchor.Year() < 1 {
			return page, v2request.Invalid()
		}
	}
	data, err := s.repo.Data(ctx, id)
	if err != nil {
		return page, err
	}
	if len(data.Workouts) == 0 && len(data.Weights) == 0 {
		return page, nil
	}
	first := now
	for _, w := range data.Workouts {
		if w.FinishedAt.Before(first) {
			first = w.FinishedAt
		}
	}
	for _, w := range data.Weights {
		t, e := time.ParseInLocation(time.DateOnly, w.Day, loc)
		if e == nil && t.Before(first) {
			first = t
		}
	}
	for len(page.Items) <= f.Limit {
		start, end, e := workouts_domain.Bounds(period, anchor, loc)
		if e != nil {
			return page, e
		}
		if end.Before(first) || end.Equal(first) {
			break
		}
		key := fmt.Sprintf("%d:%s:%s:%s", id, period, start.Format(time.DateOnly), loc)
		item := workouts_domain.Summary{ID: uuid.NewSHA1(uuid.NameSpaceOID, []byte(key)).String(), Period: period, PeriodStart: start.Format(time.DateOnly), Status: "ready", Metrics: workouts_domain.Aggregate(data, period, start, end, now, loc), Messages: []string{}}
		page.Items = append(page.Items, item)
		anchor = start.AddDate(0, 0, -1)
	}
	if len(page.Items) > f.Limit {
		last := page.Items[f.Limit]
		page.Items = page.Items[:f.Limit]
		page.NextCursor = v2request.Next(f, last.PeriodStart)
	}
	return page, nil
}
