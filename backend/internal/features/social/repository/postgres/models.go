package social_postgres_repository

import "time"

type ReviewInput struct {
	Overall          int    `json:"overall"`
	Difficulty       *int   `json:"difficulty,omitempty"`
	Structure        *int   `json:"structure,omitempty"`
	DescriptionMatch *int   `json:"description_match,omitempty"`
	Usability        *int   `json:"usability,omitempty"`
	Result           *int   `json:"result,omitempty"`
	Text             string `json:"text"`
}
type Review struct {
	ID               string `json:"id"`
	ProgramID        string `json:"program_id"`
	ProgramVersionID string `json:"program_version_id"`
	UserID           int    `json:"user_id"`
	ReviewInput
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
type Eligibility struct {
	OwnReview *Review `json:"own_review,omitempty"`
	Eligible  bool    `json:"eligible"`
	Completed int     `json:"completed_workouts"`
	Required  int     `json:"required_workouts"`
	Reason    *string `json:"reason"`
}
type Page[T any] struct {
	Items      []T     `json:"items"`
	NextCursor *string `json:"next_cursor"`
}
type ReviewPage struct {
	Page[Review]
	Average *float64 `json:"average"`
	Count   int      `json:"count"`
}
type ReportInput struct {
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Reason     string `json:"reason"`
}
type Report struct {
	TargetReview *Review `json:"target_review,omitempty"`
	ID           string  `json:"id"`
	ReporterID   int     `json:"reporter_id"`
	ReportInput
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	Resolution *string   `json:"resolution"`
}
type Decision struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}
type Notification struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	CreatedAt   time.Time  `json:"created_at"`
	ReadAt      *time.Time `json:"read_at"`
	Title       string     `json:"title"`
	Message     string     `json:"message"`
	TargetRoute string     `json:"target_route"`
}
type NotificationPage struct {
	Page[Notification]
	UnreadCount int `json:"unread_count"`
}
type Event struct {
	ID         string            `json:"event_id"`
	Type       string            `json:"type"`
	ProgramID  string            `json:"program_id"`
	OccurredAt time.Time         `json:"occurred_at"`
	Context    map[string]string `json:"context,omitempty"`
}
type Accepted struct {
	Accepted       int `json:"accepted"`
	DuplicateCount int `json:"duplicate_count"`
}
