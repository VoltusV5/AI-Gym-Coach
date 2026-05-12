package users_service

import (
	"context"
	"fmt"

	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
)

func (s *UsersService) CompleteWorkout(
	ctx context.Context,
	userID string,
	req users_postgres_repository.WorkoutCompleteRequest,
) (*users_postgres_repository.WorkoutCompleteServiceResult, error) {
	history, err := s.usersRepository.LoadCompletedWorkouts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load completed workouts: %w", err)
	}

	dataVer, err := s.usersRepository.GetUserDataVersion(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user_data version: %w", err)
	}

	existing, err := s.usersRepository.GetWorkingWeights(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get working weights: %w", err)
	}

	newWW, err := users_postgres_repository.MergeWorkingWeightsWithProgression(existing, req)
	if err != nil {
		return nil, fmt.Errorf("merge working weights progression: %w", err)
	}

	ids := users_postgres_repository.CollectExerciseIDsOrdered(req)
	names, err := s.usersRepository.GetExerciseNamesByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("get exercise names: %w", err)
	}

	highlights := users_postgres_repository.BuildSessionHighlights(history, req, names)

	if err := s.usersRepository.CompleteWorkout(ctx, userID, req, newWW, dataVer); err != nil {
		return nil, fmt.Errorf("complete workout: %w", err)
	}

	newAch, err := s.usersRepository.UnlockNewAchievements(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("unlock achievements: %w", err)
	}

	if newAch == nil {
		newAch = []users_postgres_repository.AchievementUnlocked{}
	}
	if highlights == nil {
		highlights = []users_postgres_repository.SessionHighlight{}
	}

	return &users_postgres_repository.WorkoutCompleteServiceResult{
		NewAchievements:   newAch,
		SessionHighlights: highlights,
	}, nil
}
