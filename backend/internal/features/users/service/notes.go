package users_service

import (
	"context"
	"fmt"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
)

// GetListNotes retrieves a paginated slice of user workout notes sorted deterministically.
//
// Parameters:
//   - ctx: Operation context.
//   - userID: Authenticated user ID.
//   - limit: Number of items per page.
//   - offset: Starting item offset.
//
// Returns:
//   - []users_postgres_repository.Note: List of notes belonging to the user.
//   - int: Total count of non-deleted notes.
//   - error: Non-nil if database query fails.
func (s *UsersService) GetListNotes(
	ctx context.Context,
	userID string,
	limit int,
	offset int,
) ([]users_postgres_repository.Note, int, error) {
	notes, total, err := s.usersRepository.GetListNotes(ctx, userID, limit, offset)
	if err != nil {
		return []users_postgres_repository.Note{}, 0, fmt.Errorf("get list notes: %w", err)
	}

	return notes, total, nil
}

// CreateNotesUser creates a new workout note for the user.
//
// Parameters:
//   - ctx: Operation context.
//   - userID: Authenticated user ID.
//   - title: Title of the note.
//   - body: Content of the note.
//
// Returns:
//   - users_postgres_repository.Note: Created note entity with assigned ID and timestamps.
//   - error: Non-nil if database insertion fails.
//
// Side effects:
//   - Inserts row into user_notes table.
func (s *UsersService) CreateNotesUser(
	ctx context.Context,
	userID string,
	title string,
	body string,
) (users_postgres_repository.Note, error) {
	note, err := s.usersRepository.CreateNotesUser(ctx, userID, title, body)
	if err != nil {
		return users_postgres_repository.Note{}, fmt.Errorf("create note: %w", err)
	}

	return note, nil
}

// UpdateNotesUser updates the title and body of an existing note owned by the user.
//
// Parameters:
//   - ctx: Operation context.
//   - userID: Authenticated user ID.
//   - noteID: ID of note to update.
//   - title: New note title.
//   - body: New note body text.
//
// Returns:
//   - users_postgres_repository.Note: Updated note entity.
//   - error: core_errors.ErrNotFound if note does not exist or does not belong to user, core_errors.ErrConflict on version mismatch.
//
// Side effects:
//   - Updates row and increments version in user_notes table.
func (s *UsersService) UpdateNotesUser(
	ctx context.Context,
	userID string,
	noteID int,
	title string,
	body string,
) (users_postgres_repository.Note, error) {
	currentNote, err := s.usersRepository.GetNoteByID(ctx, userID, noteID)
	if err != nil {
		return users_postgres_repository.Note{}, fmt.Errorf("get note: %w", err)
	}

	note, err := s.usersRepository.UpdateNotesUser(ctx, userID, noteID, currentNote.Version, title, body)
	if err != nil {
		return users_postgres_repository.Note{}, fmt.Errorf("apdate note: %w", err)
	}

	return note, nil
}

// DeleteNotesUser soft-deletes a note verifying ownership (preventing IDOR).
//
// Parameters:
//   - ctx: Operation context.
//   - userID: Authenticated user ID.
//   - noteID: Note ID to delete.
//
// Returns:
//   - error: core_errors.ErrNotFound if note does not exist, belongs to another user, or was already deleted.
//
// Side effects:
//   - Sets deleted_at = NOW() on the target note row.
func (s *UsersService) DeleteNotesUser(
	ctx context.Context,
	userID string,
	noteID int,
) error {
	if err := s.usersRepository.DeleteNotesUser(ctx, userID, noteID); err != nil {
		return fmt.Errorf("delete note: %w", err)
	}

	return nil
}
