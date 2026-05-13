package users_service

import (
	"context"
	"fmt"
	"time"

	core_logger "sport_app/internal/core/logger"
	"sport_app/internal/features/mlclient"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"

	"go.uber.org/zap"
)

const mlGenerationTimeout = 52 * time.Second

type generationResult struct {
	plan mlclient.Plan
	err  error
}

func (s *UsersService) GeneratePlan(
	ctx context.Context,
	userID string,
) (users_postgres_repository.EPlanWithWeight, error) {
	log := core_logger.FromContext(ctx)

	profile, err := s.usersRepository.GetProfile(ctx, userID)
	if err != nil {
		return users_postgres_repository.EPlanWithWeight{}, fmt.Errorf("get profile: %w", err)
	}

	mlCtx, cancel := context.WithTimeout(ctx, mlGenerationTimeout)
	defer cancel()

	log.Debug("generating plan from ML", zap.String("user_id", userID))

	plan, err := s.requestMLPlan(mlCtx, profile, log)
	if err != nil {
		return users_postgres_repository.EPlanWithWeight{}, err
	}

	planWithWeight, planNoWeight, err := s.usersRepository.GetExercises(ctx, plan, userID)
	if err != nil {
		return users_postgres_repository.EPlanWithWeight{}, fmt.Errorf("get exercises: %w", err)
	}

	programVersion, err := s.usersRepository.GetUserProgramsVersion(ctx, userID)
	if err != nil {
		return users_postgres_repository.EPlanWithWeight{}, fmt.Errorf("get user_programs version: %w", err)
	}

	if err := s.usersRepository.SaveProgram(ctx, userID, programVersion, true, plan, planNoWeight); err != nil {
		return users_postgres_repository.EPlanWithWeight{}, fmt.Errorf("save program: %w", err)
	}

	dataVersion, err := s.usersRepository.GetUserDataVersion(ctx, userID)
	if err != nil {
		return users_postgres_repository.EPlanWithWeight{}, fmt.Errorf("get user_data version: %w", err)
	}

	existingWeights, err := s.usersRepository.GetWorkingWeights(ctx, userID)
	if err != nil {
		return users_postgres_repository.EPlanWithWeight{}, fmt.Errorf("load working weights: %w", err)
	}

	users_postgres_repository.ApplyExistingWeightsToPlan(&planWithWeight, existingWeights)

	mergedJSON, err := users_postgres_repository.MergeWorkingWeightsJSON(existingWeights, planWithWeight)
	if err != nil {
		return users_postgres_repository.EPlanWithWeight{}, fmt.Errorf("merge working weights: %w", err)
	}

	if err := s.usersRepository.SaveWorkingWeights(ctx, userID, dataVersion, mergedJSON); err != nil {
		return users_postgres_repository.EPlanWithWeight{}, fmt.Errorf("save working weights: %w", err)
	}

	return planWithWeight, nil
}

func (s *UsersService) requestMLPlan(
	ctx context.Context,
	profile users_postgres_repository.Profile,
	log *core_logger.Logger,
) (mlclient.Plan, error) {
	ch := make(chan generationResult, 1)
	start := time.Now()

	go func() {
		plan, err := s.mlClient.GeneratePlan(ctx, profile)
		if err != nil {
			ch <- generationResult{err: err}
			return
		}
		ch <- generationResult{plan: *plan}
	}()

	select {
	case <-ctx.Done():
		log.Debug("ML request cancelled",
			zap.Duration("duration", time.Since(start)),
			zap.Error(ctx.Err()),
		)
		return mlclient.Plan{}, fmt.Errorf("ML request: %w", ctx.Err())
	case res := <-ch:
		duration := time.Since(start)
		if res.err != nil {
			log.Debug("ML request failed",
				zap.Duration("duration", duration),
				zap.Error(res.err),
			)
			return mlclient.Plan{}, fmt.Errorf("ML generate plan: %w", res.err)
		}
		log.Debug("ML request succeeded", zap.Duration("duration", duration))
		return res.plan, nil
	}
}