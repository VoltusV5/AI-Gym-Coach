package users_service_test

import (
	"context"

	"sport_app/internal/features/mlclient"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
)

type mockUsersRepository struct {
	EnsureExercisesSeededFunc  func(ctx context.Context) error
	EnsureAchievementsFunc     func(ctx context.Context) error
	CreateGuestUserFunc        func(ctx context.Context) (string, error)
	CreateUserFunc             func(ctx context.Context, email, passwordHash string) (users_postgres_repository.User, error)
	PromoteGuestToUserFunc     func(ctx context.Context, userID, email, passwordHash string) (users_postgres_repository.User, error)
	GetUserByEmailFunc         func(ctx context.Context, email string) (users_postgres_repository.User, string, error)
	GetUserByIDFunc            func(ctx context.Context, userID string) (users_postgres_repository.User, string, error)
	GetProfileFunc             func(ctx context.Context, userID string) (users_postgres_repository.Profile, error)
	UpdateProfileFunc          func(ctx context.Context, userID string, expectedVersion int64, updates map[string]any) error
	GetUserProgramsVersionFunc func(ctx context.Context, userID string) (int64, error)
	GetUserDataVersionFunc     func(ctx context.Context, userID string) (int64, error)
	GetExercisesFunc           func(ctx context.Context, plan mlclient.Plan, userID string) (users_postgres_repository.EPlanWithWeight, users_postgres_repository.EPlanNoWeight, error)
	SaveProgramFunc            func(ctx context.Context, userID string, expectedVersion int64, isActive bool, planTemplate mlclient.Plan, planExercises users_postgres_repository.EPlanNoWeight) error
	GetWorkingWeightsFunc      func(ctx context.Context, userID string) (map[string]float64, error)
	SaveWorkingWeightsFunc     func(ctx context.Context, userID string, expectedVersion int64, workingWeights []byte) error
	CompleteWorkoutFunc        func(ctx context.Context, userID string, req users_postgres_repository.WorkoutCompleteRequest, newWorkingWeights []byte, expectedVersion int64) error
	UnlockNewAchievementsFunc  func(ctx context.Context, userID string) ([]users_postgres_repository.AchievementUnlocked, error)
	UpdatePasswordFunc         func(ctx context.Context, userID string, expectedVersion int64, newPasswordHash string) error
	GetListNotesFunc           func(ctx context.Context, userID string, limit, offset int) ([]users_postgres_repository.Note, int, error)
	GetNoteByIDFunc            func(ctx context.Context, userID string, noteID int) (users_postgres_repository.Note, error)
	CreateNotesUserFunc        func(ctx context.Context, userID, title, body string) (users_postgres_repository.Note, error)
	UpdateNotesUserFunc        func(ctx context.Context, userID string, noteID int, expectedVersion int64, title, body string) (users_postgres_repository.Note, error)
	DeleteNotesUserFunc        func(ctx context.Context, userID string, noteID int) error
	GetAchievementsFunc        func(ctx context.Context, userID string) ([]users_postgres_repository.UserAchievement, error)
	LoadCompletedWorkoutsFunc  func(ctx context.Context, userID string) ([]users_postgres_repository.WorkoutCompleteRequest, error)
	GetExerciseNamesByIDsFunc  func(ctx context.Context, ids []int) (map[int]string, error)
}

