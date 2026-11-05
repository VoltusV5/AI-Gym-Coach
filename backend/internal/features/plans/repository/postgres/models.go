package plans_postgres_repository

import "time"

type Prescription struct {
	ExerciseID int      `json:"exercise_id"`
	Position   int      `json:"position"`
	Sets       int      `json:"sets"`
	RepsMin    int      `json:"reps_min"`
	RepsMax    int      `json:"reps_max"`
	WeightKg   *float64 `json:"weight_kg"`
	RestSec    int      `json:"rest_sec"`
}

type Day struct {
	ID        string         `json:"id,omitempty"`
	Title     string         `json:"title"`
	Position  int            `json:"position"`
	Exercises []Prescription `json:"exercises"`
}

type ScheduleEntry struct {
	Weekday   int    `json:"weekday"`
	PlanDayID string `json:"plan_day_id"`
}

type Schedule struct {
	TimeZone  string          `json:"time_zone"`
	StartDate string          `json:"start_date"`
	Entries   []ScheduleEntry `json:"entries"`
}

type Input struct {
	Version  int64    `json:"version,omitempty"`
	Title    string   `json:"title"`
	Days     []Day    `json:"days"`
	Schedule Schedule `json:"schedule"`
}

type Plan struct {
	ID                     string    `json:"id"`
	Version                int64     `json:"version"`
	Title                  string    `json:"title"`
	SourceType             string    `json:"source_type"`
	SourceProgramVersionID *string   `json:"source_program_version_id"`
	Days                   []Day     `json:"days"`
	Schedule               Schedule  `json:"schedule"`
	IsActive               bool      `json:"is_active"`
	CreatedAt              time.Time `json:"-"`
}

type Page struct {
	Items      []Plan  `json:"items"`
	NextCursor *string `json:"next_cursor"`
}

type Filter struct {
	Limit          int
	AfterID        string
	AfterCreatedAt time.Time
}
