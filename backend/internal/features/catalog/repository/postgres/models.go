package catalog_postgres_repository

import "time"

type Page[T any] struct {
	Items      []T     `json:"items"`
	NextCursor *string `json:"next_cursor"`
}

type Equipment struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type Exercise struct {
	ID              int      `json:"id"`
	Name            string   `json:"name"`
	MuscleGroup     string   `json:"muscle_group"`
	MuscleSubgroup  *string  `json:"muscle_subgroup"`
	EquipmentIDs    []string `json:"equipment_ids"`
	SafeForInjuries bool     `json:"safe_for_injuries"`
	ImageURL        *string  `json:"image_url"`
	VideoURL        *string  `json:"video_url"`
	IsActive        bool     `json:"is_active"`
}

type Trainer struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
	Verified    bool   `json:"verified"`
}

type ProgramCard struct {
	CreatedAt          time.Time   `json:"-"`
	ID                 string      `json:"id"`
	Title              string      `json:"title"`
	Trainer            Trainer     `json:"trainer"`
	Goal               string      `json:"goal"`
	Level              string      `json:"level"`
	DurationWeeks      int         `json:"duration_weeks"`
	SessionsPerWeek    int         `json:"sessions_per_week"`
	Equipment          []Equipment `json:"equipment"`
	PriceMinor         int64       `json:"price_minor"`
	Currency           string      `json:"currency"`
	Rating             *float64    `json:"rating"`
	ReviewCount        int         `json:"review_count"`
	PublishedVersionID string      `json:"published_version_id"`
}

type Prescription struct {
	ExerciseID int      `json:"exercise_id"`
	Position   int      `json:"position"`
	Sets       int      `json:"sets"`
	RepsMin    int      `json:"reps_min"`
	RepsMax    int      `json:"reps_max"`
	WeightKg   *float64 `json:"weight_kg"`
	RestSec    int      `json:"rest_sec"`
}

type ProgramDay struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	Position  int            `json:"position"`
	Exercises []Prescription `json:"exercises"`
}

type ProgramDetail struct {
	ProgramCard
	Description      string       `json:"description"`
	Access           string       `json:"access"`
	AllowedActions   []string     `json:"allowed_actions"`
	Days             []ProgramDay `json:"days,omitempty"`
	AllowPublicRemix bool         `json:"allow_public_remix"`
}

type Entitlement struct {
	ID        string    `json:"id"`
	ProgramID string    `json:"program_id"`
	Status    string    `json:"status"`
	GrantedAt time.Time `json:"granted_at"`
	OrderID   *string   `json:"order_id"`
}

type Filter struct {
	AfterCreatedAt time.Time
	AfterPrice     int64
	Limit          int
	After          string
	Query          string
	MuscleGroup    string
	Equipment      []string
	EquipmentSet   bool
	Goal           string
	Level          string
	DurationWeeks  int
	TrainerID      int
	PriceType      string
	Sort           string
}
