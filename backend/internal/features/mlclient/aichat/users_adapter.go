package aichat

import (
	"context"
	"encoding/json"
	"errors"

	core_errors "sport_app/internal/core/errors"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
)

type usersRepoAdapter struct {
	repo *users_postgres_repository.UsersRepository
}

func NewUsersReader(repo *users_postgres_repository.UsersRepository) UsersReader {
	return &usersRepoAdapter{repo: repo}
}

func (a *usersRepoAdapter) GetProfile(
	ctx context.Context,
	userID string,
) (users_postgres_repository.Profile, error) {
	p, err := a.repo.GetProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, core_errors.ErrNotFound) {
			return users_postgres_repository.Profile{}, nil
		}
		return users_postgres_repository.Profile{}, err
	}
	return p, nil
}

func (a *usersRepoAdapter) GetPlanTemplateJSON(
	ctx context.Context,
	userID string,
) (json.RawMessage, error) {
	return a.repo.GetPlanTemplateJSON(ctx, userID)
}

func (a *usersRepoAdapter) InsertChatMessage(
	ctx context.Context,
	userID string,
	role string,
	content string,
) error {
	return a.repo.InsertChatMessage(ctx, userID, role, content)
}

func (a *usersRepoAdapter) GetChatMessages(
	ctx context.Context,
	userID string,
	limit int,
) ([]users_postgres_repository.ChatMessage, error) {
	return a.repo.GetChatMessages(ctx, userID, limit)
}
