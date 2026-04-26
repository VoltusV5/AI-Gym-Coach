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
	OK      bool   `json:"ok"`
	SavedID string `json:"saved_id"`
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

	if err := h.usersService.CompleteWorkout(ctx, userID, request); err != nil {
		responseHandler.ErrorResponse(err, "failed to complete workout")
		return
	}

	responseHandler.JSONResponse(CompleteWorkoutResponse{
		OK:      true,
		SavedID: fmt.Sprintf("real-%d", time.Now().Unix()),
	}, http.StatusOK)
}
