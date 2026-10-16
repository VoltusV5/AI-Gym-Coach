package users_transport_http

import (
	"net/http"
	core_logger "sport_app/internal/core/logger"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_request "sport_app/internal/core/transport/http/request"
	core_http_responce "sport_app/internal/core/transport/http/responce"
)

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=72"`
}

func (h *UsersHTTPHandler) ChangeUserPassword(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_responce.NewHTTPResponce(log, rw)

	userID, err := core_http_middleware.UserIDFromContext(ctx)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get user id from context")
		return
	}

	var req ChangePasswordRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &req); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode change user request")
		return
	}
	defer r.Body.Close()

	if err := h.usersService.ChangeUserPassword(ctx, userID, req.CurrentPassword, req.NewPassword); err != nil {
		responseHandler.ErrorResponse(err, "failed change password")
		return
	}

	responseHandler.NoContentResponse()
}