func (m *mockUsersRepository) EnsureExercisesSeeded(ctx context.Context) error {
	if m.EnsureExercisesSeededFunc != nil {
		return m.EnsureExercisesSeededFunc(ctx)
	}
	return nil
}
func (m *mockUsersRepository) EnsureAchievements(ctx context.Context) error {
	if m.EnsureAchievementsFunc != nil {
		return m.EnsureAchievementsFunc(ctx)
	}
	return nil
}
func (m *mockUsersRepository) CreateGuestUser(ctx context.Context) (string, error) {
	if m.CreateGuestUserFunc != nil {
		return m.CreateGuestUserFunc(ctx)
	}
	return "guest-uuid-1", nil
}
func (m *mockUsersRepository) CreateUser(ctx context.Context, email, passwordHash string) (users_postgres_repository.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, email, passwordHash)
	}
	e := email
	return users_postgres_repository.User{ID: 101, Email: &e}, nil
}
func (m *mockUsersRepository) PromoteGuestToUser(ctx context.Context, userID, email, passwordHash string) (users_postgres_repository.User, error) {
	if m.PromoteGuestToUserFunc != nil {
		return m.PromoteGuestToUserFunc(ctx, userID, email, passwordHash)
	}
	e := email
	return users_postgres_repository.User{ID: 202, Email: &e}, nil
}
func (m *mockUsersRepository) GetUserByEmail(ctx context.Context, email string) (users_postgres_repository.User, string, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return users_postgres_repository.User{}, "", nil
}
func (m *mockUsersRepository) GetUserByID(ctx context.Context, userID string) (users_postgres_repository.User, string, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, userID)
	}
	return users_postgres_repository.User{}, "", nil
}
func (m *mockUsersRepository) GetProfile(ctx context.Context, userID string) (users_postgres_repository.Profile, error) {
	if m.GetProfileFunc != nil {
		return m.GetProfileFunc(ctx, userID)
	}
	return users_postgres_repository.Profile{Version: 1}, nil
}
func (m *mockUsersRepository) UpdateProfile(ctx context.Context, userID string, expectedVersion int64, updates map[string]any) error {
	if m.UpdateProfileFunc != nil {
		return m.UpdateProfileFunc(ctx, userID, expectedVersion, updates)
	}
	return nil
}
func (m *mockUsersRepository) GetUserProgramsVersion(ctx context.Context, userID string) (int64, error) {
	if m.GetUserProgramsVersionFunc != nil {
		return m.GetUserProgramsVersionFunc(ctx, userID)
	}
	return 1, nil
}
func (m *mockUsersRepository) GetUserDataVersion(ctx context.Context, userID string) (int64, error) {
	if m.GetUserDataVersionFunc != nil {
		return m.GetUserDataVersionFunc(ctx, userID)
	}
	return 1, nil
}
func (m *mockUsersRepository) GetExercises(ctx context.Context, plan mlclient.Plan, userID string) (users_postgres_repository.EPlanWithWeight, users_postgres_repository.EPlanNoWeight, error) {
	if m.GetExercisesFunc != nil {
		return m.GetExercisesFunc(ctx, plan, userID)
	}
	return users_postgres_repository.EPlanWithWeight{}, users_postgres_repository.EPlanNoWeight{}, nil
}
func (m *mockUsersRepository) SaveProgram(ctx context.Context, userID string, expectedVersion int64, isActive bool, planTemplate mlclient.Plan, planExercises users_postgres_repository.EPlanNoWeight) error {
	if m.SaveProgramFunc != nil {
		return m.SaveProgramFunc(ctx, userID, expectedVersion, isActive, planTemplate, planExercises)
	}
	return nil
}
func (m *mockUsersRepository) GetWorkingWeights(ctx context.Context, userID string) (map[string]float64, error) {
	if m.GetWorkingWeightsFunc != nil {
		return m.GetWorkingWeightsFunc(ctx, userID)
	}
	return make(map[string]float64), nil
}
func (m *mockUsersRepository) SaveWorkingWeights(ctx context.Context, userID string, expectedVersion int64, workingWeights []byte) error {
	if m.SaveWorkingWeightsFunc != nil {
		return m.SaveWorkingWeightsFunc(ctx, userID, expectedVersion, workingWeights)
	}
	return nil
}
func (m *mockUsersRepository) CompleteWorkout(ctx context.Context, userID string, req users_postgres_repository.WorkoutCompleteRequest, newWorkingWeights []byte, expectedVersion int64) error {
	if m.CompleteWorkoutFunc != nil {
		return m.CompleteWorkoutFunc(ctx, userID, req, newWorkingWeights, expectedVersion)
	}
	return nil
}
func (m *mockUsersRepository) UnlockNewAchievements(ctx context.Context, userID string) ([]users_postgres_repository.AchievementUnlocked, error) {
	if m.UnlockNewAchievementsFunc != nil {
		return m.UnlockNewAchievementsFunc(ctx, userID)
	}
	return nil, nil
}
func (m *mockUsersRepository) UpdatePassword(ctx context.Context, userID string, expectedVersion int64, newPasswordHash string) error {
	if m.UpdatePasswordFunc != nil {
		return m.UpdatePasswordFunc(ctx, userID, expectedVersion, newPasswordHash)
	}
	return nil
}
func (m *mockUsersRepository) GetListNotes(ctx context.Context, userID string, limit, offset int) ([]users_postgres_repository.Note, int, error) {
	if m.GetListNotesFunc != nil {
		return m.GetListNotesFunc(ctx, userID, limit, offset)
	}
	return nil, 0, nil
}
func (m *mockUsersRepository) GetNoteByID(ctx context.Context, userID string, noteID int) (users_postgres_repository.Note, error) {
	if m.GetNoteByIDFunc != nil {
		return m.GetNoteByIDFunc(ctx, userID, noteID)
	}
	return users_postgres_repository.Note{ID: noteID, UserID: 1, Version: 1}, nil
}
func (m *mockUsersRepository) CreateNotesUser(ctx context.Context, userID, title, body string) (users_postgres_repository.Note, error) {
	if m.CreateNotesUserFunc != nil {
		return m.CreateNotesUserFunc(ctx, userID, title, body)
	}
	return users_postgres_repository.Note{ID: 1, UserID: 1, Title: title, Body: body, Version: 1}, nil
}
func (m *mockUsersRepository) UpdateNotesUser(ctx context.Context, userID string, noteID int, expectedVersion int64, title, body string) (users_postgres_repository.Note, error) {
	if m.UpdateNotesUserFunc != nil {
		return m.UpdateNotesUserFunc(ctx, userID, noteID, expectedVersion, title, body)
	}
	return users_postgres_repository.Note{ID: noteID, UserID: 1, Title: title, Body: body, Version: expectedVersion + 1}, nil
}
func (m *mockUsersRepository) DeleteNotesUser(ctx context.Context, userID string, noteID int) error {
	if m.DeleteNotesUserFunc != nil {
		return m.DeleteNotesUserFunc(ctx, userID, noteID)
	}
	return nil
}
func (m *mockUsersRepository) GetAchievements(ctx context.Context, userID string) ([]users_postgres_repository.UserAchievement, error) {
	if m.GetAchievementsFunc != nil {
		return m.GetAchievementsFunc(ctx, userID)
	}
	return nil, nil
}
func (m *mockUsersRepository) LoadCompletedWorkouts(ctx context.Context, userID string) ([]users_postgres_repository.WorkoutCompleteRequest, error) {
	if m.LoadCompletedWorkoutsFunc != nil {
		return m.LoadCompletedWorkoutsFunc(ctx, userID)
	}
	return nil, nil
}
func (m *mockUsersRepository) GetExerciseNamesByIDs(ctx context.Context, ids []int) (map[int]string, error) {
	if m.GetExerciseNamesByIDsFunc != nil {
		return m.GetExerciseNamesByIDsFunc(ctx, ids)
	}
	return make(map[int]string), nil
}

type mockMLClient struct {
	GeneratePlanFunc func(ctx context.Context, profile any) (*mlclient.Plan, error)
}

func (m *mockMLClient) GeneratePlan(ctx context.Context, profile any) (*mlclient.Plan, error) {
	if m.GeneratePlanFunc != nil {
		return m.GeneratePlanFunc(ctx, profile)
	}
	return &mlclient.Plan{Split_type: "Fullbody"}, nil
}
