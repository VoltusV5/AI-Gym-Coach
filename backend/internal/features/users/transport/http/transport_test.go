package users_transport_http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	core_auth "sport_app/internal/core/auth"
	core_errors "sport_app/internal/core/errors"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
	users_transport_http "sport_app/internal/features/users/transport/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUsersService struct {
	CreateGuestUserFunc    func(ctx context.Context) (string, error)
	RegisterUserFunc       func(ctx context.Context, guestUserID, email, password string) (string, users_postgres_repository.User, error)
	LoginUserFunc          func(ctx context.Context, email, password string) (string, users_postgres_repository.User, error)
	GetProfileFunc         func(ctx context.Context, userID string) (users_postgres_repository.Profile, error)
	UpdateProfileFunc      func(ctx context.Context, userID string, expectedVersion int64, updates map[string]any) (users_postgres_repository.Profile, error)
	GeneratePlanFunc       func(ctx context.Context, userID string) (users_postgres_repository.EPlanWithWeight, error)
	CompleteWorkoutFunc    func(ctx context.Context, userID string, req users_postgres_repository.WorkoutCompleteRequest) (*users_postgres_repository.WorkoutCompleteServiceResult, error)
	ChangeUserPasswordFunc func(ctx context.Context, userID, current_password, new_password string) error
	GetListNotesFunc       func(ctx context.Context, userID string, limit, offset int) ([]users_postgres_repository.Note, int, error)
	CreateNotesUserFunc    func(ctx context.Context, userID, title, body string) (users_postgres_repository.Note, error)
	UpdateNotesUserFunc    func(ctx context.Context, userID string, noteID int, title, body string) (users_postgres_repository.Note, error)
	DeleteNotesUserFunc    func(ctx context.Context, userID string, noteID int) error
	GetAchievementsFunc    func(ctx context.Context, userID string) ([]users_postgres_repository.UserAchievement, error)
}

func (m *mockUsersService) CreateGuestUser(ctx context.Context) (string, error) {
	if m.CreateGuestUserFunc != nil {
		return m.CreateGuestUserFunc(ctx)
	}
	return "token-guest", nil
}
func (m *mockUsersService) RegisterUser(ctx context.Context, guestUserID, email, password string) (string, users_postgres_repository.User, error) {
	if m.RegisterUserFunc != nil {
		return m.RegisterUserFunc(ctx, guestUserID, email, password)
	}
	e := email
	return "token-reg", users_postgres_repository.User{ID: 10, Email: &e}, nil
}
func (m *mockUsersService) LoginUser(ctx context.Context, email, password string) (string, users_postgres_repository.User, error) {
	if m.LoginUserFunc != nil {
		return m.LoginUserFunc(ctx, email, password)
	}
	e := email
	return "token-login", users_postgres_repository.User{ID: 10, Email: &e}, nil
}
func (m *mockUsersService) GetProfile(ctx context.Context, userID string) (users_postgres_repository.Profile, error) {
	if m.GetProfileFunc != nil {
		return m.GetProfileFunc(ctx, userID)
	}
	return users_postgres_repository.Profile{}, nil
}
func (m *mockUsersService) UpdateProfile(ctx context.Context, userID string, expectedVersion int64, updates map[string]any) (users_postgres_repository.Profile, error) {
	if m.UpdateProfileFunc != nil {
		return m.UpdateProfileFunc(ctx, userID, expectedVersion, updates)
	}
	return users_postgres_repository.Profile{}, nil
}
func (m *mockUsersService) GeneratePlan(ctx context.Context, userID string) (users_postgres_repository.EPlanWithWeight, error) {
	if m.GeneratePlanFunc != nil {
		return m.GeneratePlanFunc(ctx, userID)
	}
	return users_postgres_repository.EPlanWithWeight{}, nil
}
func (m *mockUsersService) CompleteWorkout(ctx context.Context, userID string, req users_postgres_repository.WorkoutCompleteRequest) (*users_postgres_repository.WorkoutCompleteServiceResult, error) {
	if m.CompleteWorkoutFunc != nil {
		return m.CompleteWorkoutFunc(ctx, userID, req)
	}
	return &users_postgres_repository.WorkoutCompleteServiceResult{}, nil
}
func (m *mockUsersService) ChangeUserPassword(ctx context.Context, userID, current_password, new_password string) error {
	if m.ChangeUserPasswordFunc != nil {
		return m.ChangeUserPasswordFunc(ctx, userID, current_password, new_password)
	}
	return nil
}
func (m *mockUsersService) GetListNotes(ctx context.Context, userID string, limit, offset int) ([]users_postgres_repository.Note, int, error) {
	if m.GetListNotesFunc != nil {
		return m.GetListNotesFunc(ctx, userID, limit, offset)
	}
	return nil, 0, nil
}
func (m *mockUsersService) CreateNotesUser(ctx context.Context, userID, title, body string) (users_postgres_repository.Note, error) {
	if m.CreateNotesUserFunc != nil {
		return m.CreateNotesUserFunc(ctx, userID, title, body)
	}
	return users_postgres_repository.Note{ID: 1, UserID: 1, Title: title, Body: body}, nil
}
func (m *mockUsersService) UpdateNotesUser(ctx context.Context, userID string, noteID int, title, body string) (users_postgres_repository.Note, error) {
	if m.UpdateNotesUserFunc != nil {
		return m.UpdateNotesUserFunc(ctx, userID, noteID, title, body)
	}
	return users_postgres_repository.Note{ID: noteID, UserID: 1, Title: title, Body: body}, nil
}
func (m *mockUsersService) DeleteNotesUser(ctx context.Context, userID string, noteID int) error {
	if m.DeleteNotesUserFunc != nil {
		return m.DeleteNotesUserFunc(ctx, userID, noteID)
	}
	return nil
}
func (m *mockUsersService) GetAchievements(ctx context.Context, userID string) ([]users_postgres_repository.UserAchievement, error) {
	if m.GetAchievementsFunc != nil {
		return m.GetAchievementsFunc(ctx, userID)
	}
	return nil, nil
}

