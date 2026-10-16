// Package users_service implements core business logic for user management, authentication,
// profile maintenance with optimistic locking, AI plan generation, workout completion, and user notes.
package users_service

import (
	"context"

	core_auth "sport_app/internal/core/auth"
	"sport_app/internal/features/mlclient"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
)

// UsersService coordinates users business operations across repositories and external ML clients.
type UsersService struct {
	usersRepository UsersRepository
	mlClient        MLClient
	jwt             *core_auth.JWT
}

// UsersRepository defines the database interface contract required by UsersService.
type UsersRepository interface {
	EnsureExercisesSeeded(
		ctx context.Context,
	) error
	EnsureAchievements(
		ctx context.Context,
	) error
	CreateGuestUser(
		ctx context.Context,
	) (string, error)
	CreateUser(
		ctx context.Context,
		email string,
		passwordHash string,
	) (users_postgres_repository.User, error)
	PromoteGuestToUser(
		ctx context.Context,
		userID string,
		email string,
		passwordHash string,
	) (users_postgres_repository.User, error)
	GetUserByEmail(
		ctx context.Context,
		email string,
	) (users_postgres_repository.User, string, error)
	GetUserByID(
		ctx context.Context,
		userID string,
	) (users_postgres_repository.User, string, error)
	GetProfile(
		ctx context.Context,
		userID string,
	) (users_postgres_repository.Profile, error)
	UpdateProfile(
		ctx context.Context,
		userID string,
		expectedVersion int64,
		updates map[string]any,
	) error
	GetUserProgramsVersion(
		ctx context.Context,
		userID string,
	) (int64, error)
	GetUserDataVersion(
		ctx context.Context,
		userID string,
	) (int64, error)
	GetExercises(
		ctx context.Context,
		plan mlclient.Plan,
		userID string,
	) (
		users_postgres_repository.EPlanWithWeight,
		users_postgres_repository.EPlanNoWeight,
		error,
	)
	SaveProgram(
		ctx context.Context,
		userID string,
		expectedVersion int64,
		isActive bool,
		planTemplate mlclient.Plan,
		planExercises users_postgres_repository.EPlanNoWeight,
	) error
	GetWorkingWeights(
		ctx context.Context,
		userID string,
	) (map[string]float64, error)
	SaveWorkingWeights(
		ctx context.Context,
		userID string,
		expectedVersion int64,
		workingWeights []byte,
	) error
	CompleteWorkout(
		ctx context.Context,
		userID string,
		req users_postgres_repository.WorkoutCompleteRequest,
		newWorkingWeights []byte,
		expectedVersion int64,
	) error
	UnlockNewAchievements(
		ctx context.Context,
		userID string,
	) ([]users_postgres_repository.AchievementUnlocked, error)
	UpdatePassword(
		ctx context.Context,
		userID string,
		expectedVersion int64,
		newPasswordHash string,
	) error
	GetListNotes(
		ctx context.Context,
		userID string,
		limit int,
		offset int,
	) ([]users_postgres_repository.Note, int, error)
	GetNoteByID(
		ctx context.Context,
		userID string,
		noteID int,
	) (users_postgres_repository.Note, error)
	CreateNotesUser(
		ctx context.Context,
		userID string,
		title string,
		body string,
	) (users_postgres_repository.Note, error)
	UpdateNotesUser(
		ctx context.Context,
		userID string,
		noteID int,
		expectedVersion int64,
		title string,
		body string,
	) (users_postgres_repository.Note, error)
	DeleteNotesUser(
		ctx context.Context,
		userID string,
		noteID int,
	) error
	GetAchievements(
		ctx context.Context,
		userID string,
	) ([]users_postgres_repository.UserAchievement, error)
	LoadCompletedWorkouts(
		ctx context.Context,
		userID string,
	) ([]users_postgres_repository.WorkoutCompleteRequest, error)
	GetExerciseNamesByIDs(
		ctx context.Context,
		ids []int,
	) (map[int]string, error)
}

// MLClient defines the contract for communicating with ML model services.
type MLClient interface {
	GeneratePlan(
		ctx context.Context,
		profile any,
	) (*mlclient.Plan, error)
}

// NewUsersService constructs a new UsersService with database repository, ML client, and JWT provider.
func NewUsersService(
	usersRepository UsersRepository,
	mlClient MLClient,
	jwt *core_auth.JWT,
) *UsersService {
	return &UsersService{
		usersRepository: usersRepository,
		mlClient:        mlClient,
		jwt:             jwt,
	}
}
