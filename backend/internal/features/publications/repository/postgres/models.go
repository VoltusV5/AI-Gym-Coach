package publications_postgres_repository

import plans_postgres_repository "sport_app/internal/features/plans/repository/postgres"

type Draft struct {
	Version          int64                           `json:"version"`
	Title            string                          `json:"title"`
	Description      string                          `json:"description"`
	Goal             string                          `json:"goal"`
	Level            string                          `json:"level"`
	DurationWeeks    int                             `json:"duration_weeks"`
	SessionsPerWeek  int                             `json:"sessions_per_week"`
	EquipmentIDs     []string                        `json:"equipment_ids"`
	CategoryIDs      []string                        `json:"category_ids"`
	PriceMinor       int64                           `json:"price_minor"`
	Currency         string                          `json:"currency"`
	AllowPublicRemix bool                            `json:"allow_public_remix"`
	Days             []plans_postgres_repository.Day `json:"days"`
}
type Program struct {
	ID                 string  `json:"id"`
	Version            int64   `json:"version"`
	PublishedVersionID *string `json:"published_version_id"`
	Status             string  `json:"status"`
	Draft              Draft   `json:"draft"`
	VersionID          string  `json:"version_id"`
}
type Moderation struct {
	ProgramVersionID string `json:"program_version_id,omitempty"`
	Status           string `json:"status"`
	Reason           string `json:"reason"`
}
type Workflow struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Version int64  `json:"version"`
}
type Page struct {
	Items      []Program `json:"items"`
	NextCursor *string   `json:"next_cursor"`
}

type Filter struct {
	Limit int
	After string
}
