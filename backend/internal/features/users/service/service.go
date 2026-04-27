package users_service

import (
	"context"

	core_auth "sport_app/internal/core/auth"
	"sport_app/internal/features/mlclient"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
)

type UsersService struct {
	usersRepository UsersRepository
	mlClient        MLClient
	jwt             *core_auth.JWT
}

type UsersRepository interface {
	EnsureExercisesSeeded(
		ctx context.Context,
	) error
	CreateGuestUser(
		ctx context.Context,
	) (string, error)
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
	) error
	UpdatePassword(
		ctx context.Context,
		userID string,
		expectedVersion int64,
		newPasswordHash string,
	) error
	GetListNotes(
		ctx context.Context,
		userID string,
	) ([]users_postgres_repository.Note, error)
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
}

type MLClient interface {
	GeneratePlan(
		ctx context.Context,
		profile any,
	) (*mlclient.Plan, error)
}

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
