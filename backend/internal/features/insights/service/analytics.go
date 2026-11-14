package insights_service

import (
	"context"
	"sport_app/internal/core/transport/http/v2request"
	repository "sport_app/internal/features/insights/repository/postgres"
	"time"
)

type Period = repository.Period

func ParsePeriod(from, to string, now time.Time) (Period, error) {
	out := Period{From: now.UTC().AddDate(0, 0, -30), To: now.UTC()}
	var err error
	parse := func(value string) (time.Time, error) {
		if len(value) == 10 {
			return time.Parse("2006-01-02", value)
		}
		return time.Parse(time.RFC3339, value)
	}
	if from != "" {
		out.From, err = parse(from)
		if err != nil {
			return out, v2request.Invalid()
		}
	}
	if to != "" {
		out.To, err = parse(to)
		if err != nil {
			return out, v2request.Invalid()
		}
	}
	if !out.From.Before(out.To) || out.To.Sub(out.From) > 366*24*time.Hour || out.To.After(now.Add(5*time.Minute)) {
		return out, v2request.Invalid()
	}
	return out, nil
}
func (s *Service) Analytics(ctx context.Context, raw string, admin bool, from, to string) (map[string]any, error) {
	user, err := v2request.UserID(raw)
	if err != nil {
		return nil, err
	}
	period, err := ParsePeriod(from, to, time.Now())
	if err != nil {
		return nil, err
	}
	return s.repo.Analytics(ctx, user, admin, period)
}
