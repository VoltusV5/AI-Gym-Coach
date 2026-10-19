// Package users_service реализует регистрацию, авторизацию, обновление профиля, генерацию планов, завершение тренировок и работу с заметками.
package users_service

import (
	"context"
	"time"

	core_auth "sport_app/internal/core/auth"
	"sport_app/internal/features/mlclient"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
)

// UsersService координирует бизнес-операции пользователей, репозиторий и ML-клиент.
type UsersService struct {
	usersRepository       UsersRepository
	mlClient              MLClient
	jwt                   *core_auth.JWT
	planGenerationTimeout time.Duration
}

// WithPlanGenerationTimeout задаёт максимальную длительность генерации плана.
func WithPlanGenerationTimeout(timeout time.Duration) func(*UsersService) {
	return func(s *UsersService) {
		if timeout > 0 {
			s.planGenerationTimeout = timeout
		}
	}
}

// UsersRepository задаёт операции хранения, необходимые сервису пользователей.
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

// MLClient задаёт контракт генерации тренировочного плана через ML-сервис.
type MLClient interface {
	GeneratePlan(
		ctx context.Context,
		profile any,
	) (*mlclient.Plan, error)
}

// NewUsersService создаёт сервис с репозиторием usersRepository, ML-клиентом mlClient, менеджером jwt и необязательными настройками.
func NewUsersService(
	usersRepository UsersRepository,
	mlClient MLClient,
	jwt *core_auth.JWT,
	options ...func(*UsersService),
) *UsersService {
	s := &UsersService{
		usersRepository:       usersRepository,
		mlClient:              mlClient,
		jwt:                   jwt,
		planGenerationTimeout: mlGenerationTimeout,
	}
	for _, option := range options {
		option(s)
	}
	return s
}
