package trainers_postgres_repository

import "time"

type Profile struct {
	ID                int      `json:"id"`
	DisplayName       string   `json:"display_name"`
	PhotoURL          *string  `json:"photo_url"`
	Bio               string   `json:"bio"`
	Specializations   []string `json:"specializations"`
	ExperienceYears   int      `json:"experience_years"`
	Qualifications    string   `json:"qualifications"`
	Verified          bool     `json:"verified"`
	Rating            *float64 `json:"rating"`
	ReviewCount       int      `json:"review_count"`
	PublishedPrograms []string `json:"published_programs"`
}
type Patch struct {
	DisplayName     string    `json:"display_name"`
	PhotoRef        *string   `json:"photo_ref"`
	Bio             *string   `json:"bio,omitempty"`
	Specializations *[]string `json:"specializations,omitempty"`
	ExperienceYears *int      `json:"experience_years,omitempty"`
	Qualifications  *string   `json:"qualifications,omitempty"`
	PhotoSet        bool      `json:"-"`
}
type VerificationInput struct {
	Qualification string   `json:"qualification"`
	DocumentRefs  []string `json:"document_refs"`
}
type Verification struct {
	ID            string     `json:"id"`
	TrainerID     int        `json:"trainer_id"`
	Status        string     `json:"status"`
	Qualification string     `json:"qualification"`
	DocumentRefs  []string   `json:"document_refs"`
	Reason        *string    `json:"reason"`
	CreatedAt     time.Time  `json:"created_at"`
	ReviewedAt    *time.Time `json:"reviewed_at"`
}
type Decision struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}
type Page[T any] struct {
	Items      []T     `json:"items"`
	NextCursor *string `json:"next_cursor"`
}

type Filter struct {
	Limit int
	After string
}
