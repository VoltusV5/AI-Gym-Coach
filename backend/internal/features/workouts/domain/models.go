package workouts_domain

import (
	plans_postgres_repository "sport_app/internal/features/plans/repository/postgres"
	"time"
)

type Set struct {
	Reps     int     `json:"reps"`
	WeightKg float64 `json:"weight_kg"`
	IsWarmup bool    `json:"is_warmup"`
}
type Slot struct {
	ExerciseID int   `json:"exercise_id"`
	Position   int   `json:"position"`
	Sets       []Set `json:"sets"`
}
type Input struct {
	ClientSessionID string    `json:"client_session_id"`
	PlanID          string    `json:"plan_id"`
	PlanVersion     int64     `json:"plan_version"`
	PlanDayID       string    `json:"plan_day_id"`
	StartedAt       time.Time `json:"started_at"`
	FinishedAt      time.Time `json:"finished_at"`
	Slots           []Slot    `json:"slots"`
}
type Workout struct {
	ID string `json:"id"`
	Input
}
type Metrics struct {
	VolumeKg      float64 `json:"volume_kg"`
	WorkingSets   int     `json:"working_sets"`
	ExerciseCount int     `json:"exercise_count"`
	DurationSec   int64   `json:"duration_sec"`
}
type Highlight struct {
	ExerciseID   *int     `json:"exercise_id"`
	Metric       string   `json:"metric"`
	Previous     *float64 `json:"previous"`
	Current      float64  `json:"current"`
	DeltaPercent *float64 `json:"delta_percent"`
	Message      string   `json:"message"`
}
type Result struct {
	SessionID       string      `json:"session_id"`
	Metrics         Metrics     `json:"metrics"`
	Highlights      []Highlight `json:"highlights"`
	NewAchievements []any       `json:"new_achievements"`
}
type Point struct {
	SessionID       string    `json:"session_id"`
	FinishedAt      time.Time `json:"finished_at"`
	WorkingWeightKg float64   `json:"working_weight_kg"`
	Reps            int       `json:"reps"`
	VolumeKg        float64   `json:"volume_kg"`
}
type ExerciseProgress struct {
	ExerciseID int         `json:"exercise_id"`
	Points     []Point     `json:"points"`
	Records    []Highlight `json:"records"`
}
type Weight struct {
	Day      string  `json:"day"`
	WeightKg float64 `json:"weight_kg"`
}
type Revision struct {
	Plan plans_postgres_repository.Plan
	At   time.Time
}
type Data struct {
	Workouts  []Workout
	Weights   []Weight
	Revisions []Revision
}
type Progress struct {
	Period         string      `json:"period"`
	Start          time.Time   `json:"start"`
	End            time.Time   `json:"end"`
	TimeZone       string      `json:"time_zone"`
	WorkoutCount   int         `json:"workout_count"`
	VolumeKg       float64     `json:"volume_kg"`
	Adherence      *float64    `json:"adherence"`
	StreakWeeks    int         `json:"streak_weeks"`
	WeightChangeKg *float64    `json:"weight_change_kg"`
	Highlights     []Highlight `json:"highlights"`
	ComputedAt     time.Time   `json:"computed_at"`
}
type Summary struct {
	ID          string   `json:"id"`
	Period      string   `json:"period"`
	PeriodStart string   `json:"period_start"`
	Status      string   `json:"status"`
	Metrics     Progress `json:"metrics"`
	Messages    []string `json:"messages"`
}
type Page[T any] struct {
	Items      []T     `json:"items"`
	NextCursor *string `json:"next_cursor"`
}
type Filter struct {
	Limit     int
	AfterID   string
	AfterTime time.Time
	From      time.Time
	To        time.Time
	PlanID    string
}
