package users_transport_http

import (
	"net/http"

	core_logger "sport_app/internal/core/logger"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_responce "sport_app/internal/core/transport/http/responce"
)

func (h *UsersHTTPHandler) GeneratePlan(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_responce.NewHTTPResponce(log, rw)

	userID, err := core_http_middleware.UserIDFromContext(ctx)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get user id from context")
		return
	}

	plan, err := h.usersService.GeneratePlan(ctx, userID)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to generate plan")
		return
	}

	responseHandler.JSONResponse(plan, http.StatusOK)
}