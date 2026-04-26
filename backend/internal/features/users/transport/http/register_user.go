package users_transport_http

import (
	"fmt"
	"net/http"

	core_logger "sport_app/internal/core/logger"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_request "sport_app/internal/core/transport/http/request"
	core_http_responce "sport_app/internal/core/transport/http/responce"
)

type RegisterUserRequest struct {
	Email    string `json:"email"    validate:"required,email,max=200"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type RegisterUserResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID          int     `json:"id"`
	Email       *string `json:"email"`
	IsAnonymous bool    `json:"is_anonymous"`
}

func (h *UsersHTTPHandler) RegisterUser(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_responce.NewHTTPResponce(log, rw)

	guestUserID, err := core_http_middleware.UserIDFromContext(ctx)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get user id from context")
		return
	}

	var req RegisterUserRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &req); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode register user request")
		return
	}
	defer r.Body.Close()

	token, user, err := h.usersService.RegisterUser(ctx, guestUserID, req.Email, req.Password)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to register user")
		return
	}

	rw.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	responseHandler.JSONResponse(RegisterUserResponse{
		Token: token,
		User: UserResponse{
			ID:          user.ID,
			Email:       user.Email,
			IsAnonymous: user.IsAnonymous,
		},
	}, http.StatusOK)
}
