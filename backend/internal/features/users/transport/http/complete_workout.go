package users_transport_http

import (
	"fmt"
	"net/http"
	"time"

	core_logger "sport_app/internal/core/logger"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_request "sport_app/internal/core/transport/http/request"
	core_http_responce "sport_app/internal/core/transport/http/responce"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
)

type CompleteWorkoutResponse struct {
	OK                bool                                            `json:"ok"`
	SavedID           string                                          `json:"saved_id"`
	NewAchievements   []users_postgres_repository.AchievementUnlocked `json:"new_achievements"`
	SessionHighlights []users_postgres_repository.SessionHighlight    `json:"session_highlights"`
}

func (h *UsersHTTPHandler) CompleteWorkout(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_responce.NewHTTPResponce(log, rw)

	userID, err := core_http_middleware.UserIDFromContext(ctx)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get user id from context")
		return
	}

	var request users_postgres_repository.WorkoutCompleteRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode workout complete body")
		return
	}

	result, err := h.usersService.CompleteWorkout(ctx, userID, request)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to complete workout")
		return
	}
	if result == nil {
		result = &users_postgres_repository.WorkoutCompleteServiceResult{}
	}
	if result.NewAchievements == nil {
		result.NewAchievements = []users_postgres_repository.AchievementUnlocked{}
	}
	if result.SessionHighlights == nil {
		result.SessionHighlights = []users_postgres_repository.SessionHighlight{}
	}

	responseHandler.JSONResponse(CompleteWorkoutResponse{
		OK:                true,
		SavedID:           fmt.Sprintf("real-%d", time.Now().Unix()),
		NewAchievements:   result.NewAchievements,
		SessionHighlights: result.SessionHighlights,
	}, http.StatusOK)
}
