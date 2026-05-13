package users_transport_http

import (
	"fmt"
	"net/http"

	core_logger "sport_app/internal/core/logger"
	core_http_request "sport_app/internal/core/transport/http/request"
	core_http_responce "sport_app/internal/core/transport/http/responce"
)

type LoginUserRequest struct {
	Email    string `json:"email"    validate:"required,email,max=200"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type LoginUserResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

func (h *UsersHTTPHandler) LoginUser(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_responce.NewHTTPResponce(log, rw)

	var req LoginUserRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &req); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode login user request")
		return
	}
	defer r.Body.Close()

	token, user, err := h.usersService.LoginUser(ctx, req.Email, req.Password)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to login user")
		return
	}

	rw.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	responseHandler.JSONResponse(LoginUserResponse{
		Token: token,
		User: UserResponse{
			ID:          user.ID,
			Email:       user.Email,
			IsAnonymous: user.IsAnonymous,
		},
	}, http.StatusOK)
}