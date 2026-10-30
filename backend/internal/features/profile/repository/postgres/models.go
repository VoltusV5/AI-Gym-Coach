package profile_postgres_repository

// Profile представляет анкету текущего пользователя, включая незаполненные поля.
// InjuriesNotes сохраняет исходный boolean и не определяет исключённые упражнения.
type Profile struct {
	Version             int64    `json:"version"`
	Age                 *int     `json:"age"`
	HeightCm            *int     `json:"height_cm"`
	WeightKg            *float64 `json:"weight_kg"`
	Gender              *string  `json:"gender"`
	ActivityLevel       *string  `json:"activity_level"`
	InjuriesNotes       *bool    `json:"injuries_notes"`
	Goal                *string  `json:"goal"`
	Level               *string  `json:"level"`
	TrainingDays        []int    `json:"training_days"`
	EquipmentIDs        []string `json:"equipment_ids"`
	SessionDurationMin  *int     `json:"session_duration_min"`
	ExcludedExerciseIDs []int    `json:"excluded_exercise_ids"`
}

// Patch различает пропущенные поля и явное очищение списка пустым массивом.
// Version обязателен для оптимистической блокировки; nil означает отсутствие изменения.
type Patch struct {
	Version             int64     `json:"version"`
	Age                 *int      `json:"age,omitempty"`
	HeightCm            *int      `json:"height_cm,omitempty"`
	WeightKg            *float64  `json:"weight_kg,omitempty"`
	Gender              *string   `json:"gender,omitempty"`
	ActivityLevel       *string   `json:"activity_level,omitempty"`
	InjuriesNotes       *bool     `json:"injuries_notes,omitempty"`
	Goal                *string   `json:"goal,omitempty"`
	Level               *string   `json:"level,omitempty"`
	TrainingDays        *[]int    `json:"training_days,omitempty"`
	EquipmentIDs        *[]string `json:"equipment_ids,omitempty"`
	SessionDurationMin  *int      `json:"session_duration_min,omitempty"`
	ExcludedExerciseIDs *[]int    `json:"excluded_exercise_ids,omitempty"`
}