func newTestHandler(svc *mockUsersService) (*users_transport_http.UsersHTTPHandler, *core_auth.JWT) {
	jwt := core_auth.NewJWT(core_auth.Config{
		Secret:   "test-secret-users-transport-key-12345",
		TokenTTL: 24 * time.Hour,
	})
	handler := users_transport_http.NewUsersHTTPHandler(svc, jwt)
	return handler, jwt
}

func contextWithUser(userID string) context.Context {
	return context.WithValue(context.Background(), core_http_middleware.UserIDKey, userID)
}

func TestUsersHTTP_CreateGuestUser(t *testing.T) {
	mockSvc := &mockUsersService{
		CreateGuestUserFunc: func(ctx context.Context) (string, error) {
			return "guest-jwt-token-xyz", nil
		},
	}
	handler, _ := newTestHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/auth/guest", nil)
	rec := httptest.NewRecorder()

	handler.CreateGuestUser(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Bearer guest-jwt-token-xyz", rec.Header().Get("Authorization"))

	var resp users_transport_http.CreateGuestUserResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "guest-jwt-token-xyz", resp.Token)
}

func TestUsersHTTP_RegisterUser(t *testing.T) {
	t.Run("Direct registration without guest token returns 200", func(t *testing.T) {
		mockSvc := &mockUsersService{
			RegisterUserFunc: func(ctx context.Context, guestUserID, email, password string) (string, users_postgres_repository.User, error) {
				assert.Empty(t, guestUserID)
				assert.Equal(t, "new@user.com", email)
				e := email
				return "new-user-token", users_postgres_repository.User{ID: 101, Email: &e}, nil
			},
		}
		handler, _ := newTestHandler(mockSvc)

		body := `{"email": "new@user.com", "password": "Password123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		handler.RegisterUser(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Bearer new-user-token", rec.Header().Get("Authorization"))

		var resp users_transport_http.RegisterUserResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "new-user-token", resp.Token)
		assert.Equal(t, 101, resp.User.ID)
	})

	t.Run("Promoting guest with valid guest token returns 200", func(t *testing.T) {
		mockSvc := &mockUsersService{
			RegisterUserFunc: func(ctx context.Context, guestUserID, email, password string) (string, users_postgres_repository.User, error) {
				assert.Equal(t, "guest-42", guestUserID)
				e := email
				return "promoted-user-token", users_postgres_repository.User{ID: 42, Email: &e}, nil
			},
		}
		handler, jwt := newTestHandler(mockSvc)

		guestToken, err := jwt.CreateToken("guest-42", true) // Гостевая сессия.
		require.NoError(t, err)

		body := `{"email": "guest@user.com", "password": "Password123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+guestToken)
		rec := httptest.NewRecorder()

		handler.RegisterUser(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Bearer promoted-user-token", rec.Header().Get("Authorization"))
	})

	t.Run("Attempting to register with non-guest token returns 409 Conflict", func(t *testing.T) {
		mockSvc := &mockUsersService{}
		handler, jwt := newTestHandler(mockSvc)

		registeredToken, err := jwt.CreateToken("user-99", false) // Сессия зарегистрированного пользователя.
		require.NoError(t, err)

		body := `{"email": "already@user.com", "password": "Password123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+registeredToken)
		rec := httptest.NewRecorder()

		handler.RegisterUser(rec, req)

		assert.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("Invalid email or short password returns 400 Bad Request", func(t *testing.T) {
		mockSvc := &mockUsersService{}
		handler, _ := newTestHandler(mockSvc)

		body := `{"email": "not-an-email", "password": "short"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		handler.RegisterUser(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestUsersHTTP_LoginUser(t *testing.T) {
	t.Run("Valid credentials returns 200", func(t *testing.T) {
		mockSvc := &mockUsersService{
			LoginUserFunc: func(ctx context.Context, email, password string) (string, users_postgres_repository.User, error) {
				e := email
				return "login-jwt-token", users_postgres_repository.User{ID: 200, Email: &e}, nil
			},
		}
		handler, _ := newTestHandler(mockSvc)

		body := `{"email": "valid@user.com", "password": "Password123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		handler.LoginUser(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Bearer login-jwt-token", rec.Header().Get("Authorization"))
	})

	t.Run("Invalid credentials returns 401 Unauthorized", func(t *testing.T) {
		mockSvc := &mockUsersService{
			LoginUserFunc: func(ctx context.Context, email, password string) (string, users_postgres_repository.User, error) {
				return "", users_postgres_repository.User{}, core_errors.ErrUnauthorized
			},
		}
		handler, _ := newTestHandler(mockSvc)

		body := `{"email": "valid@user.com", "password": "WrongPassword"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		handler.LoginUser(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestUsersHTTP_Notes_PaginationAndIDOR(t *testing.T) {
	t.Run("ListNotes returns paginated list and total", func(t *testing.T) {
		mockSvc := &mockUsersService{
			GetListNotesFunc: func(ctx context.Context, userID string, limit, offset int) ([]users_postgres_repository.Note, int, error) {
				assert.Equal(t, "user-10", userID)
				assert.Equal(t, 20, limit)
				assert.Equal(t, 40, offset)
				return []users_postgres_repository.Note{
					{ID: 1, UserID: 10, Title: "Note 1"},
				}, 75, nil
			},
		}
		handler, _ := newTestHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/notes?limit=20&offset=40", nil)
		req = req.WithContext(contextWithUser("user-10"))
		rec := httptest.NewRecorder()

		handler.ListNotes(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp map[string]any
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, float64(75), resp["total"])
		assert.Equal(t, float64(20), resp["limit"])
		assert.Equal(t, float64(40), resp["offset"])
	})

	t.Run("DeleteNote succeeds for owner and returns 204", func(t *testing.T) {
		mockSvc := &mockUsersService{
			DeleteNotesUserFunc: func(ctx context.Context, userID string, noteID int) error {
				assert.Equal(t, "user-10", userID)
				assert.Equal(t, 55, noteID)
				return nil
			},
		}
		handler, _ := newTestHandler(mockSvc)

		req := httptest.NewRequest(http.MethodDelete, "/notes/55", nil)
		req.SetPathValue("id", "55")
		req = req.WithContext(contextWithUser("user-10"))
		rec := httptest.NewRecorder()

		handler.DeleteNote(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("DeleteNote returns 404 for other user's note (IDOR protection)", func(t *testing.T) {
		mockSvc := &mockUsersService{
			DeleteNotesUserFunc: func(ctx context.Context, userID string, noteID int) error {
				return core_errors.ErrNotFound
			},
		}
		handler, _ := newTestHandler(mockSvc)

		req := httptest.NewRequest(http.MethodDelete, "/notes/99", nil)
		req.SetPathValue("id", "99")
		req = req.WithContext(contextWithUser("attacker-user"))
		rec := httptest.NewRecorder()

		handler.DeleteNote(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("DeleteNote returns 400 for invalid ID parameter", func(t *testing.T) {
		mockSvc := &mockUsersService{}
		handler, _ := newTestHandler(mockSvc)

		req := httptest.NewRequest(http.MethodDelete, "/notes/abc", nil)
		req.SetPathValue("id", "abc")
		req = req.WithContext(contextWithUser("user-10"))
		rec := httptest.NewRecorder()

		handler.DeleteNote(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestUsersHTTP_Profile(t *testing.T) {
	t.Run("GetProfile returns 200", func(t *testing.T) {
		mockSvc := &mockUsersService{
			GetProfileFunc: func(ctx context.Context, userID string) (users_postgres_repository.Profile, error) {
				assert.Equal(t, "user-42", userID)
				return users_postgres_repository.Profile{ID: 1, UserID: 42, Version: 2}, nil
			},
		}
		handler, _ := newTestHandler(mockSvc)

		req := httptest.NewRequest(http.MethodGet, "/profile", nil)
		req = req.WithContext(contextWithUser("user-42"))
		rec := httptest.NewRecorder()

		handler.GetProfile(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var p users_postgres_repository.Profile
		err := json.NewDecoder(rec.Body).Decode(&p)
		require.NoError(t, err)
		assert.Equal(t, int64(2), p.Version)
	})

	t.Run("UpdateProfile missing version returns 400", func(t *testing.T) {
		mockSvc := &mockUsersService{}
		handler, _ := newTestHandler(mockSvc)

		body := `{"height_cm": 180}`
		req := httptest.NewRequest(http.MethodPost, "/profile", bytes.NewBufferString(body))
		req = req.WithContext(contextWithUser("user-42"))
		rec := httptest.NewRecorder()

		handler.UpdateProfile(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("UpdateProfile empty update returns 400", func(t *testing.T) {
		mockSvc := &mockUsersService{}
		handler, _ := newTestHandler(mockSvc)

		body := `{"version": 1}`
		req := httptest.NewRequest(http.MethodPost, "/profile", bytes.NewBufferString(body))
		req = req.WithContext(contextWithUser("user-42"))
		rec := httptest.NewRecorder()

		handler.UpdateProfile(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("UpdateProfile version conflict returns 409", func(t *testing.T) {
		mockSvc := &mockUsersService{
			UpdateProfileFunc: func(ctx context.Context, userID string, expectedVersion int64, updates map[string]any) (users_postgres_repository.Profile, error) {
				return users_postgres_repository.Profile{}, core_errors.ErrConflict
			},
		}
		handler, _ := newTestHandler(mockSvc)

		body := `{"version": 1, "height_cm": 185}`
		req := httptest.NewRequest(http.MethodPost, "/profile", bytes.NewBufferString(body))
		req = req.WithContext(contextWithUser("user-42"))
		rec := httptest.NewRecorder()

		handler.UpdateProfile(rec, req)

		assert.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("UpdateProfile successful returns 200", func(t *testing.T) {
		mockSvc := &mockUsersService{
			UpdateProfileFunc: func(ctx context.Context, userID string, expectedVersion int64, updates map[string]any) (users_postgres_repository.Profile, error) {
				assert.Equal(t, int64(3), expectedVersion)
				return users_postgres_repository.Profile{ID: 1, UserID: 42, Version: 4}, nil
			},
		}
		handler, _ := newTestHandler(mockSvc)

		body := `{"version": 3, "height_cm": 185}`
		req := httptest.NewRequest(http.MethodPost, "/profile", bytes.NewBufferString(body))
		req = req.WithContext(contextWithUser("user-42"))
		rec := httptest.NewRecorder()

		handler.UpdateProfile(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var p users_postgres_repository.Profile
		err := json.NewDecoder(rec.Body).Decode(&p)
		require.NoError(t, err)
		assert.Equal(t, int64(4), p.Version)
	})
}
