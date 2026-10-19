package users_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	core_auth "sport_app/internal/core/auth"
	core_errors "sport_app/internal/core/errors"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
	users_service "sport_app/internal/features/users/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func newTestUsersService(repo *mockUsersRepository) (*users_service.UsersService, *core_auth.JWT) {
	jwtService := core_auth.NewJWT(core_auth.Config{
		Secret:   "test-secret-key-users-service-12345",
		TokenTTL: 24 * time.Hour,
	})
	ml := &mockMLClient{}
	svc := users_service.NewUsersService(repo, ml, jwtService)
	return svc, jwtService
}

func TestUsersService_CreateGuestUser(t *testing.T) {
	repo := &mockUsersRepository{
		CreateGuestUserFunc: func(ctx context.Context) (string, error) {
			return "guest-uuid-999", nil
		},
	}
	svc, jwt := newTestUsersService(repo)

	token, err := svc.CreateGuestUser(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := jwt.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, "guest-uuid-999", claims.UserID)
	assert.True(t, claims.IsAnonymous)
}

func TestUsersService_RegisterUser_DirectRegistration(t *testing.T) {
	var createdEmail, createdHash string
	repo := &mockUsersRepository{
		CreateUserFunc: func(ctx context.Context, email, passwordHash string) (users_postgres_repository.User, error) {
			createdEmail = email
			createdHash = passwordHash
			e := email
			return users_postgres_repository.User{ID: 55, Email: &e}, nil
		},
	}
	svc, jwt := newTestUsersService(repo)

	token, user, err := svc.RegisterUser(context.Background(), "", "Athlete@Example.COM", "SecretPass123!")
	require.NoError(t, err)
	assert.Equal(t, 55, user.ID)
	assert.Equal(t, "athlete@example.com", createdEmail) // Адрес приведён к нижнему регистру.

	// Проверяем хеш пароля.
	err = bcrypt.CompareHashAndPassword([]byte(createdHash), []byte("SecretPass123!"))
	assert.NoError(t, err)

	// Проверяем данные JWT.
	claims, err := jwt.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, "55", claims.UserID)
	assert.False(t, claims.IsAnonymous)
}

func TestUsersService_RegisterUser_GuestPromotion(t *testing.T) {
	var promotedID, promotedEmail string
	repo := &mockUsersRepository{
		PromoteGuestToUserFunc: func(ctx context.Context, userID, email, passwordHash string) (users_postgres_repository.User, error) {
			promotedID = userID
			promotedEmail = email
			e := email
			return users_postgres_repository.User{ID: 77, Email: &e}, nil
		},
	}
	svc, jwt := newTestUsersService(repo)

	token, user, err := svc.RegisterUser(context.Background(), "guest-123", "User@Test.com", "Password456")
	require.NoError(t, err)
	assert.Equal(t, 77, user.ID)
	assert.Equal(t, "guest-123", promotedID)
	assert.Equal(t, "user@test.com", promotedEmail)

	// Проверяем данные JWT.
	claims, err := jwt.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, "guest-123", claims.UserID)
	assert.False(t, claims.IsAnonymous)
}

func TestUsersService_LoginUser(t *testing.T) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("CorrectPassword"), bcrypt.DefaultCost)
	require.NoError(t, err)

	repo := &mockUsersRepository{
		GetUserByEmailFunc: func(ctx context.Context, email string) (users_postgres_repository.User, string, error) {
			if email == "found@user.com" {
				e := email
				return users_postgres_repository.User{ID: 10, Email: &e}, string(hashedPassword), nil
			}
			return users_postgres_repository.User{}, "", core_errors.ErrNotFound
		},
	}
	svc, jwt := newTestUsersService(repo)

	t.Run("Valid credentials returns token", func(t *testing.T) {
		token, user, err := svc.LoginUser(context.Background(), "Found@User.com", "CorrectPassword")
		require.NoError(t, err)
		assert.Equal(t, 10, user.ID)

		claims, err := jwt.ParseToken(token)
		require.NoError(t, err)
		assert.Equal(t, "10", claims.UserID)
		assert.False(t, claims.IsAnonymous)
	})

	t.Run("Incorrect password returns ErrUnauthorized", func(t *testing.T) {
		_, _, err := svc.LoginUser(context.Background(), "found@user.com", "WrongPassword")
		require.Error(t, err)
		assert.True(t, errors.Is(err, core_errors.ErrUnauthorized))
	})

	t.Run("Non-existent user returns ErrUnauthorized", func(t *testing.T) {
		_, _, err := svc.LoginUser(context.Background(), "unknown@user.com", "AnyPassword")
		require.Error(t, err)
		assert.True(t, errors.Is(err, core_errors.ErrUnauthorized))
	})
}

func TestUsersService_UpdateProfile(t *testing.T) {
	t.Run("Empty updates returns ErrInvalidArgument", func(t *testing.T) {
		svc, _ := newTestUsersService(&mockUsersRepository{})
		_, err := svc.UpdateProfile(context.Background(), "user-1", 1, map[string]any{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, core_errors.ErrInvalidArgument))
	})

	t.Run("Version mismatch returns ErrConflict", func(t *testing.T) {
		repo := &mockUsersRepository{
			UpdateProfileFunc: func(ctx context.Context, userID string, expectedVersion int64, updates map[string]any) error {
				return core_errors.ErrConflict
			},
		}
		svc, _ := newTestUsersService(repo)
		_, err := svc.UpdateProfile(context.Background(), "user-1", 1, map[string]any{"height": 180})
		require.Error(t, err)
		assert.True(t, errors.Is(err, core_errors.ErrConflict))
	})

	t.Run("Successful update returns fresh profile", func(t *testing.T) {
		repo := &mockUsersRepository{
			UpdateProfileFunc: func(ctx context.Context, userID string, expectedVersion int64, updates map[string]any) error {
				assert.Equal(t, int64(2), expectedVersion)
				return nil
			},
			GetProfileFunc: func(ctx context.Context, userID string) (users_postgres_repository.Profile, error) {
				return users_postgres_repository.Profile{
					Version:  3,
					HeightCm: intPtr(182),
				}, nil
			},
		}
		svc, _ := newTestUsersService(repo)
		profile, err := svc.UpdateProfile(context.Background(), "user-1", 2, map[string]any{"height_cm": 182})
		require.NoError(t, err)
		assert.Equal(t, int64(3), profile.Version)
		assert.Equal(t, 182, *profile.HeightCm)
	})
}

func intPtr(i int) *int {
	return &i
}
