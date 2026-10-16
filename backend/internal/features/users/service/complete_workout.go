package users_service

import (
	"context"
	"fmt"

	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
)

// CompleteWorkout records a completed workout session, computes weight progression,
// builds session highlights (e.g. personal records, consistency streaks), updates training plan weights,
// and unlocks any newly achieved milestones.
//
// Parameters:
//   - ctx: Operation context.
//   - userID: Authenticated user ID.
//   - req: Workout payload containing completed exercises, sets, reps, and weights.
//
// Returns:
//   - *users_postgres_repository.WorkoutCompleteServiceResult: Summary of new achievements, highlights, and plan weight updates.
//   - error: Non-nil if loading workout history, progression calculation, DB persistence, or achievement unlocking fails.
//
// Side effects:
//   - Appends completed workout record to database.
//   - Updates user working weights in user_data with optimistic concurrency check.
//   - Persists newly unlocked achievements.
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
	planUpdates := users_postgres_repository.BuildPlanUpdates(existing, newWW, names)

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
	if planUpdates == nil {
		planUpdates = []users_postgres_repository.PlanUpdate{}
	}

	return &users_postgres_repository.WorkoutCompleteServiceResult{
		NewAchievements:   newAch,
		SessionHighlights: highlights,
		PlanUpdates:       planUpdates,
	}, nil
}
