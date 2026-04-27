package notes_service

import (
	"context"
	notes_postgres_repository "sport_app/internal/features/notes/repository/postgres"
)

type Repository interface {
	ListNotes(ctx context.Context, userID int) ([]notes_postgres_repository.Note, error)
	CreateNote(ctx context.Context, userID int, title, body string) (notes_postgres_repository.Note, error)
	UpdateNote(ctx context.Context, userID int, noteID int, title, body string) (notes_postgres_repository.Note, error)
	DeleteNote(ctx context.Context, userID int, noteID int) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListNotes(ctx context.Context, userID int) ([]notes_postgres_repository.Note, error) {
	return s.repo.ListNotes(ctx, userID)
}

func (s *Service) CreateNote(ctx context.Context, userID int, title, body string) (notes_postgres_repository.Note, error) {
	return s.repo.CreateNote(ctx, userID, title, body)
}

func (s *Service) UpdateNote(ctx context.Context, userID int, noteID int, title, body string) (notes_postgres_repository.Note, error) {
	return s.repo.UpdateNote(ctx, userID, noteID, title, body)
}

func (s *Service) DeleteNote(ctx context.Context, userID int, noteID int) error {
	return s.repo.DeleteNote(ctx, userID, noteID)
}
