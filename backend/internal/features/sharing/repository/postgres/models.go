package sharing_postgres_repository

import (
	plans_postgres_repository "sport_app/internal/features/plans/repository/postgres"
	"time"
)

type ForkInput struct {
	SourceVersionID string `json:"source_version_id"`
	Title           string `json:"title"`
}
type ForkResult struct {
	PlanID          string  `json:"plan_id"`
	SourceVersionID *string `json:"source_program_version_id"`
	Version         int64   `json:"version"`
}
type ShareInput struct {
	TargetType string     `json:"target_type"`
	TargetID   string     `json:"target_id"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}
type ShareResult struct {
	ID         string     `json:"id"`
	URL        string     `json:"url"`
	ExpiresAt  *time.Time `json:"expires_at"`
	Visibility string     `json:"visibility"`
}
type Author struct {
	ID          *int   `json:"id,omitempty"`
	DisplayName string `json:"display_name"`
	Verified    bool   `json:"verified"`
}
type PublicShare struct {
	CanRevoke       bool                            `json:"can_revoke"`
	ID              string                          `json:"id"`
	Title           string                          `json:"title"`
	TargetType      string                          `json:"target_type"`
	Author          Author                          `json:"author"`
	Days            []plans_postgres_repository.Day `json:"days"`
	RequiresAccess  bool                            `json:"requires_access"`
	CanFork         bool                            `json:"can_fork"`
	SourceProgramID *string                         `json:"source_program_id"`
	SourceVersionID *string                         `json:"source_program_version_id"`
	SnapshotAt      time.Time                       `json:"snapshot_at"`
}
