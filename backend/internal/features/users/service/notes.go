package users_service

import (
	"context"
	"fmt"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
)

// GetListNotes возвращает заметки userID со стабильной сортировкой, лимитом limit и смещением offset.
// Возвращает страницу, общее число неудалённых заметок и ошибку БД.
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

// CreateNotesUser создаёт заметку пользователя userID с заданными title и body.
// Возвращает созданную заметку либо ошибку БД.
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

// UpdateNotesUser обновляет заметку noteID пользователя userID с проверкой предварительно загруженной версии.
// Сохраняет title и body, увеличивает версию и возвращает заметку.
// Возвращает ErrNotFound при отсутствии своей заметки, ErrConflict при конфликте версии либо ошибку БД.
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

// DeleteNotesUser помечает заметку noteID пользователя userID удалённой, устанавливая deleted_at.
// Для чужой, отсутствующей или уже удалённой заметки возвращает ErrNotFound; также возможна ошибка БД.
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
