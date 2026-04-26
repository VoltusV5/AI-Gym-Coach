package users_postgres_repository

import "time"

type User struct {
	ID                 int        `json:"id"`
	Version            int64      `json:"version"`
	IsAnonymous        bool       `json:"is_anonymous"`
	Email              *string    `json:"email"`
	SubscriptionStatus *string    `json:"subscription_status"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          *time.Time `json:"updated_at"`
}

type Profile struct {
	ID              int        `json:"id"`
	Version         int64      `json:"version"`
	UserID          int        `json:"user_id"`
	Age             *int       `json:"age"`
	Gender          *string    `json:"gender"`
	HeightCm        *int       `json:"height_cm"`
	WeightKg        *int       `json:"weight_kg"`
	ActivityLevel   *string    `json:"activity_level"`
	InjuriesNotes   *bool      `json:"injuries_notes"`
	Goal            *string    `json:"goal"`
	FitnessLevel    *string    `json:"fitness_level"`
	TrainingDaysMap []string   `json:"training_days_map"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
}

type ExWithWeight struct {
	ID     int      `json:"id"`
	EXName string   `json:"exercise_name"`
	Weight *float64 `json:"weight"`
}

type EDaysWithWeight struct {
	Day       string           `json:"day"`
	DName     string           `json:"day_name"`
	Exercises [][]ExWithWeight `json:"exercises"`
}

type EPlanWithWeight struct {
	Split string            `json:"split"`
	Plan  []EDaysWithWeight `json:"plan"`
}

type ExNoWeight struct {
	ID     int    `json:"id"`
	EXName string `json:"exercise_name"`
}

type EDaysNoWeight struct {
	Day       string         `json:"day"`
	DName     string         `json:"day_name"`
	Exercises [][]ExNoWeight `json:"exercises"`
}

type EPlanNoWeight struct {
	Split string          `json:"split"`
	Plan  []EDaysNoWeight `json:"plan"`
}

type WorkoutCompleteRequest struct {
	DayCode    string     `json:"day_code"`
	FinishedAt time.Time  `json:"finished_at"`
	Slots      []SlotData `json:"slots"`
}

type SlotData struct {
	SlotIndex  int       `json:"slot_index"`
	ExerciseID int       `json:"exercise_id"`
	Sets       []SetData `json:"sets"`
}

type SetData struct {
	WeightKg float64 `json:"weight_kg"`
	Reps     int     `json:"reps"`
}
