package insights_service

import (
	"context"
	"sort"
	"sport_app/internal/core/transport/http/v2request"
	repository "sport_app/internal/features/insights/repository/postgres"
	domain "sport_app/internal/features/workouts/domain"
	"time"
)

type Block struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Priority int               `json:"priority"`
	Status   string            `json:"status"`
	Data     any               `json:"data"`
	Action   map[string]string `json:"action"`
}
type Dashboard struct {
	State       string    `json:"state"`
	Blocks      []Block   `json:"blocks"`
	GeneratedAt time.Time `json:"generated_at"`
}

func buildDashboard(snapshot repository.DashboardSnapshot, now time.Time, inactiveDays int) Dashboard {
	out := Dashboard{State: "new_user", Blocks: []Block{}, GeneratedAt: now}
	loc := time.UTC
	if snapshot.ActivePlan != nil {
		if zone, err := time.LoadLocation(snapshot.ActivePlan.Schedule.TimeZone); err == nil {
			loc = zone
		}
	}
	local := now.In(loc)
	today := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	end := today.AddDate(0, 0, 1)
	var latest *domain.Workout
	for i := range snapshot.Data.Workouts {
		w := &snapshot.Data.Workouts[i]
		if !w.FinishedAt.After(now) && (latest == nil || w.FinishedAt.After(latest.FinishedAt)) {
			latest = w
		}
	}
	add := func(kind string, priority int, status string, data any, path, label string) {
		out.Blocks = append(out.Blocks, Block{ID: kind, Type: kind, Priority: priority, Status: status, Data: data, Action: map[string]string{"route": path, "label": label}})
	}
	if snapshot.Unavailable["today_workout"] {
		add("today_workout", 100, "unavailable", nil, "/plans", "Мои планы")
	} else if snapshot.ActivePlan == nil {
		add("today_workout", 100, "empty", map[string]string{"message": "Заполните профиль и создайте свой план."}, "/profile", "Настроить профиль")
	} else {
		out.State = "rest_day"
		plan := snapshot.ActivePlan
		var dayID string
		if plan.Schedule.StartDate <= today.Format("2006-01-02") {
			weekday := int(today.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			for _, e := range plan.Schedule.Entries {
				if e.Weekday == weekday {
					dayID = e.PlanDayID
				}
			}
		}
		status := "empty"
		data := map[string]any{"plan_id": plan.ID, "title": plan.Title, "message": "Сегодня день отдыха. Следующее занятие доступно в вашем плане."}
		if dayID != "" {
			out.State = "workout_today"
			status = "ready"
			data["plan_day_id"] = dayID
			data["message"] = "На сегодня запланирована тренировка."
		}
		add("today_workout", 100, status, data, "/training/"+plan.ID, "Открыть тренировку")
		if latest != nil && latest.FinishedAt.Before(now.AddDate(0, 0, -inactiveDays)) {
			out.State = "inactive"
			out.Blocks[0].Data = map[string]any{"plan_id": plan.ID, "title": plan.Title, "message": "Вернитесь к привычному ритму: начните с посильной тренировки."}
		}
	}
	if !snapshot.Unavailable["today_workout"] && latest != nil && !latest.FinishedAt.Before(today) && latest.FinishedAt.Before(end) {
		out.State = "just_completed"
		add("recent_progress", 120, "ready", map[string]any{"session_id": latest.ID, "metrics": domain.Measure(latest.Input)}, "/progress", "Посмотреть результат")
	}
	progress := domain.Aggregate(snapshot.Data, "28d", today.AddDate(0, 0, -27), end, now, loc)
	progressStatus := "ready"
	var progressData any = progress
	if snapshot.Unavailable["recent_progress"] {
		progressStatus = "unavailable"
		progressData = nil
	}
	add("recent_progress", 70, progressStatus, progressData, "/progress", "Ваш прогресс")
	if out.State == "just_completed" {
		out.Blocks[len(out.Blocks)-1].ID = "progress_28d"
	}
	message := "Регулярность важнее спешки. Выбирайте посильную нагрузку."
	if out.State == "new_user" {
		message = "Начните с профиля: цели, опыт и оборудование помогут выбрать программу."
	}
	if out.State == "just_completed" {
		message = "Занятие сохранено. Дайте себе время на восстановление."
	}
	add("ai_message", 60, "ready", map[string]string{"message": message, "mode": "template"}, "/profile", "Мой профиль")
	if progress.WorkoutCount > 0 {
		add("achievement", 50, "ready", map[string]any{"workout_count": progress.WorkoutCount, "message": "Каждая завершённая тренировка — шаг вперёд."}, "/progress", "История занятий")
	}
	followedStatus := "ready"
	if snapshot.Unavailable["followed_trainers"] {
		followedStatus = "unavailable"
	} else if len(snapshot.Followed) == 0 {
		followedStatus = "empty"
	}
	add("followed_trainers", 30, followedStatus, map[string]any{"items": snapshot.Followed}, "/trainers", "Мои тренеры")
	sort.SliceStable(out.Blocks, func(i, j int) bool { return out.Blocks[i].Priority > out.Blocks[j].Priority })
	return out
}
func (s *Service) Dashboard(ctx context.Context, raw string) (Dashboard, error) {
	user, err := v2request.UserID(raw)
	if err != nil {
		return Dashboard{}, err
	}
	snapshot, err := s.repo.Dashboard(ctx, user)
	if err != nil {
		return Dashboard{}, err
	}
	out := buildDashboard(snapshot, time.Now().UTC(), s.inactiveDays)
	recommendations, err := s.Recommendations(ctx, raw, 3)
	status := "ready"
	var data any = recommendations
	if err != nil {
		status = "unavailable"
		data = map[string]any{"items": []any{}}
	} else if len(recommendations.Items) == 0 {
		status = "empty"
	}
	out.Blocks = append(out.Blocks, Block{ID: "recommendation", Type: "recommendation", Priority: 40, Status: status, Data: data, Action: map[string]string{"route": "/recommendations", "label": "Подобрать программу"}})
	sort.SliceStable(out.Blocks, func(i, j int) bool { return out.Blocks[i].Priority > out.Blocks[j].Priority })
	return out, nil
}
