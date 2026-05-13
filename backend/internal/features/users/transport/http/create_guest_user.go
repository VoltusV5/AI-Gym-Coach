package users_transport_http

import (
	"fmt"
	"net/http"

	core_logger "sport_app/internal/core/logger"
	core_http_responce "sport_app/internal/core/transport/http/responce"
)

type CreateGuestUserResponse struct {
	Token string `json:"token"`
}

func (h *UsersHTTPHandler) CreateGuestUser(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_responce.NewHTTPResponce(log, rw)

	token, err := h.usersService.CreateGuestUser(ctx)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to create guest user")
		return
	}

	rw.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	responseHandler.JSONResponse(CreateGuestUserResponse{Token: token}, http.StatusOK)
}