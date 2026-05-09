package users_transport_http

import (
	"context"
	"net/http"
	"time"

	core_auth "sport_app/internal/core/auth"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_server "sport_app/internal/core/transport/http/server"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
)

type UsersHTTPHandler struct {
	usersService UsersService
	jwt          *core_auth.JWT
}

type UsersService interface {
	CreateGuestUser(
		ctx context.Context,
	) (string, error)
	RegisterUser(
		ctx context.Context,
		guestUserID string,
		email string,
		password string,
	) (string, users_postgres_repository.User, error)
	LoginUser(
		ctx context.Context,
		email string,
		password string,
	) (string, users_postgres_repository.User, error)
	GetProfile(
		ctx context.Context,
		userID string,
	) (users_postgres_repository.Profile, error)
	UpdateProfile(
		ctx context.Context,
		userID string,
		expectedVersion int64,
		updates map[string]any,
	) (users_postgres_repository.Profile, error)
	GeneratePlan(
		ctx context.Context,
		userID string,
	) (users_postgres_repository.EPlanWithWeight, error)
	CompleteWorkout(
		ctx context.Context,
		userID string,
		req users_postgres_repository.WorkoutCompleteRequest,
	) (*users_postgres_repository.WorkoutCompleteServiceResult, error)
	ChangeUserPassword(
		ctx context.Context,
		userID string,
		current_password string,
		new_password string,
	) error
	GetListNotes(
		ctx context.Context,
		userID string,
	) ([]users_postgres_repository.Note, error)
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
}

func NewUsersHTTPHandler(
	usersService UsersService,
	jwt *core_auth.JWT,
) *UsersHTTPHandler {
	return &UsersHTTPHandler{
		usersService: usersService,
		jwt:          jwt,
	}
}

func (h *UsersHTTPHandler) Routes() []core_http_server.Route {
	protect := core_http_middleware.Protect(h.jwt)
	guestCreateRateLimit := core_http_middleware.RateLimit(10, time.Minute)
	registerRateLimit := core_http_middleware.RateLimit(20, time.Minute)
	loginRateLimit := core_http_middleware.RateLimit(30, time.Minute)

	return []core_http_server.Route{
		core_http_server.NewRoute(
			http.MethodPost, "/auth/guest", h.CreateGuestUser, guestCreateRateLimit,
		),
		core_http_server.NewRoute(
			http.MethodPost, "/auth/register", h.RegisterUser, registerRateLimit,
		),
		core_http_server.NewRoute(
			http.MethodPost, "/auth/login", h.LoginUser, loginRateLimit,
		),
		core_http_server.NewRoute(
			http.MethodPost, "/auth/change-password", h.ChangeUserPassword, protect,
		),
		core_http_server.NewRoute(
			http.MethodGet, "/profile", h.GetProfile, protect,
		),
		core_http_server.NewRoute(
			http.MethodPost, "/profile", h.UpdateProfile, protect,
		),
		core_http_server.NewRoute(
			http.MethodPost, "/plans/generate", h.GeneratePlan, protect,
		),
		core_http_server.NewRoute(
			http.MethodPost, "/workouts/complete", h.CompleteWorkout, protect,
		),
		core_http_server.NewRoute(
			http.MethodGet, "/notes", h.ListNotes, protect,
		),
		core_http_server.NewRoute(
			http.MethodPost, "/notes", h.CreateNote, protect,
		),
		core_http_server.NewRoute(
			http.MethodPatch, "/notes/{id}", h.UpdateNote, protect,
		),
		core_http_server.NewRoute(
			http.MethodDelete, "/notes/{id}", h.DeleteNote, protect,
		),
		core_http_server.NewRoute(
			http.MethodGet, "/achievements/get_achievements", h.GetAchievements, protect,
		),
	}
}
