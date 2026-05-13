package users_service

import (
	"context"
	"fmt"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
)

func (s *UsersService) GetListNotes(
	ctx context.Context,
	userID string,
) ([]users_postgres_repository.Note, error) {
	notes, err := s.usersRepository.GetListNotes(ctx, userID)
	if err != nil {
		return []users_postgres_repository.Note{}, fmt.Errorf("get list notes: %w", err)
	}

	return notes, nil
}

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